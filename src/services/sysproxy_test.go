package services

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRemoveRTXBlock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		content     string
		want        string
		wantRemoved bool
	}{
		{
			name:        "complete block",
			content:     "before\n" + rtxProxyMarkerStart + "\nold\nexport http_proxy=x\n" + rtxProxyMarkerEnd + "\nafter\n",
			want:        "before\nafter",
			wantRemoved: true,
		},
		{
			name:        "unterminated block is preserved",
			content:     "before\n" + rtxProxyMarkerStart + "\nexport http_proxy=x\n",
			want:        "before\n" + rtxProxyMarkerStart + "\nexport http_proxy=x\n",
			wantRemoved: false,
		},
		{
			name:        "no block",
			content:     "export PATH=/bin\n",
			want:        "export PATH=/bin\n",
			wantRemoved: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, removed := removeRTXBlock(tc.content)
			if removed != tc.wantRemoved {
				t.Errorf("removeRTXBlock() removed = %v, want %v", removed, tc.wantRemoved)
			}
			if got != tc.want {
				t.Errorf("removeRTXBlock() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestProxyEnvLinesQuoteValues(t *testing.T) {
	t.Parallel()

	lines, err := proxyEnvLines("proxy.example.com:8080", "bash")
	if err != nil {
		t.Fatalf("proxyEnvLines() failed: %v", err)
	}

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "export http_proxy='http://proxy.example.com:8080'") {
		t.Errorf("proxy value is not safely quoted: %q", joined)
	}
	if !strings.Contains(joined, "export no_proxy='localhost,127.0.0.1,::1,.local,.internal'") {
		t.Errorf("no_proxy value is not safely quoted: %q", joined)
	}

	if _, err := proxyEnvLines("http://proxy.example.com:8080\nrm -rf /", "bash"); err == nil {
		t.Fatal("proxyEnvLines() accepted a proxy URL with a newline")
	}
}

func TestWriteProxyToFileRejectsSymlink(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outside := filepath.Join(dir, "outside")
	if err := os.WriteFile(outside, []byte("original\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, ".bashrc")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	if _, err := writeProxyToFile(link, "http://proxy.example.com:8080", "bash"); err == nil {
		t.Fatal("writeProxyToFile() wrote through a symlink")
	}
	got, err := os.ReadFile(outside)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "original\n" {
		t.Errorf("outside file changed to %q", got)
	}
}

func TestSetAndUnsetSystemProxy(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("shell configuration test is Linux-specific")
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("SHELL", "/bin/bash")

	result, err := SetSystemProxy(context.Background(), "http://user:secret@proxy.example.com:8080")
	if err != nil {
		t.Fatalf("SetSystemProxy() failed: %v", err)
	}
	if strings.Contains(result.ProxyURL, "secret") {
		t.Errorf("SetSystemProxy() leaked credentials in result: %q", result.ProxyURL)
	}
	if len(result.FilesUpdated) == 0 {
		t.Fatal("SetSystemProxy() did not update any file")
	}

	bashrc := filepath.Join(home, ".bashrc")
	info, err := os.Stat(bashrc)
	if err != nil {
		t.Fatalf("stat .bashrc: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf(".bashrc mode = %o, want 600", perm)
	}

	content, err := os.ReadFile(bashrc)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "export http_proxy='http://user:secret@proxy.example.com:8080'") {
		t.Errorf("proxy was not written safely: %q", content)
	}

	if _, err := UnsetSystemProxy(context.Background()); err != nil {
		t.Fatalf("UnsetSystemProxy() failed: %v", err)
	}
	content, err = os.ReadFile(bashrc)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), rtxProxyMarkerStart) {
		t.Errorf("RTX block was not removed: %q", content)
	}
}

func TestSetSystemProxyRejectsMaliciousURL(t *testing.T) {
	t.Parallel()

	if _, err := SetSystemProxy(context.Background(), "http://proxy.example.com:8080'; touch /tmp/rtx-pwned; #"); err == nil {
		t.Fatal("SetSystemProxy() accepted an injection payload")
	}
}
