package services

import (
	"errors"
	"testing"
)

func TestGetPublicIP_Success(t *testing.T) {
	origFetch := fetchURLFn
	defer func() { fetchURLFn = origFetch }()

	fetchURLFn = func(url string) (string, error) {
		if url == "https://api.ipify.org" {
			return "1.2.3.4\n", nil
		}
		return "", errors.New("fail")
	}

	ip, err := GetPublicIP()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "1.2.3.4" {
		t.Errorf("expected 1.2.3.4, got %s", ip)
	}
}

func TestGetPublicIP_AllFail(t *testing.T) {
	origFetch := fetchURLFn
	defer func() { fetchURLFn = origFetch }()

	fetchURLFn = func(url string) (string, error) {
		return "", errors.New("network down")
	}

	ip, err := GetPublicIP()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "" {
		t.Errorf("expected empty, got %s", ip)
	}
}

func TestGetPublicIP_EmptyResponse(t *testing.T) {
	origFetch := fetchURLFn
	defer func() { fetchURLFn = origFetch }()

	fetchURLFn = func(url string) (string, error) {
		return "   \n", nil
	}

	ip, err := GetPublicIP()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "" {
		t.Errorf("expected empty, got %s", ip)
	}
}

func TestGetPublicIPViaProxy_Success(t *testing.T) {
	origFetch := fetchURLWithProxyFn
	defer func() { fetchURLWithProxyFn = origFetch }()

	fetchURLWithProxyFn = func(targetURL, proxyURL string) (string, error) {
		if proxyURL == "http://1.2.3.4:8080" {
			return "5.6.7.8", nil
		}
		return "", errors.New("fail")
	}

	ip, err := GetPublicIPViaProxy("http://1.2.3.4:8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "5.6.7.8" {
		t.Errorf("expected 5.6.7.8, got %s", ip)
	}
}

func TestGetPublicIPViaProxy_AllFail(t *testing.T) {
	origFetch := fetchURLWithProxyFn
	defer func() { fetchURLWithProxyFn = origFetch }()

	fetchURLWithProxyFn = func(targetURL, proxyURL string) (string, error) {
		return "", errors.New("fail")
	}

	ip, err := GetPublicIPViaProxy("http://1.2.3.4:8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "" {
		t.Errorf("expected empty, got %s", ip)
	}
}