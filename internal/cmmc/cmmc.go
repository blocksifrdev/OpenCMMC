// Package cmmc supplies an entirely offline, version-pinned CMMC evidence workbench.
// Linked evidence is never represented as independently proven control effectiveness.
package cmmc

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	Schema                 = "freecompliance.cmmc.workspace.v1"
	ReportSchema           = "freecompliance.cmmc.readiness.v1"
	Level1                 = 1
	Level2                 = 2
	MaxEvidenceBytes int64 = 50 << 20
)

// Family describes the 14 NIST SP 800-171 Revision 2 requirement families.
// The counts represent 110 requirements. The authoritative source is the NIST PDF,
// not this machine-readable index of identifiers.
type Family struct {
	Number int
	Count  int
	Name   string
}

var families = []Family{
	{1, 22, "Access Control"}, {2, 3, "Awareness and Training"},
	{3, 9, "Audit and Accountability"}, {4, 9, "Configuration Management"},
	{5, 11, "Identification and Authentication"}, {6, 3, "Incident Response"},
	{7, 6, "Maintenance"}, {8, 9, "Media Protection"},
	{9, 2, "Personnel Security"}, {10, 6, "Physical Protection"},
	{11, 3, "Risk Assessment"}, {12, 4, "Security Assessment"},
	{13, 16, "System and Communications Protection"}, {14, 7, "System and Information Integrity"},
}
var romans = []string{"i", "ii", "iii", "iv", "v", "vi", "vii", "viii", "ix", "x", "xi", "xii", "xiii", "xiv", "xv"}
var hexDigest = regexp.MustCompile(`^[0-9a-f]{64}$`)

type Control struct {
	ID        string     `json:"id"`
	Family    string     `json:"family"`
	Claim     string     `json:"claim"` // UNASSESSED, IMPLEMENTED, GAP, NOT_APPLICABLE
	Reviewer  string     `json:"reviewer,omitempty"`
	Rationale string     `json:"rationale,omitempty"`
	Evidence  []Evidence `json:"evidence,omitempty"`
}
type Evidence struct {
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	CapturedAt string `json:"captured_at"`
}
type Workspace struct {
	SchemaVersion         string    `json:"schema_version"`
	Baseline              string    `json:"baseline"`
	Level                 int       `json:"level"`
	Scope                 string    `json:"scope"`
	SystemSecurityPlanRef string    `json:"system_security_plan_ref,omitempty"`
	Controls              []Control `json:"controls"`
}
type Finding struct {
	ID             string          `json:"id"`
	Status         string          `json:"status"`
	EvidenceChecks []EvidenceCheck `json:"evidence_checks,omitempty"`
	NextAction     string          `json:"next_action"`
}
type EvidenceCheck struct {
	Path   string `json:"path"`
	Status string `json:"status"`
	SHA256 string `json:"sha256,omitempty"`
}
type Summary struct {
	Total        int `json:"total"`
	Documented   int `json:"documented"`
	Gaps         int `json:"gaps"`
	Unassessed   int `json:"unassessed"`
	Unverifiable int `json:"unverifiable"`
	ScopeReview  int `json:"scope_review"`
}
type Report struct {
	SchemaVersion      string    `json:"schema_version"`
	Baseline           string    `json:"baseline"`
	Level              int       `json:"level"`
	Scope              string    `json:"scope"`
	AsOf               string    `json:"as_of"`
	WorkspaceSHA256    string    `json:"workspace_sha256"`
	Summary            Summary   `json:"summary"`
	Findings           []Finding `json:"findings"`
	Certification      string    `json:"certification"`
	OfficialSubmission string    `json:"official_submission"`
	Notice             string    `json:"notice"`
}

