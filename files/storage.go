package files

import (
	"time"

	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// StorageOptions describes how to persist a generated asset via the Files API.
// Mapped to xaiv1.StorageOptions when present on generate RPCs.
type StorageOptions struct {
	Filename     string
	ExpiresAfter *time.Duration
	PublicURL    bool
	// PublicURLExpiresAfter sets independent public URL TTL when PublicURL is true.
	PublicURLExpiresAfter *time.Duration
}

// Proto converts StorageOptions to the API message. Returns nil if o is nil.
// A non-nil zero value still produces an (empty) message: passing
// StorageOptions{} to WithStorage explicitly requests storage with server
// defaults rather than disabling it.
func (o *StorageOptions) Proto() *xaiv1.StorageOptions {
	if o == nil {
		return nil
	}
	out := &xaiv1.StorageOptions{Filename: o.Filename}
	if o.ExpiresAfter != nil {
		s := int64(o.ExpiresAfter.Seconds())
		out.ExpiresAfter = &s
	}
	if o.PublicURL {
		pu := &xaiv1.PublicUrlOptions{}
		if o.PublicURLExpiresAfter != nil {
			s := int64(o.PublicURLExpiresAfter.Seconds())
			pu.ExpiresAfter = &s
		}
		out.PublicUrl = pu
	}
	return out
}

// StorageFromProto converts a proto StorageOptions to the SDK type.
func StorageFromProto(p *xaiv1.StorageOptions) *StorageOptions {
	if p == nil {
		return nil
	}
	out := &StorageOptions{Filename: p.Filename}
	if p.ExpiresAfter != nil {
		d := time.Duration(*p.ExpiresAfter) * time.Second
		out.ExpiresAfter = &d
	}
	if p.PublicUrl != nil {
		out.PublicURL = true
		if p.PublicUrl.ExpiresAfter != nil {
			d := time.Duration(*p.PublicUrl.ExpiresAfter) * time.Second
			out.PublicURLExpiresAfter = &d
		}
	}
	return out
}
