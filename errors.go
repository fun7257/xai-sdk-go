package xai

import (
	"github.com/fun7257/xai-sdk-go/collections"
	"github.com/fun7257/xai-sdk-go/internal/conn"
)

// Sentinel errors for client construction and domain configuration.
// Failure paths wrap with %w so callers can use errors.Is.
var (
	// ErrNoAPIKey is returned when NewClient has no key from options or XAI_API_KEY.
	ErrNoAPIKey = conn.ErrNoAPIKey

	// ErrEmptyAPIKey is returned when an empty key is used for a dialed connection.
	ErrEmptyAPIKey = conn.ErrEmptyAPIKey

	// ErrNoManagementKey is returned when Collections CRUD needs a management key
	// but none was configured (XAI_MANAGEMENT_KEY / WithManagementAPIKey / WithManagementConn).
	ErrNoManagementKey = collections.ErrNoManagementKey
)
