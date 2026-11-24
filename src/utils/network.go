package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	// DefaultRequestTimeout limits how long a single HTTP request may take.
	DefaultRequestTimeout = 10 * time.Second

	maxResponseBytes = 1 << 20 // 1 MiB is plenty for IP and proxy-list endpoints.
)

// ErrCrossHostRedirect is returned when an endpoint tries to redirect to another host.
var ErrCrossHostRedirect = errors.New("redirect to a different host is not allowed")

// HTTPFetcher retrieves remote content, optionally through a proxy.
type HTTPFetcher interface {
	Fetch(ctx context.Context, targetURL string) ([]byte, error)
	FetchViaProxy(ctx context.Context, targetURL, proxyURL string) ([]byte, error)
}

// HTTPClient is the default HTTPFetcher implementation with bounded responses.
type HTTPClient struct {
	timeout  time.Duration
	maxBytes int64
}

// NewHTTPClient builds a client with a per-request timeout and response cap.
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	if timeout <= 0 {
		timeout = DefaultRequestTimeout
	}
	return &HTTPClient{timeout: timeout, maxBytes: maxResponseBytes}
}

// DefaultFetcher is used by the convenience functions and service wrappers.
var DefaultFetcher HTTPFetcher = NewHTTPClient(DefaultRequestTimeout)

// Fetch retrieves a URL without a proxy.
func (c *HTTPClient) Fetch(ctx context.Context, targetURL string) ([]byte, error) {
	return c.do(ctx, targetURL, nil)
}

// FetchViaProxy retrieves a URL through a validated proxy.
func (c *HTTPClient) FetchViaProxy(ctx context.Context, targetURL, proxyURL string) ([]byte, error) {
	proxy, err := ParseProxyURL(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("parse proxy URL: %w", err)
	}
	return c.do(ctx, targetURL, proxy)
}

func (c *HTTPClient) do(ctx context.Context, targetURL string, proxy *url.URL) ([]byte, error) {
	requestCtx := ctx
	if c.timeout > 0 {
		var cancel context.CancelFunc
		requestCtx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if proxy != nil {
		transport.Proxy = http.ProxyURL(proxy)
		defer transport.CloseIdleConnections()
	}

	client := &http.Client{
		Transport:     transport,
		CheckRedirect: sameHostRedirects(req.URL),
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", req.URL.Redacted(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("GET %s: unexpected status %s", req.URL.Redacted(), resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, c.maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read response from %s: %w", req.URL.Redacted(), err)
	}
	if int64(len(body)) > c.maxBytes {
		return nil, fmt.Errorf("response from %s exceeds %d bytes", req.URL.Redacted(), c.maxBytes)
	}

	return body, nil
}

func sameHostRedirects(initial *url.URL) func(*http.Request, []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return errors.New("too many redirects")
		}
		if req.URL.Host != initial.Host || req.URL.Scheme != initial.Scheme {
			return ErrCrossHostRedirect
		}
		return nil
	}
}

// FetchURL is a compatibility wrapper around DefaultFetcher.
func FetchURL(targetURL string) (string, error) {
	body, err := DefaultFetcher.Fetch(context.Background(), targetURL)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// FetchURLWithProxy is a compatibility wrapper around DefaultFetcher.
func FetchURLWithProxy(targetURL, proxyURL string) (string, error) {
	body, err := DefaultFetcher.FetchViaProxy(context.Background(), targetURL, proxyURL)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

var (
	jsonURLPattern      = regexp.MustCompile(`https?://[^\s"'{},]+`)
	jsonHostPortPattern = regexp.MustCompile(`"([A-Za-z0-9.-]+:[0-9]+)"`)
)

// ExtractProxyURLs finds and normalizes proxy candidates in text or JSON responses.
func ExtractProxyURLs(text string) []string {
	seen := make(map[string]struct{})
	var proxies []string

	addCandidate := func(candidate string) {
		candidate = strings.TrimSpace(strings.Trim(candidate, `,;"`))
		if candidate == "" {
			return
		}
		if !strings.Contains(candidate, "://") {
			candidate = "http://" + candidate
		}
		proxy, err := ParseProxyURL(candidate)
		if err != nil {
			return
		}
		normalized := proxy.String()
		if _, ok := seen[normalized]; ok {
			return
		}
		seen[normalized] = struct{}{}
		proxies = append(proxies, normalized)
	}

	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, `"`) {
			for _, match := range jsonURLPattern.FindAllString(line, -1) {
				addCandidate(match)
			}
			for _, match := range jsonHostPortPattern.FindAllStringSubmatch(line, -1) {
				if len(match) == 2 {
					addCandidate(match[1])
				}
			}
			continue
		}
		addCandidate(line)
	}

	return proxies
}
