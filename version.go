// Package xai is the idiomatic Go client for the xAI API (gRPC).
//
// Construct a [Client] with [NewClient], then use domain clients such as
// Client.Chat, Client.Image, Client.Video, Client.Files, Client.Batch, and
// Client.Collections. Version is sent as the xai-sdk-version metadata value.
package xai

// Version is the SDK semantic version (semver) and the single public version
// source of truth. Bump this for releases (see docs/RELEASE.md) together with
// CHANGELOG and the matching default in internal/conn (wire metadata).
// Root package init copies Version into conn.SDKVersion for gRPC metadata.
const Version = "0.2.0"
