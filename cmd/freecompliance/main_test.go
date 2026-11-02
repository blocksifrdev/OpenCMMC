package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanAndCompare(t *testing.T) {
	root := t.TempDir()
	out := t.TempDir()
	oldFile := filepath.Join(out, "old.json")
	newFile := filepath.Join(out, "new.json")
	htmlFile := filepath.Join(out, "report.html")
	if err := run([]string{"scan", "--path", root, "--json", oldFile}, new(bytes.Buffer)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "SECURITY.md"), []byte("Report issues securely."), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"scan", "--path", root, "--json", newFile, "--html", htmlFile}, new(bytes.Buffer)); err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	if err := run([]string{"compare", "--before", oldFile, "--after", newFile}, &b); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "SECURITY_POLICY: MISSING -> VERIFIED") {
		t.Fatal(b.String())
	}
	data, err := os.ReadFile(htmlFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "FreeCompliance") {
		t.Fatal("html missing brand")
	}
}
func TestFailOnObservedGap(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".github", "workflows", "test.yml")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("- uses: actions/checkout@v4\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"scan", "--path", root, "--fail-on-gap"}, new(bytes.Buffer)); err == nil {
		t.Fatal("expected gap failure")
	}
}
