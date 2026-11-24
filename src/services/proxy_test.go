package services

import (
	"context"
	"errors"
	"testing"

	"rtx/src/models"
)

func TestProxyCollectorCollect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		responses map[string]string
		wantURLs  []string
		wantErr   bool
	}{
		{
			name: "text and JSON sources",
			responses: map[string]string{
				"https://gimmeproxy.test":   "http://1.1.1.1:8080\n2.2.2.2:3128\n",
				"https://hproxy.test":       `{"proxies":[{"proxy":"http://3.3.3.3:8080"}]}`,
				"https://proxyscraper.test": "4.4.4.4:3128\nhttp://1.1.1.1:8080\n",
			},
			wantURLs: []string{"http://1.1.1.1:8080", "http://2.2.2.2:3128", "http://3.3.3.3:8080", "http://4.4.4.4:3128"},
		},
		{
			name: "invalid candidates are ignored",
			responses: map[string]string{
				"https://gimmeproxy.test": "http://proxy;rm -rf /:8080\nhttp://5.5.5.5:8080\n",
			},
			wantURLs: []string{"http://5.5.5.5:8080"},
		},
		{
			name: "all sources fail",
			responses: map[string]string{
				"https://gimmeproxy.test": "",
			},
			wantErr: true,
		},
	}

	sources := []ProxySource{
		{Name: "gimmeproxy", URL: "https://gimmeproxy.test"},
		{Name: "hproxy", URL: "https://hproxy.test"},
		{Name: "proxyscraper", URL: "https://proxyscraper.test"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fetcher := &stubFetcher{fetch: func(_ context.Context, url string) ([]byte, error) {
				body, ok := tc.responses[url]
				if !ok || body == "" {
					return nil, errors.New("source unavailable")
				}
				return []byte(body), nil
			}}
			collector := NewProxyCollector(fetcher, sources, 0)

			got, err := collector.Collect(context.Background())
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Collect() = %v, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Collect() failed: %v", err)
			}
			if len(got) != len(tc.wantURLs) {
				t.Fatalf("Collect() returned %d proxies, want %d: %+v", len(got), len(tc.wantURLs), got)
			}
			for i, want := range tc.wantURLs {
				if got[i].URL != want {
					t.Errorf("proxy[%d].URL = %q, want %q", i, got[i].URL, want)
				}
			}
		})
	}
}

func TestProxyCollectorRespectsLimit(t *testing.T) {
	t.Parallel()

	fetcher := &stubFetcher{fetch: func(_ context.Context, _ string) ([]byte, error) {
		return []byte("http://1.1.1.1:8080\nhttp://2.2.2.2:8080\nhttp://3.3.3.3:8080\n"), nil
	}}
	collector := NewProxyCollector(fetcher, []ProxySource{{Name: "test", URL: "https://proxy.test"}}, 2)

	got, err := collector.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() failed: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("Collect() returned %d proxies, want 2", len(got))
	}
	if got[0] != (models.Proxy{URL: "http://1.1.1.1:8080", Source: "test"}) {
		t.Errorf("unexpected first proxy: %+v", got[0])
	}
}
