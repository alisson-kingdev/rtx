package utils

import (
	"errors"
	"testing"
)

func TestParseProxyURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "host and port", input: "proxy.example.com:3128", want: "http://proxy.example.com:3128"},
		{name: "explicit http", input: "http://user:pass@proxy.example.com:8080", want: "http://user:pass@proxy.example.com:8080"},
		{name: "ipv6", input: "http://[::1]:8080", want: "http://[::1]:8080"},
		{name: "https", input: "https://proxy.example.com:443", want: "https://proxy.example.com:443"},
		{name: "missing port", input: "http://proxy.example.com", wantErr: true},
		{name: "invalid port", input: "http://proxy.example.com:99999", wantErr: true},
		{name: "unsupported scheme", input: "socks5://proxy.example.com:1080", wantErr: true},
		{name: "path injection", input: "http://proxy.example.com:8080/;curl evil", wantErr: true},
		{name: "newline", input: "http://proxy.example.com:8080\nrm -rf /", wantErr: true},
		{name: "empty", input: "   ", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeProxyURL(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("NormalizeProxyURL(%q) succeeded, want error", tc.input)
				}
				if !errors.Is(err, ErrInvalidProxyURL) {
					t.Fatalf("error = %v, want ErrInvalidProxyURL", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeProxyURL(%q) failed: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("NormalizeProxyURL(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestRedactURL(t *testing.T) {
	t.Parallel()

	got := RedactURL("http://user:secret@proxy.example.com:8080")
	if got == "http://user:secret@proxy.example.com:8080" {
		t.Errorf("RedactURL leaked credentials: %s", got)
	}
	if got != "http://user:xxxxx@proxy.example.com:8080" {
		t.Errorf("RedactURL() = %q, want masked password", got)
	}
	if got := RedactURL("not a url"); got != "<invalid URL>" {
		t.Errorf("RedactURL() = %q, want <invalid URL>", got)
	}
}
