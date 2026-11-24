package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"rtx/src/models"
	"rtx/src/utils"
)

const (
	rtxProxyMarkerStart = "# >>> RTX PROXY >>>"
	rtxProxyMarkerEnd   = "# <<< RTX PROXY <<<"
)

func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	return filepath.Base(shell)
}

func shellConfigFiles(shell string) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	if resolved, err := filepath.EvalSymlinks(home); err == nil {
		home = resolved
	}

	var files []string
	switch shell {
	case "zsh":
		files = []string{
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".zshenv"),
			filepath.Join(home, ".zprofile"),
		}
	case "bash":
		files = []string{
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".bash_profile"),
			filepath.Join(home, ".profile"),
		}
	case "fish":
		files = []string{
			filepath.Join(home, ".config", "fish", "config.fish"),
		}
	case "sh", "ash", "dash":
		files = []string{
			filepath.Join(home, ".profile"),
			filepath.Join(home, ".shrc"),
		}
	default:
		files = []string{
			filepath.Join(home, ".profile"),
		}
	}

	return files
}

func proxyEnvLines(proxyURL string, shell string) ([]string, error) {
	normalized, err := utils.NormalizeProxyURL(proxyURL)
	if err != nil {
		return nil, err
	}
	quoted := utils.ShellQuote(normalized)
	noProxy := utils.ShellQuote("localhost,127.0.0.1,::1,.local,.internal")

	if shell == "fish" {
		return []string{
			rtxProxyMarkerStart,
			fmt.Sprintf("set -gx http_proxy %s", quoted),
			fmt.Sprintf("set -gx https_proxy %s", quoted),
			fmt.Sprintf("set -gx HTTP_PROXY %s", quoted),
			fmt.Sprintf("set -gx HTTPS_PROXY %s", quoted),
			fmt.Sprintf("set -gx no_proxy %s", noProxy),
			fmt.Sprintf("set -gx NO_PROXY %s", noProxy),
			rtxProxyMarkerEnd,
		}, nil
	}

	return []string{
		rtxProxyMarkerStart,
		fmt.Sprintf("export http_proxy=%s", quoted),
		fmt.Sprintf("export https_proxy=%s", quoted),
		fmt.Sprintf("export HTTP_PROXY=%s", quoted),
		fmt.Sprintf("export HTTPS_PROXY=%s", quoted),
		fmt.Sprintf("export no_proxy=%s", noProxy),
		fmt.Sprintf("export NO_PROXY=%s", noProxy),
		rtxProxyMarkerEnd,
	}, nil
}

// removeRTXBlock removes complete RTX-managed blocks without deleting
// surrounding user configuration. An unterminated marker is left untouched.
func removeRTXBlock(content string) (string, bool) {
	lines := strings.Split(content, "\n")
	kept := make([]string, 0, len(lines))
	removed := false

	for i := 0; i < len(lines); {
		if strings.TrimSpace(lines[i]) != rtxProxyMarkerStart {
			kept = append(kept, lines[i])
			i++
			continue
		}

		end := -1
		for j := i + 1; j < len(lines); j++ {
			trimmed := strings.TrimSpace(lines[j])
			if trimmed == rtxProxyMarkerEnd {
				end = j
				break
			}
			if trimmed == rtxProxyMarkerStart {
				break
			}
		}
		if end < 0 {
			kept = append(kept, lines[i])
			i++
			continue
		}

		i = end + 1
		removed = true
	}

	result := strings.Join(kept, "\n")
	if removed {
		result = strings.TrimRight(result, "\n")
	}
	return result, removed
}

func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)

	existing, err := os.Lstat(path)
	switch {
	case err == nil:
		if existing.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to write through symlink %s", path)
		}
		if !existing.Mode().IsRegular() {
			return fmt.Errorf("refusing to write to non-regular file %s", path)
		}
		mode = existing.Mode().Perm()
	case errors.Is(err, os.ErrNotExist):
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	default:
		return fmt.Errorf("inspect %s: %w", path, err)
	}

	tmp, err := os.CreateTemp(dir, ".rtx-tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	keepTemp := true
	defer func() {
		if keepTemp {
			_ = os.Remove(tmpName)
		}
	}()

	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("set permissions on temporary file: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temporary file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("sync temporary file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temporary file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}

	keepTemp = false
	return nil
}

