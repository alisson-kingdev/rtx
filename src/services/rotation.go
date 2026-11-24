package services

import (
	"rtx/src/models"
)

func RotateIP() models.RotationResult {
	result := models.RotationResult{}

	oldIP, err := GetPublicIP()
	if err != nil || oldIP == "" {
		if err != nil {
			result.Error = err
		}
		return result
	}
	result.OldIP = oldIP

	proxies := CollectProxies()

	if len(proxies) == 0 {
		newIP, err := GetPublicIP()
		if err != nil {
			result.Error = err
			return result
		}
		result.NewIP = newIP
		result.Proxy = "Nenhuma (API direta)"
		result.Source = "api"
		result.Changed = newIP != oldIP
		return result
	}

	for _, proxy := range proxies {
		newIP, err := GetPublicIPViaProxy(proxy.URL)
		if err != nil || newIP == "" || newIP == oldIP {
			continue
		}
		result.NewIP = newIP
		result.Proxy = proxy.URL
		result.Source = proxy.Source
		result.Changed = true
		return result
	}

	// Fallback: rotacao via API direta
	newIP, err := GetPublicIP()
	if err != nil {
		result.Error = err
		return result
	}
	result.NewIP = newIP
	result.Proxy = "Nenhuma (API direta)"
	result.Source = "api"
	result.Changed = newIP != oldIP

	return result
}
