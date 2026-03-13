package fake

import "context"

const defaultResponse = "fake response"

// LLM is a test double that returns a configurable canned response.
type LLM struct {
	Response string // returned by Complete; defaults to "fake response"
}

// Complete returns the configured Response (or the default if empty).
func (l *LLM) Complete(_ context.Context, _ string) (string, error) {
	if l.Response != "" {
		return l.Response, nil
	}
	return defaultResponse, nil
}
