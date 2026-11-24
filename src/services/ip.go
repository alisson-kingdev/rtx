package services

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"rtx/src/utils"
)

// DefaultIPEndpoints are free public services that echo the caller's IP.
var DefaultIPEndpoints = []string{
	"https://api.ipify.org",
	"https://icanhazip.com",
	"https://ifconfig.me",
	"https://api.ip.sb/ip",
	"https://ipinfo.io/ip",
}

// IPLookup resolves the public IP directly or through a proxy.
type IPLookup struct {
	fetcher   utils.HTTPFetcher
	endpoints []string
}

// NewIPLookup builds an IPLookup; nil arguments select the free defaults.
func NewIPLookup(fetcher utils.HTTPFetcher, endpoints []string) *IPLookup {
	if fetcher == nil {
		fetcher = utils.DefaultFetcher
	}
	if len(endpoints) == 0 {
		endpoints = DefaultIPEndpoints
	}
	return &IPLookup{
		fetcher:   fetcher,
		endpoints: append([]string(nil), endpoints...),
	}
}

// PublicIP returns the current public IP, trying each endpoint in order.
func (l *IPLookup) PublicIP(ctx context.Context) (string, error) {
	return l.lookup(ctx, "")
}

// PublicIPViaProxy returns the public IP observed through a validated proxy.
func (l *IPLookup) PublicIPViaProxy(ctx context.Context, proxyURL string) (string, error) {
	normalized, err := utils.NormalizeProxyURL(proxyURL)
	if err != nil {
		return "", fmt.Errorf("parse proxy URL: %w", err)
	}
	return l.lookup(ctx, normalized)
}

func (l *IPLookup) lookup(ctx context.Context, proxyURL string) (string, error) {
	var errs []error

	for _, endpoint := range l.endpoints {
		if err := ctx.Err(); err != nil {
			return "", err
		}

		var (
			body []byte
			err  error
		)
		if proxyURL == "" {
			body, err = l.fetcher.Fetch(ctx, endpoint)
		} else {
			body, err = l.fetcher.FetchViaProxy(ctx, endpoint, proxyURL)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", endpoint, err))
			continue
		}

		candidate := strings.TrimSpace(string(body))
		parsed := net.ParseIP(candidate)
		if parsed == nil {
			errs = append(errs, fmt.Errorf("%s: invalid IP response %q", endpoint, candidate))
			continue
		}
		return parsed.String(), nil
	}

	if len(errs) == 0 {
		return "", errors.New("no public IP endpoints configured")
	}
	return "", errors.Join(errs...)
}

// GetPublicIP resolves the current IP using the default free endpoints.
func GetPublicIP(ctx context.Context) (string, error) {
	return NewIPLookup(nil, nil).PublicIP(ctx)
}

// GetPublicIPViaProxy resolves the IP through a proxy using default endpoints.
func GetPublicIPViaProxy(ctx context.Context, proxyURL string) (string, error) {
	return NewIPLookup(nil, nil).PublicIPViaProxy(ctx, proxyURL)
}
