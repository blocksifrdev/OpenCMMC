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

func baseObjectiveFixture(t *testing.T) (ObjectiveWorkspace, string, time.Time) {
	t.Helper()
	w, err := NewObjectiveWorkspace(2, "Demo enclave")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "review.txt"), []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256([]byte("fixture"))
	now := time.Date(2026, 11, 2, 14, 0, 0, 0, time.UTC)
	w.Controls[0].Objectives = []ObjectiveItem{{ID: w.Controls[0].ID + "[a]", Decision: "ASSERTED_MET", Method: "EXAMINE", Reviewer: "test reviewer", Rationale: "local mock evidence only", Evidence: []ObjectiveEvidence{{Path: "review.txt", SHA256: hex.EncodeToString(h[:]), CapturedAt: now.Add(-time.Hour).Format(time.RFC3339)}}}}
	return w, root, now
}
func TestObjectiveWorkspaceHasNoInventedObjectives(t *testing.T) {
	w, err := NewObjectiveWorkspace(2, "Demo")
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Controls) != 110 {
		t.Fatalf("expected 110 got %d", len(w.Controls))
	}
	for _, c := range w.Controls {
		if len(c.Objectives) != 0 {
			t.Fatal("invented official objective")
		}
	}
}
func TestObjectiveReportNeverMarksControlsCertified(t *testing.T) {
	w, root, now := baseObjectiveFixture(t)
	r, err := AssessObjectives(w, root, now, 90)
	if err != nil {
		t.Fatal(err)
	}
	if r.Summary.TotalControls != 110 || r.Summary.MissingObjectiveLists != 109 || r.Summary.ReviewReady != 1 {
		t.Fatalf("unexpected result: %+v", r.Summary)
	}
	if r.CatalogCompleteness != "NOT_INDEPENDENTLY_VALIDATED" || r.SPRSScore != "NOT_CALCULATED" || r.ControlEffectiveness != "NOT_INDEPENDENTLY_VERIFIED" {
		t.Fatal("unsupported compliance claim")
	}
}
func TestObjectiveEvidenceFailuresAreUnknown(t *testing.T) {
	checks := map[string]func(*ObjectiveWorkspace, string, time.Time){
		"mismatch": func(w *ObjectiveWorkspace, _ string, _ time.Time) {
			w.Controls[0].Objectives[0].Evidence[0].SHA256 = strings.Repeat("a", 64)
		},
		"expired": func(w *ObjectiveWorkspace, _ string, now time.Time) {
			w.Controls[0].Objectives[0].Evidence[0].CapturedAt = now.Add(-91 * 24 * time.Hour).Format(time.RFC3339)
		},
		"future": func(w *ObjectiveWorkspace, _ string, now time.Time) {
			w.Controls[0].Objectives[0].Evidence[0].CapturedAt = now.Add(time.Hour).Format(time.RFC3339)
		},
		"traversal": func(w *ObjectiveWorkspace, _ string, _ time.Time) {
			w.Controls[0].Objectives[0].Evidence[0].Path = "../review.txt"
		},
		"missing": func(w *ObjectiveWorkspace, _ string, _ time.Time) {
			w.Controls[0].Objectives[0].Evidence[0].Path = "missing.txt"
		},
		"none": func(w *ObjectiveWorkspace, _ string, _ time.Time) { w.Controls[0].Objectives[0].Evidence = nil },
	}
	for name, edit := range checks {
		t.Run(name, func(t *testing.T) {
			w, root, now := baseObjectiveFixture(t)
			edit(&w, root, now)
			r, err := AssessObjectives(w, root, now, 90)
			if err != nil {
				t.Fatal(err)
			}
			if r.Summary.ReviewReady != 0 || r.Summary.Unresolved != 110 {
				t.Fatalf("false ready for %s: %+v", name, r.Summary)
			}
		})
	}
}
func TestObjectiveRejectsManipulatedCatalog(t *testing.T) {
	for name, edit := range map[string]func(*ObjectiveWorkspace){
		"duplicate-controls":      func(w *ObjectiveWorkspace) { w.Controls[1].ID = w.Controls[0].ID },
		"missing-controls":        func(w *ObjectiveWorkspace) { w.Controls = w.Controls[1:] },
		"cross-control-objective": func(w *ObjectiveWorkspace) { w.Controls[0].Objectives[0].ID = w.Controls[1].ID + "[a]" },
		"duplicate-objectives": func(w *ObjectiveWorkspace) {
			w.Controls[0].Objectives = append(w.Controls[0].Objectives, w.Controls[0].Objectives[0])
		},
		"untrusted-baseline": func(w *ObjectiveWorkspace) { w.Baseline = "unrelated" },
		"bad-source-hash":    func(w *ObjectiveWorkspace) { w.ProcedureSHA256 = "BAD" },
		"missing-reviewer":   func(w *ObjectiveWorkspace) { w.Controls[0].Objectives[0].Reviewer = "" },
		"missing-method":     func(w *ObjectiveWorkspace) { w.Controls[0].Objectives[0].Method = "" },
		"invalid-claim":      func(w *ObjectiveWorkspace) { w.Controls[0].Objectives[0].Decision = "VERIFIED" },
	} {
		t.Run(name, func(t *testing.T) {
			w, _, _ := baseObjectiveFixture(t)
			edit(&w)
			if err := ValidateObjectiveWorkspace(w); err == nil {
				t.Fatal("accepted untrusted input")
			}
		})
	}
}
func TestObjectiveSymlinkAndContradiction(t *testing.T) {
	w, root, now := baseObjectiveFixture(t)
	p := filepath.Join(root, "link.txt")
	if err := os.Symlink("review.txt", p); err != nil {
		t.Skip("symlinks not supported")
	}
	w.Controls[0].Objectives[0].Evidence[0].Path = "link.txt"
	r, err := AssessObjectives(w, root, now, 90)
	if err != nil {
		t.Fatal(err)
	}
	if r.Summary.ReviewReady != 0 {
		t.Fatal("symlink promoted to review ready")
	}
	w.Controls[0].Objectives[0].Decision = "ASSERTED_NOT_MET"
	w.Controls[0].Objectives[0].Evidence = nil
	r, err = AssessObjectives(w, root, now, 90)
	if err != nil {
		t.Fatal(err)
	}
	if r.Summary.GapsReported != 1 || r.SPRSScore != "NOT_CALCULATED" {
		t.Fatal("reported gap was not preserved")
	}
}
func TestObjectiveNoNetworkAndLevel1(t *testing.T) {
	w, err := NewObjectiveWorkspace(1, "FCI system")
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Controls) != 15 {
		t.Fatal("bad Level 1 count")
	}
	if err := ValidateObjectiveWorkspace(w); err != nil {
		t.Fatal(err)
	}
	if _, err = NewObjectiveWorkspace(3, "bad"); err == nil {
		t.Fatal("accepted unsupported level")
	}
}