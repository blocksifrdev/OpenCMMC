package cmmc

import (
	"strings"
	"testing"
)

func allMetInput() DraftScoreInput {
	w, _ := NewWorkspace(2, "CUI enclave")
	input := DraftScoreInput{Baseline: w.Baseline, Scope: w.Scope, SSPReference: "approved-ssp.pdf", SSPConfirmedBy: "authorized reviewer"}
	for _, c := range w.Controls {
		input.Adjudications = append(input.Adjudications, Adjudication{ID: c.ID, Finding: "MET", Reviewer: "assessor", Rationale: "user supplied synthetic fixture"})
	}
	return input
}
func setFinding(x *DraftScoreInput, id, finding, partial string) {
	for i := range x.Adjudications {
		if x.Adjudications[i].ID == id {
			x.Adjudications[i].Finding = finding
			x.Adjudications[i].PartialCase = partial
			return
		}
	}
}
func TestDraftScoreNormalAndException(t *testing.T) {
	x := allMetInput()
	o, e := CalculateDraftScore(x)
	if e != nil || o.CalculatedDraftScore != 110 {
		t.Fatalf("all met %v %+v", e, o)
	}
	setFinding(&x, "3.1.1", "NOT_MET", "")
	setFinding(&x, "3.5.3", "PARTIAL", "REMOTE_AND_PRIVILEGED_MFA")
	setFinding(&x, "3.13.11", "PARTIAL", "ENCRYPTED_NOT_FIPS_VALIDATED")
	o, e = CalculateDraftScore(x)
	if e != nil || o.CalculatedDraftScore != 99 || o.AssessmentStatus != "DRAFT_USER_SUPPLIED_NOT_VALIDATED" {
		t.Fatalf("score %v %+v", e, o)
	}
	setFinding(&x, "3.1.1", "MET", "")
	setFinding(&x, "3.3.2", "NOT_MET", "")
	o, e = CalculateDraftScore(x)
	if e != nil || o.CalculatedDraftScore != 101 {
		t.Fatalf("score 3-point %v %+v", e, o)
	}
}
func TestDraftScoreRejectsIncompleteAndInvalid(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*DraftScoreInput)
	}{
		{"missing requirement", func(x *DraftScoreInput) { x.Adjudications = x.Adjudications[1:] }},
		{"empty reviewer", func(x *DraftScoreInput) { x.Adjudications[0].Reviewer = "" }},
		{"duplicate", func(x *DraftScoreInput) { x.Adjudications[1].ID = x.Adjudications[0].ID }},
		{"missing SSP", func(x *DraftScoreInput) { x.SSPReference = "" }},
		{"SSP not met", func(x *DraftScoreInput) { setFinding(x, "3.12.4", "NOT_MET", "") }},
		{"illegal partial", func(x *DraftScoreInput) { setFinding(x, "3.1.1", "PARTIAL", "REMOTE_AND_PRIVILEGED_MFA") }},
		{"invalid partial mode", func(x *DraftScoreInput) { setFinding(x, "3.5.3", "PARTIAL", "ANYTHING") }},
		{"unknown finding", func(x *DraftScoreInput) { setFinding(x, "3.1.1", "UNASSESSED", "") }},
		{"tampered baseline", func(x *DraftScoreInput) { x.Baseline = "NIST rev3" }},
		{"missing scope", func(x *DraftScoreInput) { x.Scope = "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := allMetInput()
			tt.mutate(&x)
			_, err := CalculateDraftScore(x)
			if err == nil {
				t.Fatal("accepted invalid draft score")
			}
		})
	}
	x := allMetInput()
	setFinding(&x, "3.13.11", "NOT_MET", "")
	o, e := CalculateDraftScore(x)
	if e != nil || o.CalculatedDraftScore != 105 {
		t.Fatalf("exception score: %v %+v", e, o)
	}
	if strings.Contains(o.AssessmentStatus, "VALIDATED") && o.AssessmentStatus != "DRAFT_USER_SUPPLIED_NOT_VALIDATED" {
		t.Fatal("unexpected validated status")
	}
}
