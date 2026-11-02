package sourceproof

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/blocksifrdev/freecompliance/internal/theorem"
)

var fixed = time.Date(2026, 11, 2, 16, 0, 0, 0, time.UTC)

func fixture(t *testing.T, rules string, mode string) (Client, *int, func()) {
	t.Helper()
	requests := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodGet {
			t.Errorf("write API call: %s", r.Method)
		}
		if !strings.HasPrefix(r.URL.Path, "/repos/example/project") {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Date", fixed.Format(http.TimeFormat))
		if mode == "stale" {
			w.Header().Set("Age", "800")
		}
		if strings.HasSuffix(r.URL.Path, "/project") {
			fmt.Fprint(w, `{"full_name":"example/project","default_branch":"main"}`)
			return
		}
		if mode == "denied" {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"message":"no permissions"}`)
			return
		}
		if mode == "badjson" {
			fmt.Fprint(w, `{bad`)
			return
		}
		if mode == "endless" {
			a := make([]string, 100)
			for i := range a {
				a[i] = `{"type":"creation"}`
			}
			fmt.Fprintf(w, "[%s]", strings.Join(a, ","))
			return
		}
		fmt.Fprint(w, rules)
	})
	srv := httptest.NewServer(handler)
	return Client{BaseURL: srv.URL, Clock: func() time.Time { return fixed }}, &requests, srv.Close
}

const positive = `[{"type":"pull_request","parameters":{"required_approving_review_count":1}},{"type":"required_status_checks","parameters":{"required_status_checks":[{"context":"go-test"}]}}]`

func TestSourceBoundStatementsAndNoBypassGuarantee(t *testing.T) {
	c, requests, close := fixture(t, positive, "")
	defer close()
	r, err := c.Assess(context.Background(), "example/project")
	if err != nil {
		t.Fatal(err)
	}
	if !r.SourceComplete || *requests != 2 || len(r.Sources) != 2 || !r.Sources[0].Fresh {
		t.Fatalf("bad source snapshot: %+v", r)
	}
	if r.Proof.AdmittedEvidence != 2 || r.Proof.Results[0].Status != theorem.ModelSupported {
		t.Fatalf("expected bounded declarations supported: %+v", r.Proof)
	}
	if r.Proof.Results[1].Status != theorem.ModelCounterexample || r.Proof.Results[1].Counterexample["bypass"] != "possible" {
		t.Fatalf("bypass must be unproven: %+v", r.Proof.Results[1])
	}
	if r.Proof.Results[1].EvidencePlan == nil || !r.Proof.Results[1].EvidencePlan.HypotheticalOnly {
		t.Fatal("expected bounded hypothetical next-evidence plan")
	}
	if len(r.Sources[1].SHA256) != 64 || strings.Contains(r.Proof.ModelSHA256, "token") {
		t.Fatal("provenance hash absent")
	}
}
func TestNoFalseGreenWithIncompleteOrStaleSources(t *testing.T) {
	for _, mode := range []string{"denied", "stale", "badjson", "endless"} {
		t.Run(mode, func(t *testing.T) {
			c, _, close := fixture(t, positive, mode)
			defer close()
			r, err := c.Assess(context.Background(), "example/project")
			if err != nil {
				t.Fatal(err)
			}
			if r.SourceComplete || len(r.Warnings) == 0 {
				t.Fatalf("source incorrectly complete: %+v", r)
			}
			for _, o := range r.Proof.Results {
				if o.Status != theorem.Unknown || o.EvidencePlan != nil || o.Counterexample != nil {
					t.Fatalf("false proof on %s: %+v", mode, o)
				}
			}
			if r.Proof.AdmittedEvidence != 0 {
				t.Fatalf("stale evidence admitted in %s", mode)
			}
		})
	}
}
func TestAbsentRuleIsNotEstablishedCompliance(t *testing.T) {
	c, _, close := fixture(t, `[]`, "")
	defer close()
	r, err := c.Assess(context.Background(), "example/project")
	if err != nil {
		t.Fatal(err)
	}
	if !r.SourceComplete || r.Proof.AdmittedEvidence != 0 {
		t.Fatalf("unexpected: %+v", r)
	}
	for _, o := range r.Proof.Results {
		if o.Status == theorem.ModelSupported {
			t.Fatalf("false green: %+v", o)
		}
	}
}
func TestNoReviewIfZeroApprovalsAndNoCheckIfNoContext(t *testing.T) {
	data := `[{"type":"pull_request","parameters":{"required_approving_review_count":0}},{"type":"required_status_checks","parameters":{"required_status_checks":[]}}]`
	c, _, close := fixture(t, data, "")
	defer close()
	r, err := c.Assess(context.Background(), "example/project")
	if err != nil {
		t.Fatal(err)
	}
	if r.Proof.AdmittedEvidence != 0 {
		t.Fatal("empty requirements must not be admitted")
	}
}
func TestRedirectNeverForwardsCredential(t *testing.T) {
	hits := 0
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits++ }))
	defer target.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL, http.StatusFound) }))
	defer source.Close()
	c := Client{BaseURL: source.URL, Token: "super-secret-token", Clock: func() time.Time { return fixed }}
	_, err := c.Assess(context.Background(), "example/project")
	if err == nil || hits != 0 {
		t.Fatalf("redirect followed: hits=%d err=%v", hits, err)
	}
	if strings.Contains(err.Error(), "super-secret-token") {
		t.Fatal("token leaked")
	}
}
func TestMalformedRepoRejectedBeforeNetwork(t *testing.T) {
	c := Client{}
	for _, repo := range []string{"http://bad.example", "../project", "example/project/../../foo"} {
		if _, err := c.Assess(context.Background(), repo); err == nil {
			t.Fatalf("accepted %q", repo)
		}
	}
}