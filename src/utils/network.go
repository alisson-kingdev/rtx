package utils

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var Client = &http.Client{
	Timeout: 10 * time.Second,
}

func FetchURL(targetURL string) (string, error) {
	resp, err := Client.Get(targetURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func FetchURLWithProxy(targetURL, proxyURL string) (string, error) {
	proxyURI, err := url.Parse(proxyURL)
	if err != nil {
		return "", err
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURI),
	}
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	resp, err := client.Get(targetURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func ExtractProxyURLs(text string) []string {
	var proxies []string
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			proxies = append(proxies, line)
		} else if isIPPort(line) {
			proxies = append(proxies, "http://"+line)
		}
	}
	return proxies
}

func isIPPort(s string) bool {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return false
	}
	return len(parts[0]) > 0 && len(parts[1]) > 0
}
