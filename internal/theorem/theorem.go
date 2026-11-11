// Package theorem evaluates bounded, explicitly modeled compliance obligations.
// MODEL_SUPPORTED is a statement about the enumerated model, not real-world control
// effectiveness, an audit opinion, an authorization grant, or certification.
package theorem

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

const (
	Schema              = "opencmmc.theorem.v1"
	ModelSupported      = "MODEL_SUPPORTED"
	ModelCounterexample = "MODEL_COUNTEREXAMPLE"
	Unknown             = "UNKNOWN"
	maxStates           = 4096
	maxCandidates       = 12
)

type Atom struct {
	Variable string `json:"variable"`
	Equals   string `json:"equals"`
}
type Observation struct {
	ID string `json:"id"`
	Atom
	SourceRef  string `json:"source_ref"`
	Status     string `json:"status"`
	ObservedAt string `json:"observed_at"`
	ValidUntil string `json:"valid_until"`
}
type Candidate struct {
	ID string `json:"id"`
	Atom
	Cost           int    `json:"cost"`
	SourceToVerify string `json:"source_to_verify"`
}
type Obligation struct {
	ID           string `json:"id"`
	Description  string `json:"description"`
	FrameworkRef string `json:"framework_ref,omitempty"`
	All          []Atom `json:"all"`
}
type Request struct {
	SchemaVersion string              `json:"schema_version"`
	Scope         string              `json:"scope"`
	AsOf          string              `json:"as_of"`
	Variables     map[string][]string `json:"variables"`
	Evidence      []Observation       `json:"evidence"`
	Obligations   []Obligation        `json:"obligations"`
	Candidates    []Candidate         `json:"candidate_evidence,omitempty"`
}
type IgnoredEvidence struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}
type EvidencePlan struct {
	CandidateIDs     []string `json:"candidate_ids"`
	TotalCost        int      `json:"total_cost"`
	HypotheticalOnly bool     `json:"hypothetical_only"`
	BoundedExact     bool     `json:"bounded_exact"`
}
type ClaimResult struct {
	ID                   string            `json:"id"`
	Description          string            `json:"description"`
	FrameworkRef         string            `json:"framework_ref,omitempty"`
	Status               string            `json:"status"`
	SatisfyingStates     int               `json:"satisfying_states"`
	CounterexampleStates int               `json:"counterexample_states"`
	Counterexample       map[string]string `json:"modeled_counterexample,omitempty"`
	EvidencePlan         *EvidencePlan     `json:"evidence_plan,omitempty"`
	Explanation          string            `json:"explanation"`
}
type Report struct {
	SchemaVersion    string            `json:"schema_version"`
	Scope            string            `json:"scope"`
	AsOf             string            `json:"as_of"`
	ModelSHA256      string            `json:"model_sha256"`
	PossibleStates   int               `json:"possible_states"`
	ConsistentStates int               `json:"consistent_states"`
	AdmittedEvidence int               `json:"admitted_evidence"`
	IgnoredEvidence  []IgnoredEvidence `json:"ignored_evidence"`
	Results          []ClaimResult     `json:"results"`
	Disclaimer       string            `json:"disclaimer"`
}

type state map[string]string

