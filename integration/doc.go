// Package integration is intentionally empty under default build tags.
// Live smoke lives in files tagged `//go:build integration` so
// `go test ./...` never dials the public API without an explicit tag.
package integration
