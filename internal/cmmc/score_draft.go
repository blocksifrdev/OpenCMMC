package cmmc

import (
	"errors"
	"fmt"
	"strings"
)

// ScoringAuthority describes the publicly published CMMC Level 2 scoring rule.
// Do not confuse this local arithmetic with independent assessment validation.
const ScoringAuthority = "32 CFR 170.24(c)(2), CMMC Level 2 / NIST SP 800-171 Rev 2"

var fivePointIDs = strings.Fields(`3.1.1 3.1.2 3.2.1 3.2.2 3.3.1 3.4.1 3.4.2 3.5.1 3.5.2 3.6.1 3.6.2 3.7.2 3.8.3 3.9.2 3.10.1 3.10.2 3.12.1 3.12.3 3.13.1 3.13.2 3.14.1 3.14.2 3.14.3 3.1.12 3.1.13 3.1.16 3.1.17 3.1.18 3.3.5 3.4.5 3.4.6 3.4.7 3.4.8 3.5.10 3.7.5 3.8.7 3.11.2 3.13.5 3.13.6 3.13.15 3.14.4 3.14.6`)
var threePointIDs = strings.Fields(`3.3.2 3.7.1 3.8.1 3.8.2 3.9.1 3.11.1 3.12.2 3.1.5 3.1.19 3.7.4 3.8.8 3.13.8 3.14.5 3.14.7`)

// Adjudication is a user-entered decision from an authorized human assessor.
// The tool does not establish the authority or truth of that assertion.
type Adjudication struct {
	ID          string `json:"id"`
	Finding     string `json:"finding"` // MET / NOT_MET / NOT_APPLICABLE / PARTIAL
	Reviewer    string `json:"reviewer"`
	Rationale   string `json:"rationale"`
	PartialCase string `json:"partial_case,omitempty"`
}
type DraftScoreInput struct {
	Baseline       string         `json:"baseline"`
	Scope          string         `json:"scope"`
	SSPReference   string         `json:"ssp_reference"`
	SSPConfirmedBy string         `json:"ssp_confirmed_by"`
	Adjudications  []Adjudication `json:"adjudications"`
}
type ScoreLine struct {
	ID                 string `json:"id"`
	Finding            string `json:"finding"`
	PotentialDeduction int    `json:"potential_deduction"`
	Deduction          int    `json:"deduction"`
	Reviewer           string `json:"reviewer"`
}
type DraftScore struct {
	AssessmentStatus     string      `json:"assessment_status"`
	Authority            string      `json:"authority"`
	Baseline             string      `json:"baseline"`
	Scope                string      `json:"scope"`
	Maximum              int         `json:"maximum"`
	CalculatedDraftScore int         `json:"calculated_draft_score"`
	DeductedPoints       int         `json:"deducted_points"`
	Lines                []ScoreLine `json:"lines"`
	OfficialSubmission   string      `json:"official_submission"`
	Notice               string      `json:"notice"`
}

func pointWeights() (map[string]int, error) {
	w, _ := NewWorkspace(2, "weight-check")
	m := make(map[string]int, len(w.Controls))
	for _, c := range w.Controls {
		m[c.ID] = 1
	}
	for _, id := range fivePointIDs {
		if _, ok := m[id]; !ok {
			return nil, fmt.Errorf("unrecognized 5-point requirement %s", id)
		}
		if m[id] != 1 {
			return nil, errors.New("duplicate 5-point ID")
		}
		m[id] = 5
	}
	for _, id := range threePointIDs {
		if _, ok := m[id]; !ok {
			return nil, fmt.Errorf("unrecognized 3-point requirement %s", id)
		}
		if m[id] != 1 {
			return nil, errors.New("duplicated scoring ID")
		}
		m[id] = 3
	}
	for _, id := range []string{"3.5.3", "3.13.11"} {
		if m[id] != 1 {
			return nil, errors.New("overlapping exception weight")
		}
		m[id] = 5
	}
	if len(m) != 110 {
		return nil, errors.New("unexpected scoring catalog coverage")
	}
	return m, nil
}

// CalculateDraftScore is fail-closed: all 110 requirements must have a human
// disposition; the SSP must be confirmed; scores are marked DRAFT only.
// Partial credit is allowed ONLY for the two explicitly enumerated cases in
// 32 CFR 170.24(c)(2)(i)(B)(4), subject to independent assessor verification.
func CalculateDraftScore(in DraftScoreInput) (DraftScore, error) {
	var out DraftScore
	expected, err := NewWorkspace(2, in.Scope)
	if err != nil {
		return out, err
	}
	if in.Baseline != expected.Baseline {
		return out, errors.New("unsupported scoring baseline")
	}
	if strings.TrimSpace(in.SSPReference) == "" || strings.TrimSpace(in.SSPConfirmedBy) == "" {
		return out, errors.New("cannot calculate: current system security plan and reviewer confirmation required")
	}
	weights, err := pointWeights()
	if err != nil {
		return out, err
	}
	if len(in.Adjudications) != 110 {
		return out, fmt.Errorf("requires 110 adjudications, received %d", len(in.Adjudications))
	}
	seen := map[string]bool{}
	out = DraftScore{AssessmentStatus: "DRAFT_USER_SUPPLIED_NOT_VALIDATED", Authority: ScoringAuthority, Baseline: in.Baseline, Scope: in.Scope, Maximum: 110, OfficialSubmission: "NOT_PERFORMED", Notice: "Arithmetic from user-entered decisions only. Does not verify assessment objectives, CUI boundary, reviewer authority, evidence, N/A judgments, SSP currency or official SPRS acceptance. Not certification or a government score."}
	for _, a := range in.Adjudications {
		point, ok := weights[a.ID]
		if !ok || seen[a.ID] {
			return DraftScore{}, fmt.Errorf("unknown or duplicate requirement %q", a.ID)
		}
		seen[a.ID] = true
		if strings.TrimSpace(a.Reviewer) == "" || strings.TrimSpace(a.Rationale) == "" || len(a.Rationale) > 4096 || len(a.Reviewer) > 256 {
			return DraftScore{}, fmt.Errorf("human reviewer and rationale required for %s", a.ID)
		}
		deduct := 0
		switch a.Finding {
		case "MET", "NOT_APPLICABLE":
			if a.PartialCase != "" {
				return DraftScore{}, errors.New("partial case supplied with full result")
			}
		case "NOT_MET":
			if a.PartialCase != "" {
				return DraftScore{}, errors.New("partial case supplied with NOT_MET")
			}
			deduct = point
		case "PARTIAL":
			switch {
			case a.ID == "3.5.3" && a.PartialCase == "REMOTE_AND_PRIVILEGED_MFA":
				deduct = 3
			case a.ID == "3.13.11" && a.PartialCase == "ENCRYPTED_NOT_FIPS_VALIDATED":
				deduct = 3
			default:
				return DraftScore{}, fmt.Errorf("partial credit not permitted for %s", a.ID)
			}
		default:
			return DraftScore{}, fmt.Errorf("missing or unsupported finding for %s", a.ID)
		}
		if a.ID == "3.12.4" && a.Finding != "MET" {
			return DraftScore{}, errors.New("no completed assessment or score without confirmed SSP (3.12.4)")
		}
		out.Lines = append(out.Lines, ScoreLine{ID: a.ID, Finding: a.Finding, PotentialDeduction: point, Deduction: deduct, Reviewer: a.Reviewer})
		out.DeductedPoints += deduct
	}
	out.CalculatedDraftScore = 110 - out.DeductedPoints
	return out, nil
}