func Validate(r Request) error {
	if !compatibleTheoremSchema(r.SchemaVersion) {
		return fmt.Errorf("schema_version must be %q", Schema)
	}
	if strings.TrimSpace(r.Scope) == "" || len(r.Scope) > 256 {
		return errors.New("scope must be a nonempty bounded string")
	}
	if _, err := time.Parse(time.RFC3339, r.AsOf); err != nil {
		return fmt.Errorf("as_of: %w", err)
	}
	if len(r.Variables) == 0 || len(r.Variables) > 12 {
		return errors.New("variables must contain 1 to 12 entries")
	}
	product := 1
	for k, v := range r.Variables {
		if strings.TrimSpace(k) == "" || len(k) > 96 || len(v) < 2 || len(v) > 8 {
			return fmt.Errorf("variable %q requires 2 to 8 domain values", k)
		}
		seen := map[string]bool{}
		for _, val := range v {
			if strings.TrimSpace(val) == "" || len(val) > 128 || seen[val] {
				return fmt.Errorf("variable %q contains empty, oversized or duplicate values", k)
			}
			seen[val] = true
		}
		if product > maxStates/len(v) {
			return fmt.Errorf("finite model exceeds %d states", maxStates)
		}
		product *= len(v)
	}
	checkAtom := func(a Atom) error {
		vals, ok := r.Variables[a.Variable]
		if !ok {
			return fmt.Errorf("unknown variable %q", a.Variable)
		}
		for _, v := range vals {
			if v == a.Equals {
				return nil
			}
		}
		return fmt.Errorf("value %q is not in domain of %q", a.Equals, a.Variable)
	}
	if len(r.Obligations) == 0 || len(r.Obligations) > 100 {
		return errors.New("1 to 100 obligations required")
	}
	ids := map[string]bool{}
	for _, o := range r.Obligations {
		if o.ID == "" || len(o.ID) > 128 || ids[o.ID] || len(o.All) == 0 {
			return fmt.Errorf("obligation IDs must be unique and conjunctions nonempty: %q", o.ID)
		}
		ids[o.ID] = true
		for _, a := range o.All {
			if err := checkAtom(a); err != nil {
				return fmt.Errorf("obligation %s: %w", o.ID, err)
			}
		}
	}
	seenEvidence := map[string]bool{}
	for _, e := range r.Evidence {
		if e.ID == "" || seenEvidence[e.ID] || len(e.ID) > 128 {
			return fmt.Errorf("duplicate, empty or oversized evidence ID %q", e.ID)
		}
		seenEvidence[e.ID] = true
		if err := checkAtom(e.Atom); err != nil {
			return fmt.Errorf("evidence %s: %w", e.ID, err)
		}
		if e.Status != "VERIFIED" && e.Status != "UNREVIEWED" && e.Status != "REVOKED" {
			return fmt.Errorf("evidence %s: unsupported status", e.ID)
		}
		if len(e.SourceRef) > 1024 {
			return fmt.Errorf("evidence %s: source_ref too long", e.ID)
		}
	}
	if len(r.Candidates) > maxCandidates {
		return fmt.Errorf("candidate pool exceeds %d", maxCandidates)
	}
	candidateIDs := map[string]bool{}
	for _, c := range r.Candidates {
		if c.ID == "" || len(c.ID) > 128 || candidateIDs[c.ID] || seenEvidence[c.ID] {
			return fmt.Errorf("duplicate or empty candidate ID %q", c.ID)
		}
		candidateIDs[c.ID] = true
		if err := checkAtom(c.Atom); err != nil {
			return fmt.Errorf("candidate %s: %w", c.ID, err)
		}
		if c.Cost < 1 || c.Cost > 1000000 || strings.TrimSpace(c.SourceToVerify) == "" || len(c.SourceToVerify) > 1024 {
			return fmt.Errorf("candidate %s needs cost 1..1000000 and source_to_verify", c.ID)
		}
	}
	return nil
}

