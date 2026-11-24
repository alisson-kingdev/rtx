package utils

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// ErrInvalidProxyURL is returned when a proxy candidate is unsafe or malformed.
var ErrInvalidProxyURL = errors.New("invalid proxy URL")

// ParseProxyURL validates a proxy URL and returns its normalized representation.
// Only explicit http/https proxies with a host and a valid port are accepted.
func ParseProxyURL(raw string) (*url.URL, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, fmt.Errorf("%w: empty value", ErrInvalidProxyURL)
	}
	if strings.ContainsAny(value, "\r\n\t") {
		return nil, fmt.Errorf("%w: control characters are not allowed", ErrInvalidProxyURL)
	}
	if !strings.Contains(value, "://") {
		value = "http://" + value
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidProxyURL, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("%w: unsupported scheme %q", ErrInvalidProxyURL, parsed.Scheme)
	}
	if parsed.Opaque != "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("%w: path, query and fragment are not allowed", ErrInvalidProxyURL)
	}

	host, port, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		return nil, fmt.Errorf("%w: host and port are required", ErrInvalidProxyURL)
	}
	if err := validateHost(host); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidProxyURL, err)
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return nil, fmt.Errorf("%w: invalid port %q", ErrInvalidProxyURL, port)
	}

	return parsed, nil
}

// NormalizeProxyURL returns the canonical string form of a validated proxy URL.
func NormalizeProxyURL(raw string) (string, error) {
	parsed, err := ParseProxyURL(raw)
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}

// RedactURL hides credentials so proxy URLs can be printed safely.
func RedactURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" {
		return "<invalid URL>"
	}
	return parsed.Redacted()
}

func validateHost(host string) error {
	if host == "" {
		return errors.New("empty host")
	}
	if net.ParseIP(host) != nil {
		return nil
	}
	for _, r := range host {
		isAllowed := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '.' || r == '-'
		if !isAllowed {
			return fmt.Errorf("unexpected character %q in host", r)
		}
	}
	return nil
}
