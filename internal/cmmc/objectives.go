package cmmc

// Objective assessment is an offline evidence workbench. An assessor's assertion
// and a matching evidence hash never establish independent control effectiveness.
import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

const ObjectiveSchema = "freecompliance.cmmc.objectives.v1"
const ObjectiveReportSchema = "freecompliance.cmmc.objective-findings.v1"

type ObjectiveEvidence struct {
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	CapturedAt string `json:"captured_at"`
}
type ObjectiveItem struct {
	Statement string              `json:"statement,omitempty"` // imported supplemental objective text; PDF validation pending
	ID        string              `json:"id"`
	Decision  string              `json:"decision"`         // UNASSESSED, ASSERTED_MET, ASSERTED_NOT_MET, NOT_APPLICABLE
	Method    string              `json:"method,omitempty"` // EXAMINE, INTERVIEW, TEST
	Reviewer  string              `json:"reviewer,omitempty"`
	Rationale string              `json:"rationale,omitempty"`
	Evidence  []ObjectiveEvidence `json:"evidence,omitempty"`
}
type ObjectiveControl struct {
	ID         string          `json:"id"`
	Objectives []ObjectiveItem `json:"objectives"`
}
type ObjectiveWorkspace struct {
	SchemaVersion   string             `json:"schema_version"`
	Baseline        string             `json:"baseline"`
	Level           int                `json:"level"`
	Scope           string             `json:"scope"`
	ProcedureSource string             `json:"procedure_source"`
	ProcedureSHA256 string             `json:"procedure_sha256,omitempty"`
	Controls        []ObjectiveControl `json:"controls"`
}
type ObjectiveFinding struct {
	ControlID      string   `json:"control_id"`
	Status         string   `json:"status"`
	ObjectiveCount int      `json:"objective_count"`
	Unresolved     []string `json:"unresolved,omitempty"`
	EvidenceErrors []string `json:"evidence_errors,omitempty"`
}
type ObjectiveSummary struct {
	TotalControls         int `json:"total_controls"`
	MissingObjectiveLists int `json:"missing_objective_lists"`
	ReviewReady           int `json:"review_ready"`
	GapsReported          int `json:"gaps_reported"`
	Unresolved            int `json:"unresolved"`
	ObjectiveCount        int `json:"objective_count"`
	EvidenceChecks        int `json:"evidence_checks"`
}
type ObjectiveReport struct {
	SchemaVersion        string             `json:"schema_version"`
	Baseline             string             `json:"baseline"`
	Scope                string             `json:"scope"`
	AsOf                 string             `json:"as_of"`
	InputSHA256          string             `json:"input_sha256"`
	ProcedureSource      string             `json:"procedure_source"`
	ProcedureSHA256      string             `json:"procedure_sha256,omitempty"`
	CatalogCompleteness  string             `json:"catalog_completeness"`
	ControlEffectiveness string             `json:"control_effectiveness"`
	SPRSScore            string             `json:"sprs_score"`
	GovernmentSubmission string             `json:"government_submission"`
	Summary              ObjectiveSummary   `json:"summary"`
	Findings             []ObjectiveFinding `json:"findings"`
	Notice               string             `json:"notice"`
}

// NewObjectiveWorkspace deliberately does not invent NIST 800-171A objectives.
// Users supply independently reviewed assessment IDs from the pinned authority.
func NewObjectiveWorkspace(level int, scope string) (ObjectiveWorkspace, error) {
	w, err := NewWorkspace(level, scope)
	if err != nil {
		return ObjectiveWorkspace{}, err
	}
	out := ObjectiveWorkspace{SchemaVersion: ObjectiveSchema, Baseline: w.Baseline, Level: level, Scope: scope,
		ProcedureSource: "NIST SP 800-171A (June 2018); verify applicability for current contractual CMMC baseline"}
	for _, c := range w.Controls {
		out.Controls = append(out.Controls, ObjectiveControl{ID: c.ID, Objectives: []ObjectiveItem{}})
	}
	return out, nil
}

func ValidateObjectiveWorkspace(w ObjectiveWorkspace) error {
	exp, err := NewObjectiveWorkspace(w.Level, w.Scope)
	if err != nil {
		return err
	}
	if w.SchemaVersion != ObjectiveSchema || w.Baseline != exp.Baseline {
		return errors.New("unsupported objective schema or baseline")
	}
	if strings.TrimSpace(w.ProcedureSource) == "" || len(w.ProcedureSource) > 1024 {
		return errors.New("procedure source required")
	}
	if w.ProcedureSHA256 != "" && !hexDigest.MatchString(w.ProcedureSHA256) {
		return errors.New("procedure SHA must be lowercase SHA-256")
	}
	if len(w.Controls) != len(exp.Controls) {
		return fmt.Errorf("expected exactly %d controls", len(exp.Controls))
	}
	valid := map[string]bool{}
	for _, c := range exp.Controls {
		valid[c.ID] = true
	}
	seen := map[string]bool{}
	for _, c := range w.Controls {
		if !valid[c.ID] || seen[c.ID] {
			return fmt.Errorf("missing, unknown or duplicate control %q", c.ID)
		}
		seen[c.ID] = true
		if len(c.Objectives) > 64 {
			return fmt.Errorf("too many objectives for %q", c.ID)
		}
		objectIDs := map[string]bool{}
		for _, o := range c.Objectives {
			if len(o.ID) < len(c.ID)+3 || !strings.HasPrefix(o.ID, c.ID+"[") || !strings.HasSuffix(o.ID, "]") || objectIDs[o.ID] || len(o.ID) > 64 {
				return fmt.Errorf("invalid or duplicate objective ID %q", o.ID)
			}
			suffix := o.ID[len(c.ID)+1 : len(o.ID)-1]
			if suffix == "" {
				return fmt.Errorf("empty objective ID suffix %q", o.ID)
			}
			for _, r := range suffix {
				if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '.' {
					return fmt.Errorf("invalid objective ID %q", o.ID)
				}
			}
			objectIDs[o.ID] = true
			switch o.Decision {
			case "UNASSESSED", "ASSERTED_MET", "ASSERTED_NOT_MET", "NOT_APPLICABLE":
			default:
				return fmt.Errorf("unsupported decision for %q", o.ID)
			}
			if len(o.Reviewer) > 256 || len(o.Rationale) > 2048 || len(o.Evidence) > 16 || len(o.Statement) > 4096 {
				return fmt.Errorf("oversized objective metadata %q", o.ID)
			}
			if o.Decision != "UNASSESSED" {
				if strings.TrimSpace(o.Reviewer) == "" || strings.TrimSpace(o.Rationale) == "" {
					return fmt.Errorf("reviewer and rationale required for %q", o.ID)
				}
			}
			if o.Decision == "ASSERTED_MET" {
				if o.Method != "EXAMINE" && o.Method != "INTERVIEW" && o.Method != "TEST" {
					return fmt.Errorf("assessment method required for %q", o.ID)
				}
			}
		}
	}
	return nil
}

