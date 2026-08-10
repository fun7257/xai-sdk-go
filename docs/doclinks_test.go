// Package docs is empty; this file only validates Markdown link structure.
package docs_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Relative Markdown links among project docs must resolve to existing files.
// External http(s) links and pure anchors are skipped.
func TestProjectMarkdownLinksResolve(t *testing.T) {
	root := findModuleRoot(t)
	mdFiles := []string{
		"README.md",
		"CONTRIBUTING.md",
		"CHANGELOG.md",
		"examples/README.md",
		"docs/GUIDE.zh-en.md",
		"docs/API.md",
		"docs/PARITY.md",
		"docs/DIFF.md",
		"docs/COMPATIBILITY.md",
		"docs/PROTO.md",
		"docs/SECURITY.md",
		"docs/RELEASE.md",
		"docs/CI.md",
		"docs/ROADMAP_OSS.md",
		".github/pull_request_template.md",
	}

	linkRe := regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)
	var broken []string

	for _, rel := range mdFiles {
		path := filepath.Join(root, rel)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		baseDir := filepath.Dir(path)
		for _, m := range linkRe.FindAllStringSubmatch(string(data), -1) {
			target := strings.TrimSpace(m[1])
			if target == "" || strings.HasPrefix(target, "#") {
				continue
			}
			if i := strings.IndexByte(target, '#'); i >= 0 {
				target = target[:i]
			}
			if target == "" {
				continue
			}
			if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "mailto:") {
				continue
			}
			// Skip pure code/path references that are not doc links (no extension and not dir)
			abs := filepath.Clean(filepath.Join(baseDir, target))
			if !strings.HasPrefix(abs, root) {
				// allow only under module root
				broken = append(broken, rel+": escapes root: "+m[1])
				continue
			}
			if _, err := os.Stat(abs); err != nil {
				broken = append(broken, rel+" -> "+m[1])
			}
		}
	}

	if len(broken) > 0 {
		t.Fatalf("broken in-repo markdown links (%d):\n  %s", len(broken), strings.Join(broken, "\n  "))
	}
}

// README must expose progressive structure and links for core reader paths.
func TestREADMEDocMapCoversCoreRoles(t *testing.T) {
	root := findModuleRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	needles := []string{
		"## Install",
		"## Quick start",
		"## Core patterns",
		"## Packages",
		"## Documentation",
		"## Examples",
		"## Versioning & CI",
		"## Contributing",
		"docs/GUIDE.zh-en.md",
		"docs/API.md",
		"docs/PARITY.md",
		"docs/COMPATIBILITY.md",
		"docs/SECURITY.md",
		"CONTRIBUTING.md",
		"docs/RELEASE.md",
		"docs/CI.md",
		"examples/complete",
	}
	for _, n := range needles {
		if !strings.Contains(s, n) {
			t.Errorf("README.md missing nav element %q", n)
		}
	}
}

// Each topic has one primary home; secondary files should not re-list full make gate blocks.
func TestQualityGatesSingleSourcedInContributing(t *testing.T) {
	root := findModuleRoot(t)
	contrib, err := os.ReadFile(filepath.Join(root, "CONTRIBUTING.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contrib), "make check") {
		t.Fatal("CONTRIBUTING.md must own make check gates")
	}
	// COMPATIBILITY should not duplicate the full make recipe block
	compat, err := os.ReadFile(filepath.Join(root, "docs/COMPATIBILITY.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(compat), "make lint") > 0 && strings.Count(string(compat), "make race") > 0 {
		t.Fatal("docs/COMPATIBILITY.md should link CONTRIBUTING for gates, not restate full make list")
	}
}

func findModuleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found from ", wd)
		}
		dir = parent
	}
}
