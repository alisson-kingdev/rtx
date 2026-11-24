package services

import (
	"context"
	"errors"
	"testing"
)

func TestIPLookupPublicIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		response  string
		want      string
		wantError bool
	}{
		{name: "valid IPv4", response: "1.2.3.4\n", want: "1.2.3.4"},
		{name: "valid IPv6", response: " 2001:0db8::1 \n", want: "2001:db8::1"},
		{name: "HTML error page", response: "<html>blocked</html>", wantError: true},
		{name: "empty response", response: "   \n", wantError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fetcher := &stubFetcher{fetch: func(_ context.Context, _ string) ([]byte, error) {
				return []byte(tc.response), nil
			}}
			lookup := NewIPLookup(fetcher, []string{"https://ip.test"})

			got, err := lookup.PublicIP(context.Background())
			if tc.wantError {
				if err == nil {
					t.Fatalf("PublicIP() = %q, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("PublicIP() failed: %v", err)
			}
			if got != tc.want {
				t.Errorf("PublicIP() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestIPLookupPublicIPAllEndpointsFail(t *testing.T) {
	t.Parallel()

	fetcher := &stubFetcher{fetch: func(_ context.Context, _ string) ([]byte, error) {
		return nil, errors.New("network down")
	}}
	lookup := NewIPLookup(fetcher, []string{"https://ip.test", "https://ip2.test"})

	if _, err := lookup.PublicIP(context.Background()); err == nil {
		t.Fatal("PublicIP() succeeded, want error")
	}
}

func TestIPLookupPublicIPViaProxy(t *testing.T) {
	t.Parallel()

	fetcher := &stubFetcher{via: func(_ context.Context, _, proxyURL string) ([]byte, error) {
		if proxyURL != "http://1.2.3.4:8080" {
			return nil, errors.New("unexpected proxy")
		}
		return []byte("5.6.7.8"), nil
	}}
	lookup := NewIPLookup(fetcher, []string{"https://ip.test"})

	got, err := lookup.PublicIPViaProxy(context.Background(), "1.2.3.4:8080")
	if err != nil {
		t.Fatalf("PublicIPViaProxy() failed: %v", err)
	}
	if got != "5.6.7.8" {
		t.Errorf("PublicIPViaProxy() = %q, want 5.6.7.8", got)
	}

	if _, err := lookup.PublicIPViaProxy(context.Background(), "bad host"); err == nil {
		t.Fatal("PublicIPViaProxy() accepted invalid proxy, want error")
	}
}

func TestIPLookupHonorsCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	lookup := NewIPLookup(&stubFetcher{}, []string{"https://ip.test"})
	if _, err := lookup.PublicIP(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}