func writeProxyToFile(filePath string, proxyURL string, shell string) (bool, error) {
	lines, err := proxyEnvLines(proxyURL, shell)
	if err != nil {
		return false, err
	}

	content, err := os.ReadFile(filePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("read %s: %w", filePath, err)
	}

	cleaned, _ := removeRTXBlock(string(content))
	cleaned = strings.TrimRight(cleaned, "\n")

	block := strings.Join(lines, "\n")
	final := block + "\n"
	if cleaned != "" {
		final = cleaned + "\n\n" + block + "\n"
	}
	if string(content) == final {
		return false, nil
	}

	if err := writeFileAtomic(filePath, []byte(final), 0o600); err != nil {
		return false, err
	}
	return true, nil
}

func unsetProxyFromFile(filePath string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read %s: %w", filePath, err)
	}

	cleaned, removed := removeRTXBlock(string(content))
	if !removed {
		return false, nil
	}

	cleaned = strings.TrimRight(cleaned, "\n")
	final := cleaned
	if final != "" {
		final += "\n"
	}
	if string(content) == final {
		return false, nil
	}
	if err := writeFileAtomic(filePath, []byte(final), 0o600); err != nil {
		return false, err
	}
	return true, nil
}

func applyProxyToShellFiles(proxyURL, shell string) ([]string, error) {
	var (
		updated []string
		errs    []error
	)
	for _, file := range shellConfigFiles(shell) {
		changed, err := writeProxyToFile(file, proxyURL, shell)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", file, err))
			continue
		}
		if changed {
			updated = append(updated, file)
		}
	}
	return updated, errors.Join(errs...)
}

func removeProxyFromShellFiles() ([]string, error) {
	var (
		updated []string
		errs    []error
	)
	for _, file := range shellConfigFiles(detectShell()) {
		removed, err := unsetProxyFromFile(file)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", file, err))
			continue
		}
		if removed {
			updated = append(updated, file)
		}
	}
	return updated, errors.Join(errs...)
}

func setSystemProxyWindows(ctx context.Context, proxyURL string) error {
	normalized, err := utils.NormalizeProxyURL(proxyURL)
	if err != nil {
		return err
	}
	noProxy := "localhost,127.0.0.1"

	vars := []struct {
		name  string
		value string
	}{
		{"http_proxy", normalized},
		{"https_proxy", normalized},
		{"HTTP_PROXY", normalized},
		{"HTTPS_PROXY", normalized},
		{"no_proxy", noProxy},
		{"NO_PROXY", noProxy},
	}

	for _, v := range vars {
		if err := exec.CommandContext(ctx, "setx", v.name, v.value).Run(); err != nil {
			return fmt.Errorf("setx %s: %w", v.name, err)
		}
	}

	cmd := exec.CommandContext(ctx, "netsh", "winhttp", "set", "proxy",
		fmt.Sprintf("proxy-server=%q", normalized),
		"bypass-list=localhost;127.0.0.1;*.local;*.internal")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("netsh winhttp: %w", err)
	}
	return nil
}

func unsetSystemProxyWindows(ctx context.Context) error {
	vars := []string{"http_proxy", "https_proxy", "HTTP_PROXY", "HTTPS_PROXY", "no_proxy", "NO_PROXY"}
	for _, name := range vars {
		if err := exec.CommandContext(ctx, "setx", name, "").Run(); err != nil {
			return fmt.Errorf("setx %s: %w", name, err)
		}
	}
	if err := exec.CommandContext(ctx, "netsh", "winhttp", "reset", "proxy").Run(); err != nil {
		return fmt.Errorf("netsh reset: %w", err)
	}
	return nil
}

func networkServices(ctx context.Context) ([]string, error) {
	out, err := exec.CommandContext(ctx, "networksetup", "-listallnetworkservices").Output()
	if err != nil {
		return nil, fmt.Errorf("networksetup list: %w", err)
	}

	var services []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "*" || strings.HasPrefix(line, "An asterisk") {
			continue
		}
		services = append(services, line)
	}
	return services, nil
}

func validateNetworkService(service string) error {
	if service == "" {
		return nil
	}
	if strings.HasPrefix(service, "-") || strings.ContainsAny(service, "\r\n\t") {
		return fmt.Errorf("invalid network service name %q", service)
	}
	return nil
}

