package models

type Proxy struct {
	URL    string
	Source string
}

type IPInfo struct {
	OldIP       string
	NewIP       string
	ProxyURL    string
	ProxySource string
	Success     bool
}

type RotationResult struct {
	OldIP   string
	NewIP   string
	Proxy   string
	Source  string
	Changed bool
	Error   error
}
