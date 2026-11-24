package services

import (
	"context"
	"errors"
)

type stubFetcher struct {
	fetch func(ctx context.Context, targetURL string) ([]byte, error)
	via   func(ctx context.Context, targetURL, proxyURL string) ([]byte, error)
}

func (f *stubFetcher) Fetch(ctx context.Context, targetURL string) ([]byte, error) {
	if f.fetch == nil {
		return nil, errors.New("unexpected direct fetch: " + targetURL)
	}
	return f.fetch(ctx, targetURL)
}

func (f *stubFetcher) FetchViaProxy(ctx context.Context, targetURL, proxyURL string) ([]byte, error) {
	if f.via == nil {
		return nil, errors.New("unexpected proxied fetch: " + proxyURL)
	}
	return f.via(ctx, targetURL, proxyURL)
}