func setSystemProxyMacOS(ctx context.Context, proxyURL string, networkService string) error {
	if err := validateNetworkService(networkService); err != nil {
		return err
	}

	parsed, err := utils.ParseProxyURL(proxyURL)
	if err != nil {
		return err
	}
	if parsed.User != nil {
		return errors.New("authenticated proxies are not supported by networksetup")
	}

	host := parsed.Hostname()
	port := parsed.Port()
	if host == "" || port == "" {
		return errors.New("proxy URL must include a host and port")
	}

	services := []string{networkService}
	if networkService == "" {
		services, err = networkServices(ctx)
		if err != nil {
			return err
		}
	}

	for _, service := range services {
		commands := [][]string{
			{"-setwebproxy", service, host, port},
			{"-setsecurewebproxy", service, host, port},
			{"-setwebproxystate", service, "on"},
			{"-setsecurewebproxystate", service, "on"},
		}
		for _, args := range commands {
			if err := exec.CommandContext(ctx, "networksetup", args...).Run(); err != nil {
				return fmt.Errorf("networksetup %s %s: %w", args[0], service, err)
			}
		}
	}
	return nil
}

func unsetSystemProxyMacOS(ctx context.Context, networkService string) error {
	if err := validateNetworkService(networkService); err != nil {
		return err
	}

	services := []string{networkService}
	if networkService == "" {
		var err error
		services, err = networkServices(ctx)
		if err != nil {
			return err
		}
	}

	for _, service := range services {
		commands := [][]string{
			{"-setwebproxystate", service, "off"},
			{"-setsecurewebproxystate", service, "off"},
		}
		for _, args := range commands {
			if err := exec.CommandContext(ctx, "networksetup", args...).Run(); err != nil {
				return fmt.Errorf("networksetup %s %s: %w", args[0], service, err)
			}
		}
	}
	return nil
}

// SetSystemProxy applies a validated proxy to the current user's shell
// configuration and, where supported, to OS-level network settings.
func SetSystemProxy(ctx context.Context, proxyURL string) (models.SystemProxyResult, error) {
	if err := ctx.Err(); err != nil {
		return models.SystemProxyResult{}, err
	}
	normalized, err := utils.NormalizeProxyURL(proxyURL)
	if err != nil {
		return models.SystemProxyResult{}, fmt.Errorf("validate proxy URL: %w", err)
	}

	result := models.SystemProxyResult{
		Action:       "set",
		ProxyURL:     utils.RedactURL(normalized),
		Shell:        detectShell(),
		FilesUpdated: []string{},
	}

	switch runtime.GOOS {
	case "windows":
		if err := setSystemProxyWindows(ctx, normalized); err != nil {
			return result, err
		}
		result.FilesUpdated = []string{"Windows Environment Variables (setx)", "WinHTTP"}
		return result, nil

	case "darwin":
		if err := setSystemProxyMacOS(ctx, normalized, ""); err != nil {
			return result, err
		}
		updated, err := applyProxyToShellFiles(normalized, result.Shell)
		result.FilesUpdated = updated
		return result, err

	default:
		updated, err := applyProxyToShellFiles(normalized, result.Shell)
		result.FilesUpdated = updated
		if err != nil {
			return result, err
		}
		if len(updated) == 0 {
			return result, errors.New("no shell configuration file could be updated")
		}
		return result, nil
	}
}

// UnsetSystemProxy removes RTX-managed proxy settings from the user's shell
// configuration and resets OS-level network settings where supported.
func UnsetSystemProxy(ctx context.Context) (models.SystemProxyResult, error) {
	if err := ctx.Err(); err != nil {
		return models.SystemProxyResult{}, err
	}

	result := models.SystemProxyResult{
		Action:       "unset",
		Shell:        detectShell(),
		FilesUpdated: []string{},
	}

	switch runtime.GOOS {
	case "windows":
		if err := unsetSystemProxyWindows(ctx); err != nil {
			return result, err
		}
		result.FilesUpdated = []string{"Windows Environment Variables (setx)", "WinHTTP"}
		return result, nil

	case "darwin":
		if err := unsetSystemProxyMacOS(ctx, ""); err != nil {
			return result, err
		}
		updated, err := removeProxyFromShellFiles()
		result.FilesUpdated = updated
		return result, err

	default:
		updated, err := removeProxyFromShellFiles()
		result.FilesUpdated = updated
		return result, err
	}
}
