package services

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func rotationTestDeps(fetcher *stubFetcher) (*IPLookup, *ProxyCollector) {
	ips := NewIPLookup(fetcher, []string{"https://ip.test"})
	collector := NewProxyCollector(fetcher, []ProxySource{{Name: "test", URL: "https://proxy.test"}}, 4)
	return ips, collector
}

func TestRotatorUsesWorkingProxy(t *testing.T) {
	t.Parallel()

	fetcher := &stubFetcher{
		fetch: func(_ context.Context, url string) ([]byte, error) {
			if url == "https://proxy.test" {
				return []byte("http://2.2.2.2:8080\nbad-candidate\n"), nil
			}
			return []byte("1.1.1.1"), nil
		},
		via: func(_ context.Context, _, proxyURL string) ([]byte, error) {
			if proxyURL != "http://2.2.2.2:8080" {
				return nil, errors.New("proxy failed")
			}
			return []byte("9.9.9.9"), nil
		},
	}

	ips, collector := rotationTestDeps(fetcher)
	rotator := NewRotator(ips, collector, RotationConfig{Timeout: 2 * time.Second, Concurrency: 2})

	got, err := rotator.Rotate(context.Background())
	if err != nil {
		t.Fatalf("Rotate() failed: %v", err)
	}
	if !got.Changed || got.NewIP != "9.9.9.9" || got.Proxy != "http://2.2.2.2:8080" || got.Source != "test" {
		t.Errorf("unexpected rotation result: %+v", got)
	}
}

func TestRotatorFallsBackToDirectAPIs(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	calls := 0
	fetcher := &stubFetcher{
		fetch: func(_ context.Context, url string) ([]byte, error) {
			if url == "https://proxy.test" {
				return []byte("http://2.2.2.2:8080"), nil
			}
			mu.Lock()
			defer mu.Unlock()
			calls++
			if calls == 1 {
				return []byte("1.1.1.1"), nil
			}
			return []byte("5.5.5.5"), nil
		},
		via: func(_ context.Context, _, _ string) ([]byte, error) {
			return []byte("1.1.1.1"), nil
		},
	}

	ips, collector := rotationTestDeps(fetcher)
	rotator := NewRotator(ips, collector, RotationConfig{})

	got, err := rotator.Rotate(context.Background())
	if err != nil {
		t.Fatalf("Rotate() failed: %v", err)
	}
	if !got.Changed || got.NewIP != "5.5.5.5" || got.Source != "api" {
		t.Errorf("unexpected fallback result: %+v", got)
	}
}

func TestRotatorFallsBackAfterProxyTimeout(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	calls := 0
	fetcher := &stubFetcher{
		fetch: func(_ context.Context, url string) ([]byte, error) {
			if url == "https://proxy.test" {
				return []byte("http://2.2.2.2:8080"), nil
			}
			mu.Lock()
			defer mu.Unlock()
			calls++
			if calls == 1 {
				return []byte("1.1.1.1"), nil
			}
			return []byte("5.5.5.5"), nil
		},
		via: func(ctx context.Context, _, _ string) ([]byte, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		},
	}

	ips, collector := rotationTestDeps(fetcher)
	rotator := NewRotator(ips, collector, RotationConfig{Timeout: 20 * time.Millisecond, Concurrency: 1})

	got, err := rotator.Rotate(context.Background())
	if err != nil {
		t.Fatalf("Rotate() failed: %v", err)
	}
	if got.Source != "api" || got.NewIP != "5.5.5.5" {
		t.Errorf("unexpected timeout fallback result: %+v", got)
	}
}

func TestRotatorReportsCurrentIPFailure(t *testing.T) {
	t.Parallel()

	fetcher := &stubFetcher{fetch: func(_ context.Context, url string) ([]byte, error) {
		if url == "https://ip.test" {
			return nil, errors.New("network down")
		}
		return []byte("http://2.2.2.2:8080"), nil
	}}

	ips, collector := rotationTestDeps(fetcher)
	rotator := NewRotator(ips, collector, RotationConfig{})

	if _, err := rotator.Rotate(context.Background()); err == nil {
		t.Fatal("Rotate() succeeded, want error")
	}
}

func TestRotatorHonorsCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fetcher := &stubFetcher{}
	ips, collector := rotationTestDeps(fetcher)
	rotator := NewRotator(ips, collector, RotationConfig{})

	if _, err := rotator.Rotate(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}
