package services

import (
	"context"
	"errors"
	"fmt"

	"rtx/src/models"
	"rtx/src/utils"
)

// ProxySource identifies a free public proxy-list endpoint.
type ProxySource struct {
	Name string
	URL  string
}

// DefaultProxySources are public, keyless endpoints used by default.
var DefaultProxySources = []ProxySource{
	{Name: "gimmeproxy", URL: "https://gimmeproxy.com/api/getProxy?curl=true"},
	{Name: "hproxy", URL: "https://hproxy.com/api/proxy-list?format=json"},
	{Name: "proxyscraper", URL: "https://api.proxyscraper.com/v1/http.txt"},
}

// DefaultMaxProxies bounds the number of candidates collected per rotation.
const DefaultMaxProxies = 24

// ProxyCollector gathers and validates proxies from public sources.
type ProxyCollector struct {
	fetcher utils.HTTPFetcher
	sources []ProxySource
	max     int
}

// NewProxyCollector builds a collector; nil fetcher and empty sources use defaults.
func NewProxyCollector(fetcher utils.HTTPFetcher, sources []ProxySource, max int) *ProxyCollector {
	if fetcher == nil {
		fetcher = utils.DefaultFetcher
	}
	if len(sources) == 0 {
		sources = DefaultProxySources
	}
	if max <= 0 {
		max = DefaultMaxProxies
	}
	return &ProxyCollector{
		fetcher: fetcher,
		sources: append([]ProxySource(nil), sources...),
		max:     max,
	}
}

// Collect returns unique, validated proxy candidates.
func (c *ProxyCollector) Collect(ctx context.Context) ([]models.Proxy, error) {
	seen := make(map[string]struct{})
	proxies := make([]models.Proxy, 0, c.max)
	var errs []error

	for _, source := range c.sources {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if len(proxies) >= c.max {
			break
		}

		body, err := c.fetcher.Fetch(ctx, source.URL)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", source.Name, err))
			continue
		}

		for _, raw := range utils.ExtractProxyURLs(string(body)) {
			normalized, err := utils.NormalizeProxyURL(raw)
			if err != nil {
				continue
			}
			if _, ok := seen[normalized]; ok {
				continue
			}
			seen[normalized] = struct{}{}
			proxies = append(proxies, models.Proxy{URL: normalized, Source: source.Name})
			if len(proxies) >= c.max {
				break
			}
		}
	}

	if len(proxies) == 0 {
		if len(errs) > 0 {
			return nil, fmt.Errorf("collect public proxies: %w", errors.Join(errs...))
		}
		return nil, errors.New("collect public proxies: no candidates found")
	}
	return proxies, nil
}

// CollectProxies gathers proxies with the default free sources.
func CollectProxies(ctx context.Context) ([]models.Proxy, error) {
	return NewProxyCollector(nil, nil, 0).Collect(ctx)
}
