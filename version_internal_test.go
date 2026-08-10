package xai

import "testing"

func TestNormalizeModuleVersion(t *testing.T) {
	cases := map[string]string{
		"":         "devel",
		"(devel)":  "devel",
		"v0.1.0":   "0.1.0",
		"0.3.1":    "0.3.1",
		" v1.0.0 ": "1.0.0",
	}
	for in, want := range cases {
		if got := normalizeModuleVersion(in); got != want {
			t.Errorf("normalizeModuleVersion(%q)=%q want %q", in, got, want)
		}
	}
}

func TestResolveModuleVersionNonEmpty(t *testing.T) {
	if got := resolveModuleVersion(); got == "" {
		t.Fatal("resolveModuleVersion empty")
	}
}