func baseline(level int) (string, error) {
	switch level {
	case 1:
		return "FAR 52.204-21(b)(1)(i)-(xv) / CMMC Level 1", nil
	case 2:
		return "NIST SP 800-171 Rev. 2 (2021 update) / CMMC Level 2", nil
	default:
		return "", fmt.Errorf("supported levels are 1 or 2, got %d", level)
	}
}
func NewWorkspace(level int, scope string) (Workspace, error) {
	b, err := baseline(level)
	if err != nil {
		return Workspace{}, err
	}
	if strings.TrimSpace(scope) == "" || len(scope) > 256 {
		return Workspace{}, errors.New("scope must be 1 to 256 characters")
	}
	w := Workspace{SchemaVersion: Schema, Baseline: b, Level: level, Scope: scope}
	if level == 1 {
		for _, r := range romans {
			w.Controls = append(w.Controls, Control{ID: "FAR-52.204-21-b.1." + r, Family: "Basic Safeguarding of Covered Contractor Information Systems", Claim: "UNASSESSED"})
		}
	} else {
		for _, f := range families {
			for i := 1; i <= f.Count; i++ {
				w.Controls = append(w.Controls, Control{ID: fmt.Sprintf("3.%d.%d", f.Number, i), Family: f.Name, Claim: "UNASSESSED"})
			}
		}
	}
	return w, nil
}
func ValidatedWorkspace(w Workspace) error {
	expected, err := NewWorkspace(w.Level, w.Scope)
	if err != nil {
		return err
	}
	if w.SchemaVersion != Schema || w.Baseline != expected.Baseline {
		return errors.New("unsupported or changed workspace baseline/schema")
	}
	if len(w.Controls) != len(expected.Controls) {
		return fmt.Errorf("expected exactly %d controls", len(expected.Controls))
	}
	ids := make(map[string]Control, len(w.Controls))
	for _, c := range w.Controls {
		if _, found := ids[c.ID]; found {
			return fmt.Errorf("duplicate control %q", c.ID)
		}
		if len(c.Reviewer) > 256 || len(c.Rationale) > 2048 || len(c.Evidence) > 16 {
			return fmt.Errorf("oversized control metadata for %q", c.ID)
		}
		switch c.Claim {
		case "UNASSESSED", "IMPLEMENTED", "GAP", "NOT_APPLICABLE":
		default:
			return fmt.Errorf("invalid claim for %q: %q", c.ID, c.Claim)
		}
		if c.Claim == "NOT_APPLICABLE" && (strings.TrimSpace(c.Reviewer) == "" || strings.TrimSpace(c.Rationale) == "") {
			return fmt.Errorf("not-applicable claim %q requires reviewer and scoped rationale", c.ID)
		}
		if c.Claim == "IMPLEMENTED" && strings.TrimSpace(c.Reviewer) == "" {
			return fmt.Errorf("implemented claim %q requires reviewer identification", c.ID)
		}
		ids[c.ID] = c
	}
	for _, c := range expected.Controls {
		v, ok := ids[c.ID]
		if !ok || v.Family != c.Family {
			return fmt.Errorf("missing or altered control %q", c.ID)
		}
	}
	if len(w.SystemSecurityPlanRef) > 1024 {
		return errors.New("system_security_plan_ref too long")
	}
	return nil
}

