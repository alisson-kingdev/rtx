package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"rtx/src/models"
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

	// Adicionar /etc/environment para Linux (sistema-wide)
	if runtime.GOOS == "linux" {
		files = append(files, "/etc/environment")
	}

	return files
}

func proxyEnvLines(proxyURL string, shell string) []string {
	// Normalizar: garantir http:// prefixo se não tiver esquema
	normalized := proxyURL
	if !strings.HasPrefix(normalized, "http://") && !strings.HasPrefix(normalized, "https://") {
		normalized = "http://" + normalized
	}

	noProxy := "localhost,127.0.0.1,::1,.local,.internal"

	if shell == "fish" {
		return []string{
			rtxProxyMarkerStart,
			fmt.Sprintf("set -gx http_proxy %s", normalized),
			fmt.Sprintf("set -gx https_proxy %s", normalized),
			fmt.Sprintf("set -gx HTTP_PROXY %s", normalized),
			fmt.Sprintf("set -gx HTTPS_PROXY %s", normalized),
			fmt.Sprintf("set -gx no_proxy %s", noProxy),
			fmt.Sprintf("set -gx NO_PROXY %s", noProxy),
			rtxProxyMarkerEnd,
		}
	}

	return []string{
		rtxProxyMarkerStart,
		fmt.Sprintf("export http_proxy=%s", normalized),
		fmt.Sprintf("export https_proxy=%s", normalized),
		fmt.Sprintf("export HTTP_PROXY=%s", normalized),
		fmt.Sprintf("export HTTPS_PROXY=%s", normalized),
		fmt.Sprintf("export no_proxy=%s", noProxy),
		fmt.Sprintf("export NO_PROXY=%s", noProxy),
		rtxProxyMarkerEnd,
	}
}

func removeRTXBlock(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	insideBlock := false

	for _, line := range lines {
		if strings.Contains(line, rtxProxyMarkerStart) {
			insideBlock = true
			continue
		}
		if strings.Contains(line, rtxProxyMarkerEnd) {
			insideBlock = false
			continue
		}
		if !insideBlock {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

func writeProxyToFile(filePath string, proxyURL string, shell string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Criar arquivo novo apenas com o bloco RTX
			lines := proxyEnvLines(proxyURL, shell)
			return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")+"\n"), 0644)
		}
		return err
	}

	// Remover bloco anterior se existir
	cleaned := removeRTXBlock(string(content))

	// Limpar linhas vazias duplicadas no final
	cleaned = strings.TrimRight(cleaned, "\n")

	// Adicionar novo bloco
	newBlock := strings.Join(proxyEnvLines(proxyURL, shell), "\n")

	final := cleaned + "\n\n" + newBlock + "\n"

	return os.WriteFile(filePath, []byte(final), 0644)
}

func unsetProxyFromFile(filePath string, shell string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	cleaned := removeRTXBlock(string(content))

	if string(content) == cleaned {
		return false, nil // nada foi removido
	}

	cleaned = strings.TrimRight(cleaned, "\n")
	return true, os.WriteFile(filePath, []byte(cleaned+"\n"), 0644)
}

func setSystemProxyWindows(proxyURL string) error {
	normalized := normalizeProxyURL(proxyURL)
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
		cmd := exec.Command("setx", v.name, v.value)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("setx %s: %w", v.name, err)
		}
	}

	cmd := exec.Command("netsh", "winhttp", "set", "proxy",
		fmt.Sprintf("proxy-server=\"%s\"", normalized),
		"bypass-list=\"localhost;127.0.0.1;*.local;*.internal\"")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("netsh winhttp: %w", err)
	}

	return nil
}

func unsetSystemProxyWindows() error {
	vars := []string{"http_proxy", "https_proxy", "HTTP_PROXY", "HTTPS_PROXY", "no_proxy", "NO_PROXY"}

	for _, v := range vars {
		cmd := exec.Command("setx", v, "")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("setx %s: %w", v, err)
		}
	}

	cmd := exec.Command("netsh", "winhttp", "reset", "proxy")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("netsh reset: %w", err)
	}

	return nil
}

func normalizeProxyURL(proxyURL string) string {
	if !strings.HasPrefix(proxyURL, "http://") && !strings.HasPrefix(proxyURL, "https://") {
		return "http://" + proxyURL
	}
	return proxyURL
}

