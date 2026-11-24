package utils

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestHTTPClientFetch(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("1.2.3.4\n"))
	}))
	defer server.Close()

	client := NewHTTPClient(time.Second)
	body, err := client.Fetch(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("Fetch() failed: %v", err)
	}
	if string(body) != "1.2.3.4\n" {
		t.Errorf("Fetch() = %q, want %q", body, "1.2.3.4\n")
	}
}

func TestHTTPClientFetchErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{
			name: "non-2xx status",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, "boom", http.StatusInternalServerError)
			},
		},
		{
			name: "oversized body",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(strings.Repeat("x", 4096)))
			},
		},
		{
			name: "cross host redirect",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "http://127.0.0.1:1/elsewhere", http.StatusFound)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()

			client := NewHTTPClient(time.Second)
			client.maxBytes = 32
			_, err := client.Fetch(context.Background(), server.URL)
			if err == nil {
				t.Fatal("Fetch() succeeded, want error")
			}
		})
	}
}

func TestHTTPClientFetchViaProxy(t *testing.T) {
	t.Parallel()

	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("5.6.7.8"))
	}))
	defer proxy.Close()

	client := NewHTTPClient(time.Second)
	body, err := client.FetchViaProxy(context.Background(), "http://api.ipify.org", proxy.URL)
	if err != nil {
		t.Fatalf("FetchViaProxy() failed: %v", err)
	}
	if string(body) != "5.6.7.8" {
		t.Errorf("FetchViaProxy() = %q, want 5.6.7.8", body)
	}

	if _, err := client.FetchViaProxy(context.Background(), "http://api.ipify.org", "bad host"); err == nil {
		t.Fatal("FetchViaProxy() succeeded with invalid proxy, want error")
	}
}

func TestSameHostRedirects(t *testing.T) {
	t.Parallel()

	initial, err := url.Parse("https://api.ipify.org")
	if err != nil {
		t.Fatal(err)
	}
	policy := sameHostRedirects(initial)

	req, err := http.NewRequest(http.MethodGet, "https://api.ipify.org/ip", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := policy(req, nil); err != nil {
		t.Fatalf("same-host redirect rejected: %v", err)
	}

	foreign, err := http.NewRequest(http.MethodGet, "https://example.com/ip", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := policy(foreign, make([]*http.Request, 1)); !errors.Is(err, ErrCrossHostRedirect) {
		t.Fatalf("error = %v, want ErrCrossHostRedirect", err)
	}
}
