package scan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mk(t *testing.T, root, path, body string) {
	t.Helper()
	file := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
}
func byID(t *testing.T, r Report, id string) Finding {
	t.Helper()
	for _, f := range r.Findings {
		if f.ID == id {
			return f
		}
	}
	t.Fatalf("no finding %s", id)
	return Finding{}
}
func TestEmptyNeverGreen(t *testing.T) {
	r, err := Scan(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if r.Summary.Verified != 0 {
		t.Fatalf("empty repo incorrectly verified %d", r.Summary.Verified)
	}
	if byID(t, r, "BRANCH_PROTECTION").Status != AttestationNeeded {
		t.Fatal("offline claim should remain unknown")
	}
	if r.Summary.Total < 12 {
		t.Fatalf("expected broad evidence coverage, got %d", r.Summary.Total)
	}
}
func TestVerifiedArtifactNotControl(t *testing.T) {
	root := t.TempDir()
	mk(t, root, ".github/SECURITY.md", "Report issues by private disclosure.")
	r, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	f := byID(t, r, "SECURITY_POLICY")
	if f.Status != Verified || len(f.Evidence) != 1 || len(f.Evidence[0].SHA256) != 64 {
		t.Fatalf("evidence failure: %+v", f)
	}
	if !strings.Contains(r.ScopeNotice, "not that a framework control is satisfied") {
		t.Fatal("missing disclaimer")
	}
}
func TestUnpinnedActionIsGap(t *testing.T) {
	root := t.TempDir()
	mk(t, root, ".github/workflows/ci.yml", "name: CI\non: push\njobs:\n  test:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n")
	r, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if byID(t, r, "ACTION_PINNING").Status != Gap {
		t.Fatalf("unpinned action was not recognized as gap: %+v", byID(t, r, "ACTION_PINNING"))
	}
}
func TestPinnedActionAndWriteAll(t *testing.T) {
	root := t.TempDir()
	mk(t, root, ".github/workflows/ci.yml", "name: CI\npermissions: write-all\njobs:\n  test:\n    steps:\n      - uses: actions/checkout@0123456789abcdef0123456789abcdef01234567\n      - uses: ./local\n")
	r, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if byID(t, r, "ACTION_PINNING").Status != Verified {
		t.Fatalf("pin: %+v", byID(t, r, "ACTION_PINNING"))
	}
	if byID(t, r, "WORKFLOW_PERMISSIONS").Status != Gap {
		t.Fatal("write-all not recognized")
	}
}
func TestArtifactSymlinkNotFollowed(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(outside, []byte("do not read"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "SECURITY.md")); err != nil {
		t.Skip(err)
	}
	r, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if byID(t, r, "SECURITY_POLICY").Status == Verified {
		t.Fatal("followed symlink into external file")
	}
}
func TestScopeNeverLooksAtGit(t *testing.T) {
	root := t.TempDir()
	mk(t, root, ".git/hooks/SECURITY.md", "fake")
	mk(t, root, "node_modules/foo_test.go", "fake")
	r, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if byID(t, r, "TEST_ARTIFACTS").Status != Missing {
		t.Fatal("counted ignored dependency files")
	}
}
func TestNonDirFails(t *testing.T) {
	p := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(p, []byte("hi"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Scan(p); err == nil {
		t.Fatal("non-dir accepted")
	}
}
