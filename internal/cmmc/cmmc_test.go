package cmmc

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCatalogsAndFailClosed(t *testing.T) {
	for _, tc := range []struct{ level, n int }{{1, 15}, {2, 110}} {
		w, err := NewWorkspace(tc.level, "example-enclave")
		if err != nil {
			t.Fatal(err)
		}
		if len(w.Controls) != tc.n {
			t.Fatalf("level %d had %d controls", tc.level, len(w.Controls))
		}
		if err := ValidatedWorkspace(w); err != nil {
			t.Fatal(err)
		}
		r, err := Evaluate(w, t.TempDir(), time.Now(), 90)
		if err != nil {
			t.Fatal(err)
		}
		if r.Summary.Total != tc.n || r.Summary.Unassessed != tc.n || r.Summary.Documented != 0 || r.Certification != "NOT_DETERMINED" {
			t.Fatalf("unsafe status: %+v", r.Summary)
		}
		w.Controls[0].ID = "fabricated"
		if err := ValidatedWorkspace(w); err == nil {
			t.Fatal("fabricated control should fail")
		}
	}
}
func TestEvidenceIntegrityStaleAndContradictions(t *testing.T) {
	root := t.TempDir()
	at := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
	if err := os.WriteFile(filepath.Join(root, "proof.txt"), []byte("access rule approved"), 0600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte("access rule approved"))
	w, _ := NewWorkspace(2, "system-A")
	w.Controls[0].Claim = "IMPLEMENTED"
	w.Controls[0].Reviewer = "assessor-local-01"
	w.Controls[0].Evidence = []Evidence{{Path: "proof.txt", SHA256: hex.EncodeToString(digest[:]), CapturedAt: at.Format(time.RFC3339)}}
	r, err := Evaluate(w, root, at, 90)
	if err != nil {
		t.Fatal(err)
	}
	if r.Summary.Documented != 1 || r.Summary.Unassessed != 109 || r.Certification != "NOT_DETERMINED" {
		t.Fatalf("unexpected summary %+v", r.Summary)
	}
	if err = os.WriteFile(filepath.Join(root, "proof.txt"), []byte("mutated"), 0600); err != nil {
		t.Fatal(err)
	}
	r, err = Evaluate(w, root, at, 90)
	if err != nil {
		t.Fatal(err)
	}
	if r.Summary.Unverifiable != 1 || r.Summary.Documented != 0 {
		t.Fatal("tampering must invalidate evidence")
	}
	w.Controls[0].Evidence[0].CapturedAt = at.Add(-91 * 24 * time.Hour).Format(time.RFC3339)
	r, err = Evaluate(w, root, at, 90)
	if err != nil {
		t.Fatal(err)
	}
	if r.Summary.Unverifiable != 1 {
		t.Fatal("stale must be unverifiable")
	}
	w.Controls[0].Evidence[0].CapturedAt = at.Format(time.RFC3339)
	w.Controls[0].Evidence[0].Path = "../proof.txt"
	r, err = Evaluate(w, root, at, 90)
	if err != nil {
		t.Fatal(err)
	}
	if r.Summary.Unverifiable != 1 {
		t.Fatal("traversal must be unverifiable")
	}
	w.Controls[0].Evidence[0].Path = "proof.txt"
	w.Controls[0].Evidence[0].SHA256 = strings.Repeat("A", 64)
	r, err = Evaluate(w, root, at, 90)
	if err != nil {
		t.Fatal(err)
	}
	if r.Summary.Unverifiable != 1 {
		t.Fatal("noncanonical digest must be unverifiable")
	}
}
func TestSymlinkAndExclusion(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret"), []byte("CUI"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "secret"), filepath.Join(root, "link")); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	if _, err := resolve(root, "link"); err == nil {
		t.Fatal("symlink must be denied")
	}
	w, _ := NewWorkspace(2, "secure-enclave")
	w.Controls[2].Claim = "NOT_APPLICABLE"
	if err := ValidatedWorkspace(w); err == nil {
		t.Fatal("unreviewed exception accepted")
	}
	w.Controls[2].Reviewer = "local-assessor"
	w.Controls[2].Rationale = "system boundary excludes the specified asset, subject to assessment"
	r, err := Evaluate(w, root, time.Now(), 90)
	if err != nil {
		t.Fatal(err)
	}
	if r.Summary.ScopeReview != 1 || r.Summary.Documented != 0 {
		t.Fatal("NA must not become green")
	}
}
func TestMissingEvidenceAndReview(t *testing.T) {
	w, _ := NewWorkspace(2, "secure-enclave")
	w.Controls[0].Claim = "IMPLEMENTED"
	if err := ValidatedWorkspace(w); err == nil {
		t.Fatal("implemented without reviewer must fail")
	}
	w.Controls[0].Reviewer = "unverified-human-claim"
	r, err := Evaluate(w, t.TempDir(), time.Now(), 90)
	if err != nil {
		t.Fatal(err)
	}
	if r.Summary.Unverifiable != 1 {
		t.Fatal("empty evidence cannot prove implementation")
	}
	w.Controls[0].Claim = "GAP"
	r, err = Evaluate(w, t.TempDir(), time.Now(), 90)
	if err != nil {
		t.Fatal(err)
	}
	if r.Summary.Gaps != 1 {
		t.Fatal("gap dropped")
	}
}