func setSystemProxyMacOS(proxyURL string, networkService string) error {
	normalized := normalizeProxyURL(proxyURL)
	hostPort := strings.TrimPrefix(normalized, "http://")
	hostPort = strings.TrimPrefix(hostPort, "https://")
	parts := strings.Split(hostPort, ":")
	host := parts[0]
	port := "8080"
	if len(parts) > 1 {
		port = parts[1]
	}

	services := []string{networkService}
	if networkService == "" {
		out, err := exec.Command("networksetup", "-listallnetworkservices").Output()
		if err != nil {
			return fmt.Errorf("networksetup list: %w", err)
		}
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if line != "" && line != "*" && !strings.HasPrefix(line, "An asterisk") {
				services = append(services, line)
			}
		}
	}

	for _, svc := range services {
		if err := exec.Command("networksetup", "-setwebproxy", svc, host, port).Run(); err != nil {
			return fmt.Errorf("networksetup setwebproxy %s: %w", svc, err)
		}
		if err := exec.Command("networksetup", "-setsecurewebproxy", svc, host, port).Run(); err != nil {
			return fmt.Errorf("networksetup setsecurewebproxy %s: %w", svc, err)
		}
		if err := exec.Command("networksetup", "-setwebproxystate", svc, "on").Run(); err != nil {
			return fmt.Errorf("networksetup webproxystate %s: %w", svc, err)
		}
		if err := exec.Command("networksetup", "-setsecurewebproxystate", svc, "on").Run(); err != nil {
			return fmt.Errorf("networksetup secureproxystate %s: %w", svc, err)
		}
	}

	return nil
}

func unsetSystemProxyMacOS(networkService string) error {
	services := []string{networkService}
	if networkService == "" {
		out, err := exec.Command("networksetup", "-listallnetworkservices").Output()
		if err != nil {
			return fmt.Errorf("networksetup list: %w", err)
		}
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if line != "" && line != "*" && !strings.HasPrefix(line, "An asterisk") {
				services = append(services, line)
			}
		}
	}

	for _, svc := range services {
		if err := exec.Command("networksetup", "-setwebproxystate", svc, "off").Run(); err != nil {
			return fmt.Errorf("networksetup webproxystate off %s: %w", svc, err)
		}
		if err := exec.Command("networksetup", "-setsecurewebproxystate", svc, "off").Run(); err != nil {
			return fmt.Errorf("networksetup secureproxystate off %s: %w", svc, err)
		}
	}

	return nil
}

func SetSystemProxy(proxyURL string) models.SystemProxyResult {
	result := models.SystemProxyResult{
		Action:       "set",
		ProxyURL:     proxyURL,
		Success:      false,
		FilesUpdated: []string{},
	}

	shell := detectShell()
	result.Shell = shell

	switch runtime.GOOS {
	case "windows":
		if err := setSystemProxyWindows(proxyURL); err != nil {
			result.Error = err
			return result
		}
		result.Success = true
		result.FilesUpdated = []string{"Windows Environment Variables (setx)", "WinHTTP"}
		return result

	case "darwin":
		// Tentar networksetup (precisa de sudo em alguns casos)
		if err := setSystemProxyMacOS(proxyURL, ""); err != nil {
			result.Error = err
			return result
		}
		// Também escrever nos arquivos de shell para CLI tools
		files := shellConfigFiles(shell)
		updated := []string{}
		for _, f := range files {
			if _, err := unsetProxyFromFile(f, shell); err == nil {
				// removido ou nao existia, agora escreve
			}
			if err := writeProxyToFile(f, proxyURL, shell); err == nil {
				updated = append(updated, f)
			}
		}
		result.Success = true
		result.FilesUpdated = updated
		return result

	default: // linux, android (termux), freebsd, etc.
		files := shellConfigFiles(shell)
		updated := []string{}
		for _, f := range files {
			if err := writeProxyToFile(f, proxyURL, shell); err == nil {
				updated = append(updated, f)
			}
		}
		if len(updated) == 0 {
			result.Error = fmt.Errorf("nenhum arquivo de configuracao de shell encontrado para escrever")
			return result
		}
		result.Success = true
		result.FilesUpdated = updated
		return result
	}
}

func UnsetSystemProxy() models.SystemProxyResult {
	result := models.SystemProxyResult{
		Action:       "unset",
		Success:      false,
		FilesUpdated: []string{},
	}

	shell := detectShell()
	result.Shell = shell

	switch runtime.GOOS {
	case "windows":
		if err := unsetSystemProxyWindows(); err != nil {
			result.Error = err
			return result
		}
		result.Success = true
		result.FilesUpdated = []string{"Windows Environment Variables (setx)", "WinHTTP"}
		return result

	case "darwin":
		if err := unsetSystemProxyMacOS(""); err != nil {
			result.Error = err
			return result
		}
		files := shellConfigFiles(shell)
		updated := []string{}
		for _, f := range files {
			if removed, _ := unsetProxyFromFile(f, shell); removed {
				updated = append(updated, f)
			}
		}
		result.Success = true
		result.FilesUpdated = updated
		return result

	default:
		files := shellConfigFiles(shell)
		updated := []string{}
		for _, f := range files {
			if removed, _ := unsetProxyFromFile(f, shell); removed {
				updated = append(updated, f)
			}
		}
		result.Success = true
		result.FilesUpdated = updated
		return result
	}
}
