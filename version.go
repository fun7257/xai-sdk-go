// Package xai is the idiomatic Go client for the xAI API (gRPC).
//
// Construct a [Client] with [NewClient], then use domain clients such as
// Client.Chat, Client.Image, Client.Video, Client.Files, Client.Batch, and
// Client.Collections. The SDK version reported to the wire is resolved from
// the Go module graph (git tags), not a hand-maintained constant.
package xai

import (
	"runtime/debug"
	"strings"
	"sync"
)

// modulePath is this library's module path (must match go.mod).
const modulePath = "github.com/fun7257/xai-sdk-go"

var (
	versionOnce sync.Once
	version     string
)

// Version returns this library's module version as resolved by the Go toolchain.
//
// When consumers install a release with a git tag (for example
// `go get github.com/fun7257/xai-sdk-go@v0.1.0`), this returns that version
// without inventing a second versioning scheme in-repo. Untagged local trees
// and pure development builds return "devel".
//
// The leading "v" from module versions is stripped for wire/display consistency
// (e.g. "v0.1.0" → "0.1.0"). The value is also used as gRPC metadata
// xai-sdk-version: go/<Version>.
func Version() string {
	versionOnce.Do(func() {
		version = resolveModuleVersion()
	})
	return version
}

func resolveModuleVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok || bi == nil {
		return "devel"
	}
	// Developing this module as the main module (go test / go run in-repo).
	if bi.Main.Path == modulePath {
		return normalizeModuleVersion(bi.Main.Version)
	}
	// Consumed as a dependency: version comes from the consumer's go.mod / tag.
	for _, d := range bi.Deps {
		if d.Path == modulePath {
			return normalizeModuleVersion(d.Version)
		}
	}
	return "devel"
}

func normalizeModuleVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || v == "(devel)" {
		return "devel"
	}
	return strings.TrimPrefix(v, "v")
}
