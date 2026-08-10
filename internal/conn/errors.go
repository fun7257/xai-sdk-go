package conn

import "errors"

// Sentinel errors for API key resolution and dial. Re-exported by package xai.
var (
	// ErrNoAPIKey is returned when no API key is available from options or environment.
	ErrNoAPIKey = errors.New("xai: API key not provided")

	// ErrEmptyAPIKey is returned when an empty API key is supplied for dial.
	ErrEmptyAPIKey = errors.New("xai: empty API key provided")
)