// Resolve avoids symlink and traversal escapes. All files and intermediate path
// components must be non-symlink filesystem objects inside evidenceRoot.
func resolve(root, relative string) (string, error) {
	if relative == "" || len(relative) > 1024 || filepath.IsAbs(relative) || strings.Contains(relative, "\\") {
		return "", errors.New("evidence path must be a short relative slash-separated path")
	}
	norm := filepath.Clean(relative)
	if norm == "." || norm == ".." || strings.HasPrefix(norm, ".."+string(filepath.Separator)) {
		return "", errors.New("evidence path escapes root")
	}
	base, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	current := base
	baseInfo, err := os.Lstat(current)
	if err != nil {
		return "", err
	}
	if baseInfo.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("evidence root must not be symlink")
	}
	for _, seg := range strings.Split(filepath.ToSlash(norm), "/") {
		if seg == "" || seg == "." || seg == ".." {
			return "", errors.New("invalid evidence path segment")
		}
		current = filepath.Join(current, seg)
		st, err := os.Lstat(current)
		if err != nil {
			return "", err
		}
		if st.Mode()&os.ModeSymlink != 0 {
			return "", errors.New("symlink evidence path denied")
		}
	}
	return current, nil
}
func digestFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	s, err := f.Stat()
	if err != nil {
		return "", err
	}
	if !s.Mode().IsRegular() || s.Size() > MaxEvidenceBytes {
		return "", errors.New("evidence must be a regular file under 50 MiB")
	}
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, MaxEvidenceBytes+1))
	if err != nil {
		return "", err
	}
	if n > MaxEvidenceBytes {
		return "", errors.New("evidence exceeds 50 MiB")
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func Evaluate(w Workspace, root string, at time.Time, maxAgeDays int) (Report, error) {
	if err := ValidatedWorkspace(w); err != nil {
		return Report{}, err
	}
	if maxAgeDays < 1 || maxAgeDays > 3660 {
		return Report{}, errors.New("max-age-days must be 1 through 3660")
	}
	if root == "" {
		return Report{}, errors.New("evidence directory required")
	}
	j, err := json.Marshal(w)
	if err != nil {
		return Report{}, err
	}
	s := sha256.Sum256(j)
	r := Report{SchemaVersion: ReportSchema, Baseline: w.Baseline, Level: w.Level, Scope: w.Scope, AsOf: at.UTC().Format(time.RFC3339), WorkspaceSHA256: hex.EncodeToString(s[:]), Certification: "NOT_DETERMINED", OfficialSubmission: "NOT_PERFORMED", Notice: "Offline evidence integrity and human assertions only. DOCUMENTED does not prove control effectiveness. No SPRS score, CMMC certification, government acceptance or authoritative OSCAL export is implied."}
	controls := append([]Control(nil), w.Controls...)
	sort.Slice(controls, func(i, j int) bool { return controls[i].ID < controls[j].ID })
	for _, c := range controls {
		f := Finding{ID: c.ID}
		switch c.Claim {
		case "UNASSESSED":
			f.Status = "UNASSESSED"
			f.NextAction = "Assess applicability and collect independently reviewable evidence."
			r.Summary.Unassessed++
		case "GAP":
			f.Status = "GAP"
			f.NextAction = "Create an owner-approved remediation and evidence plan."
			r.Summary.Gaps++
		case "NOT_APPLICABLE":
			f.Status = "SCOPE_REVIEW_REQUIRED"
			f.NextAction = "An authorized assessor must validate the exclusion against the actual system boundary."
			r.Summary.ScopeReview++
		case "IMPLEMENTED":
			good := len(c.Evidence) > 0
			seen := map[string]bool{}
			for _, e := range c.Evidence {
				check := EvidenceCheck{Path: e.Path}
				if seen[e.Path] {
					check.Status = "DUPLICATE_REFERENCE"
					good = false
				} else {
					seen[e.Path] = true
					switch {
					case !hexDigest.MatchString(e.SHA256):
						check.Status = "INVALID_SHA256"
						good = false
					default:
						parsed, er := time.Parse(time.RFC3339, e.CapturedAt)
						if er != nil || parsed.After(at.Add(5*time.Minute)) || at.Sub(parsed) > time.Duration(maxAgeDays)*24*time.Hour {
							check.Status = "STALE_OR_INVALID_DATE"
							good = false
						} else {
							path, er := resolve(root, e.Path)
							if er != nil {
								check.Status = "MISSING_OR_UNSAFE_FILE"
								good = false
							} else {
								hash, er := digestFile(path)
								if er != nil {
									check.Status = "UNREADABLE_OR_OVERSIZED"
									good = false
								} else if hash != strings.ToLower(e.SHA256) {
									check.Status = "DIGEST_MISMATCH"
									good = false
								} else {
									check.Status = "INTEGRITY_MATCH"
									check.SHA256 = hash
								}
							}
						}
					}
				}
				f.EvidenceChecks = append(f.EvidenceChecks, check)
			}
			if good {
				f.Status = "DOCUMENTED"
				f.NextAction = "Independently evaluate the technical requirement, evidence quality, operation period, and scope; do not infer compliance from hashes."
				r.Summary.Documented++
			} else {
				f.Status = "UNVERIFIABLE"
				f.NextAction = "Resolve missing, expired or altered evidence, then obtain independent technical assessment."
				r.Summary.Unverifiable++
			}
		}
		r.Findings = append(r.Findings, f)
	}
	r.Summary.Total = len(r.Findings)
	return r, nil
}
