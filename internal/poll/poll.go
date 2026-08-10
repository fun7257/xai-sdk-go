// Package poll provides context-aware helpers for deferred/long-running jobs.
package poll

import (
	"context"
	"fmt"
	"time"
)

// Config controls deferred/long-running polling.
type Config struct {
	Timeout  time.Duration
	Interval time.Duration
	Context  string
}

// Default is 10m timeout and 100ms interval.
func Default() Config {
	return Config{Timeout: 10 * time.Minute, Interval: 100 * time.Millisecond}
}

// Wait calls fn until done, error, context cancel, or timeout.
func Wait(ctx context.Context, cfg Config, fn func(context.Context) (done bool, err error)) error {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Minute
	}
	if cfg.Interval <= 0 {
		cfg.Interval = time.Second
	}
	deadline := time.Now().Add(cfg.Timeout)
	for {
		done, err := fn(ctx)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		if time.Now().After(deadline) {
			msg := fmt.Sprintf("polling timed out after %s", cfg.Timeout)
			if cfg.Context != "" {
				msg += ": " + cfg.Context
			}
			return fmt.Errorf("%s", msg)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(cfg.Interval):
		}
	}
}
