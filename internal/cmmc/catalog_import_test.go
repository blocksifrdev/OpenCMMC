package cmmc

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"strings"
	"testing"
)

func testCSV() []byte {
	w, _ := NewObjectiveWorkspace(2, "test enclave")
	var b strings.Builder
	writer := csv.NewWriter(&b)
	writer.Write([]string{"Security Requirement", "Assessment Objective", "Potential Assessment Methods and Objects"})
	for _, c := range w.Controls {
		writer.Write([]string{c.ID, "[a] test determination for " + c.ID, "EXAMINE; INTERVIEW; TEST"})
	}
	writer.Flush()
	return []byte(b.String())
}
func hashCSV(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func TestImport171CompleteSynthetic(t *testing.T) {
	b := testCSV()
	w, s, e := Import171ACSV(b, "approved enclave", hashCSV(b))
	if e != nil {
		t.Fatal(e)
	}
	if s.Requirements != 110 || s.Objectives != 110 || len(w.Controls) != 110 || s.IndependentPDFReconciliation != "NOT_PERFORMED" {
		t.Fatalf("bad coverage: %+v", s)
	}
	if w.Controls[0].Objectives[0].ID != "3.1.1[a]" || !strings.Contains(w.Controls[0].Objectives[0].Statement, "test determination") {
		t.Fatalf("lost objective data: %+v", w.Controls[0].Objectives)
	}
}
func TestImportRejectsIncompleteOrCorrupt(t *testing.T) {
	b := testCSV()
	tests := []struct {
		name string
		csv  []byte
		hash string
	}{
		{"wrong checksum", b, strings.Repeat("0", 64)},
		{"incomplete", []byte(strings.Replace(string(b), "3.1.1,[a] test determination for 3.1.1,", "", 1)), ""},
		{"wrong control", []byte(strings.Replace(string(b), "3.1.1,[a] test determination for 3.1.1,", "3.1.1,3.2.2[a] improper mapping,", 1)), ""},
		{"repeated objective", append(append([]byte(nil), b...), []byte("3.1.1,[a] duplicate,EXAMINE\n")...), ""},
		{"malformed header", []byte("name,text\n3.1.1,[a] foo\n"), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.hash
			if h == "" {
				h = hashCSV(tt.csv)
			}
			_, _, e := Import171ACSV(tt.csv, "enclave", h)
			if e == nil {
				t.Fatal("unexpected acceptance")
			}
		})
	}
}
func TestImportRejectsUnknownOrMixedObjectives(t *testing.T) {
	b := testCSV()
	bad := strings.Replace(string(b), "3.1.1,[a] test determination for 3.1.1,", "3.1.1,3.1.2[a] mismatched control,", 1)
	_, _, e := Import171ACSV([]byte(bad), "test", hashCSV([]byte(bad)))
	if e == nil {
		t.Fatal("must reject mismatched ID")
	}
}
