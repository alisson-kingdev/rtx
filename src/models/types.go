package models

// Proxy represents a proxy candidate collected from a public source.
type Proxy struct {
	URL    string
	Source string
}

// RotationResult describes the outcome of an IP rotation attempt.
type RotationResult struct {
	OldIP   string
	NewIP   string
	Proxy   string
	Source  string
	Changed bool
}
