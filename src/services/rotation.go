package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"rtx/src/models"
)

const (
	// DefaultRotationTimeout bounds the whole rotation, including retries.
	DefaultRotationTimeout = 45 * time.Second
	// DefaultRotationConcurrency limits simultaneous proxy checks.
	DefaultRotationConcurrency = 4
)

// RotationConfig controls time and resource limits for a rotation.
type RotationConfig struct {
	Timeout     time.Duration
	MaxProxies  int
	Concurrency int
}

// DefaultRotationConfig returns the free, conservative defaults.
func DefaultRotationConfig() RotationConfig {
	return RotationConfig{
		Timeout:     DefaultRotationTimeout,
		MaxProxies:  DefaultMaxProxies,
		Concurrency: DefaultRotationConcurrency,
	}
}

// Rotator coordinates IP lookup, proxy collection and candidate testing.
type Rotator struct {
	ips       *IPLookup
	collector *ProxyCollector
	config    RotationConfig
}

// NewRotator builds a Rotator; nil dependencies select the free defaults.
func NewRotator(ips *IPLookup, collector *ProxyCollector, config RotationConfig) *Rotator {
	defaults := DefaultRotationConfig()
	if config.Timeout <= 0 {
		config.Timeout = defaults.Timeout
	}
	if config.MaxProxies <= 0 {
		config.MaxProxies = defaults.MaxProxies
	}
	if config.Concurrency <= 0 {
		config.Concurrency = defaults.Concurrency
	}
	if ips == nil {
		ips = NewIPLookup(nil, nil)
	}
	if collector == nil {
		collector = NewProxyCollector(nil, nil, config.MaxProxies)
	}
	return &Rotator{ips: ips, collector: collector, config: config}
}

type candidateResult struct {
	proxy models.Proxy
	ip    string
}

// Rotate performs a rotation and returns the first working proxy, if any.
func (r *Rotator) Rotate(ctx context.Context) (models.RotationResult, error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	result := models.RotationResult{}
	oldIP, err := r.ips.PublicIP(ctx)
	if err != nil {
		return result, fmt.Errorf("resolve current IP: %w", err)
	}
	result.OldIP = oldIP

	proxies, collectErr := r.collector.Collect(ctx)
	if len(proxies) == 0 {
		return r.directFallback(ctx, result, collectErr)
	}

	results := make(chan candidateResult, len(proxies))
	semaphore := make(chan struct{}, r.config.Concurrency)
	var wg sync.WaitGroup

candidateLoop:
	for _, proxy := range proxies {
		if err := ctx.Err(); err != nil {
			break
		}
		select {
		case semaphore <- struct{}{}:
		case <-ctx.Done():
			break candidateLoop
		}

		wg.Add(1)
		go func(candidate models.Proxy) {
			defer wg.Done()
			defer func() { <-semaphore }()

			ip, err := r.ips.PublicIPViaProxy(ctx, candidate.URL)
			if err != nil || ip == "" || ip == oldIP {
				return
			}
			select {
			case results <- candidateResult{proxy: candidate, ip: ip}:
			default:
			}
			cancel()
		}(proxy)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for candidate := range results {
		result.NewIP = candidate.ip
		result.Proxy = candidate.proxy.URL
		result.Source = candidate.proxy.Source
		result.Changed = candidate.ip != oldIP
		return result, nil
	}

	if err := ctx.Err(); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return result, fmt.Errorf("test proxies: %w", err)
	}

	// A proxy timeout must not prevent the free direct-API fallback.
	fallbackCtx, cancelFallback := context.WithTimeout(context.Background(), r.config.Timeout)
	defer cancelFallback()
	return r.directFallback(fallbackCtx, result, nil)
}

func (r *Rotator) directFallback(ctx context.Context, result models.RotationResult, collectErr error) (models.RotationResult, error) {
	newIP, err := r.ips.PublicIP(ctx)
	if err != nil {
		if collectErr != nil {
			return result, fmt.Errorf("collect proxies: %w; direct retry: %w", collectErr, err)
		}
		return result, fmt.Errorf("direct retry: %w", err)
	}

	result.NewIP = newIP
	result.Proxy = "Nenhuma (API direta)"
	result.Source = "api"
	result.Changed = newIP != result.OldIP
	return result, nil
}

// RotateIP runs a rotation with explicit limits using the free defaults.
func RotateIP(ctx context.Context, config RotationConfig) (models.RotationResult, error) {
	return NewRotator(nil, nil, config).Rotate(ctx)
}
