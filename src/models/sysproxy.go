package models

// SystemProxyResult describes the outcome of a system proxy configuration change.
// ProxyURL is stored redacted when it contains credentials.
type SystemProxyResult struct {
	Action       string
	ProxyURL     string
	Shell        string
	FilesUpdated []string
}
