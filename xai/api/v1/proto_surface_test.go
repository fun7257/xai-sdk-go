package xaiv1_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Ensure out-of-scope Embed / Sample RPC generated surface stays deleted after cleanup.
func TestNoEmbedOrSampleRPCGeneratedFiles(t *testing.T) {
	for _, name := range []string{
		"embed.pb.go",
		"embed_grpc.pb.go",
		"sample_grpc.pb.go",
	} {
		if _, err := os.Stat(name); err == nil {
			t.Errorf("redundant generated file must not exist: %s", name)
		} else if !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat("sample.pb.go"); err != nil {
		t.Fatalf("sample.pb.go required for FinishReason: %v", err)
	}
}

// FinishReason remains; SampleText* (Sample RPC) must not.
func TestSampleProtoIsEnumOnly(t *testing.T) {
	src, err := os.ReadFile("sample.pb.go")
	if err != nil {
		t.Fatal(err)
	}
	s := string(src)
	if !strings.Contains(s, "type FinishReason ") {
		t.Fatal("FinishReason type missing from sample.pb.go")
	}
	if strings.Contains(s, "SampleTextRequest") {
		t.Fatal("SampleTextRequest should have been removed with Sample RPC")
	}
}

// third_party must not reintroduce embed.proto.
func TestNoEmbedProtoInThirdParty(t *testing.T) {
	// package dir is xai/api/v1; module root is ../../..
	root := filepath.Join("..", "..", "..")
	p := filepath.Join(root, "third_party", "xai", "api", "v1", "embed.proto")
	if _, err := os.Stat(p); err == nil {
		t.Fatalf("embed.proto must stay deleted (out of scope): %s", p)
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
}
