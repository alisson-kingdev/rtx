package models

type SystemProxyResult struct {
	Action       string
	ProxyURL     string
	Shell        string
	FilesUpdated []string
	Success      bool
	Error        error
}
