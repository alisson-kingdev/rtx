package services

import (
	"errors"
	"testing"
)

func TestCollectProxies_Success(t *testing.T) {
	origFetch := collectFetchURLFn
	defer func() { collectFetchURLFn = origFetch }()

	collectFetchURLFn = func(url string) (string, error) {
		switch url {
		case "https://gimmeproxy.com/api/getProxy?curl=true":
			return "http://1.1.1.1:8080\nhttp://2.2.2.2:8080\n", nil
		case "https://hproxy.com/api/proxy-list?format=json":
			return `{"proxies":[{"proxy":"http://3.3.3.3:8080"}]}`, nil
		case "https://api.proxyscraper.com/v1/http.txt":
			return "4.4.4.4:3128\nhttp://5.5.5.5:8080\n", nil
		}
		return "", errors.New("fail")
	}

	proxies := CollectProxies()
	if len(proxies) != 4 {
		t.Fatalf("expected 4 proxies, got %d", len(proxies))
	}
	if proxies[0].URL != "http://1.1.1.1:8080" {
		t.Errorf("expected http://1.1.1.1:8080, got %s", proxies[0].URL)
	}
	if proxies[0].Source != "gimmeproxy" {
		t.Errorf("expected source gimmeproxy, got %s", proxies[0].Source)
	}
	if proxies[3].URL != "http://5.5.5.5:8080" {
		t.Errorf("expected http://5.5.5.5:8080, got %s", proxies[3].URL)
	}
}

func TestCollectProxies_AllFail(t *testing.T) {
	origFetch := collectFetchURLFn
	defer func() { collectFetchURLFn = origFetch }()

	collectFetchURLFn = func(url string) (string, error) {
		return "", errors.New("fail")
	}

	proxies := CollectProxies()
	if len(proxies) != 0 {
		t.Fatalf("expected 0 proxies, got %d", len(proxies))
	}
}

func TestCollectProxies_EmptyBody(t *testing.T) {
	origFetch := collectFetchURLFn
	defer func() { collectFetchURLFn = origFetch }()

	collectFetchURLFn = func(url string) (string, error) {
		return "", nil
	}

	proxies := CollectProxies()
	if len(proxies) != 0 {
		t.Fatalf("expected 0 proxies, got %d", len(proxies))
	}
}

func TestCollectProxies_PartialFailure(t *testing.T) {
	origFetch := collectFetchURLFn
	defer func() { collectFetchURLFn = origFetch }()

	collectFetchURLFn = func(url string) (string, error) {
		switch url {
		case "https://gimmeproxy.com/api/getProxy?curl=true":
			return "http://1.1.1.1:8080\n", nil
		case "https://hproxy.com/api/proxy-list?format=json":
			return "", errors.New("fail")
		case "https://api.proxyscraper.com/v1/http.txt":
			return "2.2.2.2:8080\n", nil
		}
		return "", errors.New("fail")
	}

	proxies := CollectProxies()
	if len(proxies) != 2 {
		t.Fatalf("expected 2 proxies, got %d", len(proxies))
	}
	if proxies[0].Source != "gimmeproxy" {
		t.Errorf("expected source gimmeproxy, got %s", proxies[0].Source)
	}
	if proxies[1].Source != "proxyscraper" {
		t.Errorf("expected source proxyscraper, got %s", proxies[1].Source)
	}
}