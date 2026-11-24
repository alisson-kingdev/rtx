package commands

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newTestRoot() *cobra.Command {
	root := &cobra.Command{Use: "rtx"}
	root.PersistentFlags().BoolP("verbose", "v", false, "Ativar modo verboso")
	Register(root)
	return root
}

func TestRegisterCommands(t *testing.T) {
	t.Parallel()

	root := newTestRoot()
	names := make(map[string]bool)
	for _, cmd := range root.Commands() {
		names[cmd.Name()] = true
	}
	for _, name := range []string{"rotate", "set-proxy", "unset-proxy"} {
		if !names[name] {
			t.Errorf("command %q was not registered", name)
		}
	}
	if root.PersistentFlags().Lookup("verbose") == nil {
		t.Error("root is missing the persistent verbose flag")
	}
	for _, cmd := range root.Commands() {
		if cmd.Flags().Lookup("verbose") != nil {
			t.Errorf("command %q redefines the verbose flag", cmd.Name())
		}
	}
}

func TestSetProxyRejectsInvalidURL(t *testing.T) {
	t.Parallel()

	root := newTestRoot()
	root.SetArgs([]string{"set-proxy", "http://proxy.example.com:8080'; touch /tmp/rtx-pwned; #"})

	var out, errOut bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errOut)

	if err := root.ExecuteContext(context.Background()); err == nil {
		t.Fatal("set-proxy accepted an injection payload")
	}
	if strings.Contains(out.String(), "pwned") {
		t.Errorf("unexpected output: %q", out.String())
	}
}

func TestUnsetProxyRemovesManagedBlock(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("shell configuration test is Linux-specific")
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("SHELL", "/bin/bash")

	bashrc := filepath.Join(home, ".bashrc")
	content := "export PATH=/bin\n# >>> RTX PROXY >>>\nexport http_proxy=http://proxy.example.com:8080\n# <<< RTX PROXY <<<\n"
	if err := os.WriteFile(bashrc, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	root := newTestRoot()
	root.SetArgs([]string{"unset-proxy"})
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)

	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("unset-proxy failed: %v", err)
	}
	got, err := os.ReadFile(bashrc)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "# >>> RTX PROXY >>>") {
		t.Errorf("RTX block remains: %q", got)
	}
	if !strings.Contains(string(got), "export PATH=/bin") {
		t.Errorf("user configuration was removed: %q", got)
	}
}
