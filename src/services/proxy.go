package services

import (
	"rtx/src/models"
	"rtx/src/utils"
)

var ProxyAPIs = []struct {
	Name string
	URL  string
}{
	{"gimmeproxy", "https://gimmeproxy.com/api/getProxy?curl=true"},
	{"hproxy", "https://hproxy.com/api/proxy-list?format=json"},
	{"proxyscraper", "https://api.proxyscraper.com/v1/http.txt"},
}

var collectFetchURLFn = utils.FetchURL

func CollectProxies() []models.Proxy {
	var allProxies []models.Proxy

	for _, api := range ProxyAPIs {
		body, err := collectFetchURLFn(api.URL)
		if err != nil {
			continue
		}

		urls := utils.ExtractProxyURLs(body)
		for _, u := range urls {
			allProxies = append(allProxies, models.Proxy{
				URL:    u,
				Source: api.Name,
			})
		}
	}

	return allProxies
}