func AssessObjectives(w ObjectiveWorkspace, evidenceRoot string, at time.Time, maxAgeDays int) (ObjectiveReport, error) {
	if err := ValidateObjectiveWorkspace(w); err != nil {
		return ObjectiveReport{}, err
	}
	if evidenceRoot == "" || maxAgeDays < 1 || maxAgeDays > 3660 {
		return ObjectiveReport{}, errors.New("evidence-dir required and max-age-days must be 1..3660")
	}
	if at.IsZero() {
		return ObjectiveReport{}, errors.New("assessment timestamp required")
	}
	b, err := json.Marshal(w)
	if err != nil {
		return ObjectiveReport{}, err
	}
	digest := sha256.Sum256(b)
	r := ObjectiveReport{SchemaVersion: ObjectiveReportSchema, Baseline: w.Baseline, Scope: w.Scope, AsOf: at.UTC().Format(time.RFC3339), InputSHA256: hex.EncodeToString(digest[:]), ProcedureSource: w.ProcedureSource, ProcedureSHA256: w.ProcedureSHA256, CatalogCompleteness: "NOT_INDEPENDENTLY_VALIDATED", ControlEffectiveness: "NOT_INDEPENDENTLY_VERIFIED", SPRSScore: "NOT_CALCULATED", GovernmentSubmission: "NOT_PERFORMED", Notice: "Reviewer assertions and matching hashes are not independent proof of CMMC control effectiveness. Catalog objective completeness and applicability have not been verified against the authoritative source. No score or certification."}
	cs := append([]ObjectiveControl(nil), w.Controls...)
	sort.Slice(cs, func(i, j int) bool { return cs[i].ID < cs[j].ID })
	r.Summary.TotalControls = len(cs)
	for _, c := range cs {
		f := ObjectiveFinding{ControlID: c.ID, ObjectiveCount: len(c.Objectives), Status: "UNRESOLVED"}
		r.Summary.ObjectiveCount += len(c.Objectives)
		if len(c.Objectives) == 0 {
			f.Status = "OBJECTIVE_LIST_MISSING"
			r.Summary.MissingObjectiveLists++
			r.Summary.Unresolved++
			r.Findings = append(r.Findings, f)
			continue
		}
		hasGap, hasUnknown := false, false
		for _, o := range c.Objectives {
			switch o.Decision {
			case "ASSERTED_NOT_MET":
				hasGap = true
			case "UNASSESSED", "NOT_APPLICABLE":
				hasUnknown = true
				f.Unresolved = append(f.Unresolved, o.ID)
			case "ASSERTED_MET":
				if len(o.Evidence) == 0 {
					hasUnknown = true
					f.Unresolved = append(f.Unresolved, o.ID+":NO_EVIDENCE")
					continue
				}
				filesSeen := map[string]bool{}
				for _, e := range o.Evidence {
					r.Summary.EvidenceChecks++
					issue := ""
					switch {
					case filesSeen[e.Path]:
						issue = "DUPLICATE_PATH"
					case !hexDigest.MatchString(e.SHA256):
						issue = "INVALID_DIGEST"
					default:
						captured, er := time.Parse(time.RFC3339, e.CapturedAt)
						if er != nil || captured.After(at.Add(5*time.Minute)) || at.Sub(captured) > time.Duration(maxAgeDays)*24*time.Hour {
							issue = "STALE_OR_INVALID_TIME"
						} else {
							path, er := resolve(evidenceRoot, e.Path)
							if er != nil {
								issue = "MISSING_OR_UNSAFE_EVIDENCE"
							} else {
								sha, er := digestFile(path)
								if er != nil {
									issue = "UNREADABLE_EVIDENCE"
								} else if sha != e.SHA256 {
									issue = "HASH_MISMATCH"
								}
							}
						}
					}
					filesSeen[e.Path] = true
					if issue != "" {
						hasUnknown = true
						f.EvidenceErrors = append(f.EvidenceErrors, o.ID+":"+issue)
					}
				}
			}
		}
		if hasGap {
			f.Status = "REVIEWER_REPORTED_GAP"
			r.Summary.GapsReported++
		} else if hasUnknown {
			f.Status = "UNRESOLVED"
			r.Summary.Unresolved++
		} else {
			f.Status = "EVIDENCE_REVIEW_READY"
			r.Summary.ReviewReady++
		}
		r.Findings = append(r.Findings, f)
	}
	return r, nil
}
