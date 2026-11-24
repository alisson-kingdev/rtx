package services

import (
	"strings"

	"rtx/src/utils"
)

var IPAPIs = []string{
	"https://api.ipify.org",
	"https://icanhazip.com",
	"https://ifconfig.me",
	"https://api.ip.sb/ip",
	"https://ipinfo.io/ip",
}

// Injective dependencies for testing
var (
	fetchURLFn        = utils.FetchURL
	fetchURLWithProxyFn = utils.FetchURLWithProxy
)

func GetPublicIP() (string, error) {
	for _, api := range IPAPIs {
		body, err := fetchURLFn(api)
		if err != nil {
			continue
		}
		ip := strings.TrimSpace(body)
		if ip != "" {
			return ip, nil
		}
	}
	return "", nil
}

func GetPublicIPViaProxy(proxyURL string) (string, error) {
	for _, api := range IPAPIs {
		body, err := fetchURLWithProxyFn(api, proxyURL)
		if err != nil {
			continue
		}
		ip := strings.TrimSpace(body)
		if ip != "" {
			return ip, nil
		}
	}
	return "", nil
}
