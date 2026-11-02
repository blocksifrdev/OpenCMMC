package theorem

import (
	"reflect"
	"testing"
)

func request() Request {
	return Request{
		SchemaVersion: Schema, Scope: "synthetic-ci", AsOf: "2026-11-02T12:00:00Z",
		Variables: map[string][]string{"reviews": {"true", "false"}, "checks": {"true", "false"}, "bypass": {"true", "false"}},
		Evidence: []Observation{
			{ID: "e-review", Atom: Atom{Variable: "reviews", Equals: "true"}, SourceRef: "github:review-policy", Status: "VERIFIED", ObservedAt: "2026-11-01T12:00:00Z", ValidUntil: "2026-11-05T12:00:00Z"},
			{ID: "e-check", Atom: Atom{Variable: "checks", Equals: "true"}, SourceRef: "github:checks", Status: "VERIFIED", ObservedAt: "2026-11-01T12:00:00Z", ValidUntil: "2026-11-05T12:00:00Z"},
		},
		Obligations: []Obligation{{ID: "CHANGE_CONTROL", Description: "Reviews and checks required, no bypass", All: []Atom{{Variable: "reviews", Equals: "true"}, {Variable: "checks", Equals: "true"}, {Variable: "bypass", Equals: "false"}}}},
		Candidates:  []Candidate{{ID: "confirm-no-bypass", Atom: Atom{Variable: "bypass", Equals: "false"}, Cost: 2, SourceToVerify: "effective GitHub ruleset"}, {ID: "repeat-review", Atom: Atom{Variable: "reviews", Equals: "true"}, Cost: 1, SourceToVerify: "rule review"}},
	}
}
func TestCounterexampleAndRepair(t *testing.T) {
	r, err := Evaluate(request())
	if err != nil {
		t.Fatal(err)
	}
	o := r.Results[0]
	if o.Status != ModelCounterexample || o.Counterexample["bypass"] != "true" || r.ConsistentStates != 2 {
		t.Fatalf("unexpected result: %+v", r)
	}
	if o.EvidencePlan == nil || !reflect.DeepEqual(o.EvidencePlan.CandidateIDs, []string{"confirm-no-bypass"}) || o.EvidencePlan.TotalCost != 2 || !o.EvidencePlan.HypotheticalOnly {
		t.Fatalf("bad bounded repair: %+v", o.EvidencePlan)
	}
}
func TestSupportedWithAdmittedEvidence(t *testing.T) {
	r := request()
	r.Evidence = append(r.Evidence, Observation{ID: "e-bypass", Atom: Atom{Variable: "bypass", Equals: "false"}, SourceRef: "github:ruleset", Status: "VERIFIED", ObservedAt: "2026-11-01T12:00:00Z", ValidUntil: "2026-11-05T12:00:00Z"})
	o, err := Evaluate(r)
	if err != nil || o.Results[0].Status != ModelSupported || o.ConsistentStates != 1 {
		t.Fatalf("unexpected: %+v %v", o, err)
	}
}
func TestContradictionNeverVacuous(t *testing.T) {
	r := request()
	r.Evidence = append(r.Evidence, Observation{ID: "contradiction", Atom: Atom{Variable: "reviews", Equals: "false"}, SourceRef: "another:read", Status: "VERIFIED", ObservedAt: "2026-11-01T12:00:00Z", ValidUntil: "2026-11-05T12:00:00Z"})
	o, err := Evaluate(r)
	if err != nil || o.Results[0].Status != Unknown || o.ConsistentStates != 0 {
		t.Fatalf("unexpected: %+v %v", o, err)
	}
}
func TestExpiredFutureUnreviewedNotAdmitted(t *testing.T) {
	r := request()
	r.Evidence[0].ValidUntil = "2026-11-02T11:00:00Z"
	r.Evidence[1].Status = "UNREVIEWED"
	o, err := Evaluate(r)
	if err != nil || o.AdmittedEvidence != 0 || len(o.IgnoredEvidence) != 2 || o.ConsistentStates != 8 {
		t.Fatalf("unexpected: %+v %v", o, err)
	}
}
func TestInvalidDomainOrExplosionRejected(t *testing.T) {
	r := request()
	r.Obligations[0].All[0].Equals = "impossible"
	if _, err := Evaluate(r); err == nil {
		t.Fatal("invalid value accepted")
	}
	r = request()
	r.Variables = map[string][]string{}
	for i := 0; i < 13; i++ {
		r.Variables[string(rune('a'+i))] = []string{"true", "false"}
	}
	if _, err := Evaluate(r); err == nil {
		t.Fatal("too many variables accepted")
	}
}
func TestNoRepairWithContradictoryCandidates(t *testing.T) {
	r := request()
	r.Candidates = []Candidate{{ID: "not-enough", Atom: Atom{Variable: "checks", Equals: "true"}, Cost: 1, SourceToVerify: "tests"}}
	o, err := Evaluate(r)
	if err != nil || o.Results[0].EvidencePlan != nil {
		t.Fatalf("unexpected plan: %+v %v", o, err)
	}
}
func TestDeterministic(t *testing.T) {
	a, _ := Evaluate(request())
	b, _ := Evaluate(request())
	if !reflect.DeepEqual(a, b) {
		t.Fatal("nondeterministic evaluation")
	}
}
func TestVerifiedWithoutSourceIgnored(t *testing.T) {
	r := request()
	r.Evidence[0].SourceRef = ""
	o, err := Evaluate(r)
	if err != nil || o.AdmittedEvidence != 1 || o.IgnoredEvidence[0].Reason != "missing source reference" {
		t.Fatalf("unexpected %+v %v", o, err)
	}
}
func TestCandidateLimit(t *testing.T) {
	r := request()
	for i := 0; i < 13; i++ {
		r.Candidates = append(r.Candidates, Candidate{ID: string(rune('a' + i)), Atom: Atom{Variable: "reviews", Equals: "true"}, Cost: 1, SourceToVerify: "x"})
	}
	if _, err := Evaluate(r); err == nil {
		t.Fatal("excess candidates accepted")
	}
}