func holds(s state, a Atom) bool { return s[a.Variable] == a.Equals }
func all(s state, clauses []Atom) bool {
	for _, a := range clauses {
		if !holds(s, a) {
			return false
		}
	}
	return true
}
func clone(s state) map[string]string {
	c := make(map[string]string, len(s))
	for k, v := range s {
		c[k] = v
	}
	return c
}
func allStates(r Request) []state {
	keys := make([]string, 0, len(r.Variables))
	for k := range r.Variables {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	states := []state{{}}
	for _, k := range keys {
		values := append([]string(nil), r.Variables[k]...)
		sort.Strings(values)
		next := make([]state, 0, len(states)*len(values))
		for _, s := range states {
			for _, v := range values {
				n := state(clone(s))
				n[k] = v
				next = append(next, n)
			}
		}
		states = next
	}
	return states
}
func isAdmitted(e Observation, at time.Time) (bool, string) {
	if e.Status != "VERIFIED" {
		return false, "not verified (caller-provided label)"
	}
	if strings.TrimSpace(e.SourceRef) == "" {
		return false, "missing source reference"
	}
	from, err := time.Parse(time.RFC3339, e.ObservedAt)
	if err != nil {
		return false, "invalid observed_at"
	}
	until, err := time.Parse(time.RFC3339, e.ValidUntil)
	if err != nil {
		return false, "invalid valid_until"
	}
	if from.After(at) {
		return false, "future-dated observation"
	}
	if !until.After(at) {
		return false, "expired observation"
	}
	if !until.After(from) {
		return false, "invalid validity window"
	}
	return true, ""
}

// Evaluate is deterministic for a fixed input; hashes are integrity identifiers, not signatures.
func Evaluate(r Request) (Report, error) {
	if err := Validate(r); err != nil {
		return Report{}, err
	}
	raw, err := json.Marshal(r)
	if err != nil {
		return Report{}, err
	}
	hash := sha256.Sum256(raw)
	out := Report{SchemaVersion: Schema, Scope: r.Scope, AsOf: r.AsOf, ModelSHA256: hex.EncodeToString(hash[:]), IgnoredEvidence: []IgnoredEvidence{}, Results: []ClaimResult{}, Disclaimer: "Findings are conditional on a caller-supplied finite model. VERIFIED evidence labels, sources, and model completeness are not independently authenticated. A modeled counterexample is not an observed breach. An evidence plan is hypothetical. No control certification, audit opinion, real-world safety proof, or runtime authorization is implied."}
	universe := allStates(r)
	out.PossibleStates = len(universe)
	at, _ := time.Parse(time.RFC3339, r.AsOf)
	admitted := []Atom{}
	for _, e := range r.Evidence {
		ok, why := isAdmitted(e, at)
		if ok {
			admitted = append(admitted, e.Atom)
			out.AdmittedEvidence++
		} else {
			out.IgnoredEvidence = append(out.IgnoredEvidence, IgnoredEvidence{ID: e.ID, Reason: why})
		}
	}
	consistent := make([]state, 0, len(universe))
	for _, s := range universe {
		if all(s, admitted) {
			consistent = append(consistent, s)
		}
	}
	out.ConsistentStates = len(consistent)
	for _, o := range r.Obligations {
		res := ClaimResult{ID: o.ID, Description: o.Description, FrameworkRef: o.FrameworkRef, Status: Unknown}
		if len(consistent) == 0 {
			res.Explanation = "No state satisfies all admitted observations; contradictory evidence or model assumptions require review. Empty worlds never count as proof."
		} else {
			for _, s := range consistent {
				if all(s, o.All) {
					res.SatisfyingStates++
				} else {
					res.CounterexampleStates++
					if res.Counterexample == nil {
						res.Counterexample = clone(s)
					}
				}
			}
			if res.CounterexampleStates == 0 {
				res.Status = ModelSupported
				res.Explanation = "Every evidence-consistent state in the explicitly bounded model satisfies this conjunction; real-world coverage is unverified."
			} else {
				res.Status = ModelCounterexample
				res.Explanation = "At least one evidence-consistent modeled state violates the claim. This is a counterexample to model-wide entailment, not evidence of an actual incident."
				res.EvidencePlan = findPlan(r.Candidates, consistent, o.All)
			}
		}
		out.Results = append(out.Results, res)
	}
	return out, nil
}

// findPlan computes an exact minimum-cost subset within the declared finite candidate pool.
// Its candidate observations are hypothetical and must never be promoted automatically.
func findPlan(candidates []Candidate, states []state, claim []Atom) *EvidencePlan {
	n := len(candidates)
	if n == 0 {
		return nil
	}
	bestCost := int(^uint(0) >> 1)
	var best []string
	for mask := 1; mask < (1 << n); mask++ {
		cost := 0
		ids := []string{}
		atoms := []Atom{}
		for j, c := range candidates {
			if mask&(1<<j) != 0 {
				cost += c.Cost
				ids = append(ids, c.ID)
				atoms = append(atoms, c.Atom)
			}
		}
		if cost > bestCost {
			continue
		}
		sort.Strings(ids)
		count := 0
		valid := true
		for _, s := range states {
			if all(s, atoms) {
				count++
				if !all(s, claim) {
					valid = false
					break
				}
			}
		}
		if !valid || count == 0 {
			continue
		}
		key := strings.Join(ids, "\x00")
		if cost < bestCost || (cost == bestCost && (best == nil || key < strings.Join(best, "\x00"))) {
			bestCost = cost
			best = ids
		}
	}
	if best == nil {
		return nil
	}
	return &EvidencePlan{CandidateIDs: best, TotalCost: bestCost, HypotheticalOnly: true, BoundedExact: true}
}
