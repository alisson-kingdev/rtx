package utils

import "testing"

func TestShellQuote(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "plain", input: "http://proxy.example.com:8080", want: `'http://proxy.example.com:8080'`},
		{name: "single quote", input: "http://proxy:8080/it's", want: `'http://proxy:8080/it'\''s'`},
		{name: "dollar sign", input: "http://proxy:8080/$HOME", want: `'http://proxy:8080/$HOME'`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ShellQuote(tc.input); got != tc.want {
				t.Errorf("ShellQuote(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
