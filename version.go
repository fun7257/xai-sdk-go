// Package xai is the idiomatic Go client for the xAI API (gRPC).
//
// Construct a [Client] with [NewClient], then use domain clients such as
// Client.Chat, Client.Image, Client.Video, Client.Files, Client.Batch, and
// Client.Collections. Version is sent as the xai-sdk-version metadata value.
package xai

// Version is the SDK semantic version (semver). Bump with CHANGELOG entries.
const Version = "0.2.0"
