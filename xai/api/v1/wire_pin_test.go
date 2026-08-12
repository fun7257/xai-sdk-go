package xaiv1_test

import (
	"testing"

	"google.golang.org/protobuf/proto"

	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

func fieldLayout(m proto.Message) map[int32]string {
	md := m.ProtoReflect().Descriptor()
	out := map[int32]string{}
	fields := md.Fields()
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		out[int32(f.Number())] = string(f.Name())
	}
	return out
}

// Wire pins for messages that previously drifted from the official xAI protos
// (verified against the official Python SDK descriptors, 2026-08). Field
// numbers are the wire contract: if one of these assertions fails after a
// proto edit, the change is wire-breaking against the live API — reserved
// numbers must never be reused or compacted.
func TestWirePinsMatchOfficialAPI(t *testing.T) {
	cases := []struct {
		name string
		msg  proto.Message
		want map[int32]string
	}{
		{"CreateCollectionRequest", &xaiv1.CreateCollectionRequest{}, map[int32]string{
			1: "team_id", 2: "collection_name", 4: "index_configuration",
			5: "chunk_configuration", 7: "metric_space", 9: "field_definitions",
			10: "collection_description",
		}},
		{"CollectionMetadata", &xaiv1.CollectionMetadata{}, map[int32]string{
			1: "collection_id", 2: "collection_name", 3: "created_at",
			4: "index_configuration", 5: "chunk_configuration", 6: "documents_count",
			8: "field_definitions", 9: "collection_description", 10: "total_file_size",
		}},
		{"AddDocumentToCollectionRequest", &xaiv1.AddDocumentToCollectionRequest{}, map[int32]string{
			1: "file_id", 2: "team_id", 3: "collection_id", 5: "fields",
		}},
		{"ChunkConfiguration", &xaiv1.ChunkConfiguration{}, map[int32]string{
			1: "chars_configuration", 2: "tokens_configuration", 11: "bytes_configuration",
			3: "strip_whitespace", 4: "inject_name_into_chunks",
		}},
		{"File", &xaiv1.File{}, map[int32]string{
			1: "size", 2: "created_at", 3: "expires_at", 4: "filename", 5: "id",
			7: "public_url", 8: "public_url_expires_at",
		}},
		{"CreatePublicUrlResponse", &xaiv1.CreatePublicUrlResponse{}, map[int32]string{
			1: "public_url", 2: "expires_at",
		}},
		{"RevokePublicUrlResponse", &xaiv1.RevokePublicUrlResponse{}, map[int32]string{
			1: "file_id", 2: "revoked", 3: "public_url",
		}},
		{"GenerateCollectionDescriptionResponse", &xaiv1.GenerateCollectionDescriptionResponse{}, map[int32]string{
			1: "collection_description",
		}},
	}
	for _, tc := range cases {
		got := fieldLayout(tc.msg)
		if len(got) != len(tc.want) {
			t.Errorf("%s: %d fields, want %d: got %v", tc.name, len(got), len(tc.want), got)
			continue
		}
		for num, name := range tc.want {
			if got[num] != name {
				t.Errorf("%s: field #%d = %q, want %q", tc.name, num, got[num], name)
			}
		}
	}
}

// The chunk strategies must stay a oneof so at most one can be set.
func TestChunkConfigurationStrategiesAreOneof(t *testing.T) {
	md := (&xaiv1.ChunkConfiguration{}).ProtoReflect().Descriptor()
	o := md.Oneofs()
	if o.Len() != 1 || o.Get(0).Fields().Len() != 3 {
		t.Fatalf("expected one oneof with 3 strategies, got %d oneofs", o.Len())
	}
}
