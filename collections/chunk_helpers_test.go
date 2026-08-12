package collections_test

import (
	"testing"

	"github.com/fun7257/xai-sdk-go/collections"
)

// Chunk helpers must produce configurations that pass validation and carry
// the shared flags.
func TestChunkConstructors(t *testing.T) {
	chars := collections.ChunkByChars(1000, collections.WithStripWhitespace(true))
	if err := collections.ValidateChunkConfiguration(chars); err != nil {
		t.Fatal(err)
	}
	if chars.GetCharsConfiguration().GetMaxChunkSizeChars() != 1000 || !chars.GetStripWhitespace() {
		t.Fatalf("%+v", chars)
	}

	tokens := collections.ChunkByTokens(256, collections.WithInjectNameIntoChunks(true))
	if err := collections.ValidateChunkConfiguration(tokens); err != nil {
		t.Fatal(err)
	}
	if tokens.GetTokensConfiguration().GetMaxChunkSizeTokens() != 256 || !tokens.GetInjectNameIntoChunks() {
		t.Fatalf("%+v", tokens)
	}

	bytes := collections.ChunkByBytes(4096)
	if err := collections.ValidateChunkConfiguration(bytes); err != nil {
		t.Fatal(err)
	}
	if bytes.GetBytesConfiguration().GetMaxChunkSizeBytes() != 4096 {
		t.Fatalf("%+v", bytes)
	}
	if bytes.GetStripWhitespace() || bytes.GetInjectNameIntoChunks() {
		t.Fatal("flags must default to false")
	}
}

// Overlap applies to whichever strategy is active; encoding only to tokens.
func TestChunkOverlapAndEncoding(t *testing.T) {
	chars := collections.ChunkByChars(1000, collections.WithChunkOverlap(100))
	if chars.GetCharsConfiguration().GetChunkOverlapChars() != 100 {
		t.Fatalf("%+v", chars)
	}
	tokens := collections.ChunkByTokens(256,
		collections.WithChunkOverlap(32),
		collections.WithTokensEncodingName("cl100k_base"),
	)
	tc := tokens.GetTokensConfiguration()
	if tc.GetChunkOverlapTokens() != 32 || tc.GetEncodingName() != "cl100k_base" {
		t.Fatalf("%+v", tc)
	}
	bytes := collections.ChunkByBytes(4096, collections.WithChunkOverlap(512))
	if bytes.GetBytesConfiguration().GetChunkOverlapBytes() != 512 {
		t.Fatalf("%+v", bytes)
	}
	// Encoding name is a no-op for non-token strategies.
	chars2 := collections.ChunkByChars(10, collections.WithTokensEncodingName("x"))
	if chars2.GetCharsConfiguration() == nil {
		t.Fatal("chars config lost")
	}
}
