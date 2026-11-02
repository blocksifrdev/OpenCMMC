package cmmc

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"strings"
	"testing"
)

func officialShapeCSV() []byte {
	w, _ := NewObjectiveWorkspace(2, "fixture")
	var b strings.Builder
	writer := csv.NewWriter(&b)
	writer.Write([]string{"Family", "Identifier", "Sort-As", "Security Requirement", "Assessment Objective", "Potential Assessment Method and Objects: Examine", "Potential Assessment Method and Objects: Interview", "Potential Assessment Method and Objects: Test"})
	for _, c := range w.Controls {
		objective := "Determine If:"
		if c.ID == "3.4.4" {
			objective = "Determine if the security impact of changes is reviewed prior to implementation."
		}
		if c.ID == "3.1.22" {
			objective = "Determine if the external system use policy is defined."
		}
		writer.Write([]string{c.ID, c.ID, "", "synthetic requirement statement", objective, "policy", "interview", "test"})
		if c.ID != "3.4.4" {
			writer.Write([]string{c.ID, c.ID + "[a]", "", "", "synthetic determination for " + c.ID, "", "", ""})
		}
	}
	writer.Flush()
	return []byte(b.String())
}
func hashCSV(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func TestImportCompleteOfficialLayoutSynthetic(t *testing.T) {
	b := officialShapeCSV()
	w, s, e := Import171ACSV(b, "approved enclave", hashCSV(b))
	if e != nil {
		t.Fatal(e)
	}
	if s.Requirements != 110 || s.Objectives != 110 || len(w.Controls) != 110 || s.IndependentPDFReconciliation != "NOT_PERFORMED" {
		t.Fatalf("bad coverage: %+v", s)
	}
	for _, ctrl := range w.Controls {
		if ctrl.ID == "3.4.4" && (len(ctrl.Objectives) != 1 || ctrl.Objectives[0].ID != "3.4.4") {
			t.Fatalf("lost official unlettered 3.4.4 objective %+v", ctrl.Objectives)
		}
		if ctrl.ID == "3.1.22" && (len(ctrl.Objectives) != 1 || ctrl.Objectives[0].ID != "3.1.22[a]") {
			t.Fatalf("header wording duplicated lettered objective %+v", ctrl.Objectives)
		}
	}
	if w.Controls[0].Objectives[0].ID != "3.1.1[a]" || w.Controls[0].Methods.Examine != "policy" || w.Controls[0].RequirementText == "" {
		t.Fatalf("lost source metadata %+v", w.Controls[0])
	}
}
func TestImportRejectsTamperingOmissionDuplicate(t *testing.T) {
	b := officialShapeCSV()
	tests := []struct {
		name    string
		content []byte
		hash    string
	}{
		{"checksum mismatch", b, strings.Repeat("0", 64)},
		{"missing objective", []byte(strings.Replace(string(b), "3.1.1,3.1.1[a],,", "3.1.1,3.1.1[z],,", 1)), ""},
		{"missing requirement", []byte(strings.Replace(string(b), "3.1.1,3.1.1,,synthetic requirement statement,Determine If:,policy,interview,test\n", "", 1)), ""},
		{"duplicate objective", append(append([]byte(nil), b...), []byte("3.1.1,3.1.1[a],,,duplicate,,,\n")...), ""},
		{"bad header", []byte("name,word\n3.1.1,foo\n"), ""},
	}
	for _, x := range tests {
		t.Run(x.name, func(t *testing.T) {
			h := x.hash
			if h == "" {
				h = hashCSV(x.content)
			}
			_, _, e := Import171ACSV(x.content, "enclave", h)
			if e == nil {
				t.Fatal("accepted invalid")
			}
		})
	}
}
func TestImportRejectsMissingCrossControlAndPartial(t *testing.T) {
	b := officialShapeCSV()
	for _, pattern := range []string{"3.1.1,3.1.1[a],,,synthetic determination for 3.1.1,,,", "3.1.1,3.1.1,,synthetic requirement statement,Determine If:,policy,interview,test"} {
		bad := []byte(strings.Replace(string(b), pattern, "", 1))
		_, _, e := Import171ACSV(bad, "enclave", hashCSV(bad))
		if e == nil {
			t.Fatal("accepted incomplete CSV")
		}
	}
}
