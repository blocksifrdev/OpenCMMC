package githubscan

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/blocksifrdev/OpenCMMC/internal/scan"
)

func fake(t *testing.T, branchResponse string, branchCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Error("not read-only")
			w.WriteHeader(405)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-read-only" {
			t.Error("missing test token")
		}
		switch r.URL.Path {
		case "/repos/acme/demo":
			fmt.Fprint(w, `{"full_name":"acme/demo","default_branch":"main","private":true,"archived":false,"disabled":false}`)
		case "/repos/acme/demo/branches/main":
			w.WriteHeader(branchCode)
			fmt.Fprint(w, branchResponse)
		default:
			w.WriteHeader(404)
		}
	}))
}
func find(r scan.Report, id string) string {
	for _, f := range r.Findings {
		if f.ID == id {
			return f.Status
		}
	}
	return "missing"
}
func TestProtected(t *testing.T) {
	ts := fake(t, `{"name":"main","protected":true}`, 200)
	defer ts.Close()
	r, err := (Client{BaseURL: ts.URL, Token: "test-read-only"}).Assess(context.Background(), "acme/demo")
	if err != nil {
		t.Fatal(err)
	}
	if find(r, "GITHUB_BRANCH_PROTECTED") != scan.Verified {
		t.Fatal("protected branch not found")
	}
	if find(r, "GITHUB_REQUIRED_REVIEWS") != scan.AttestationNeeded {
		t.Fatal("invented required reviews")
	}
	if r.Summary.Total != 5 || r.Summary.Verified != 3 {
		t.Fatalf("wrong counters: %+v", r.Summary)
	}
	if r.Findings[0].Evidence[0].SHA256 == "" {
		t.Fatal("metadata digest missing")
	}
}
func TestUnprotected(t *testing.T) {
	ts := fake(t, `{"name":"main","protected":false}`, 200)
	defer ts.Close()
	r, err := (Client{BaseURL: ts.URL, Token: "test-read-only"}).Assess(context.Background(), "acme/demo")
	if err != nil {
		t.Fatal(err)
	}
	if find(r, "GITHUB_BRANCH_PROTECTED") != scan.Gap {
		t.Fatal("branch vulnerability not identified")
	}
}
func TestBranchUnavailableDoesNotInventGap(t *testing.T) {
	ts := fake(t, `{"message":"Forbidden"}`, 403)
	defer ts.Close()
	r, err := (Client{BaseURL: ts.URL, Token: "test-read-only"}).Assess(context.Background(), "acme/demo")
	if err != nil {
		t.Fatal(err)
	}
	if find(r, "GITHUB_BRANCH_PROTECTED") != scan.AttestationNeeded {
		t.Fatal("unknown became passing or failing")
	}
}
func TestUnreachableRepoFailsClosed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(404) }))
	defer ts.Close()
	_, err := (Client{BaseURL: ts.URL}).Assess(context.Background(), "acme/demo")
	if err == nil {
		t.Fatal("missing repo passed")
	}
}
func TestRejectInvalidNames(t *testing.T) {
	for _, s := range []string{"", "https://github.com/acme/demo", "acme/demo/foo", "a/..", "../demo", "acme/te st"} {
		if _, _, err := ParseRepo(s); err == nil {
			t.Errorf("accepted %q", s)
		}
	}
}
func TestDoesNotLeakTokenOnError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(401) }))
	defer ts.Close()
	_, err := (Client{BaseURL: ts.URL, Token: "SECRET-TOKEN"}).Assess(context.Background(), "acme/demo")
	if err == nil || strings.Contains(err.Error(), "SECRET-TOKEN") {
		t.Fatal("token leak or missing error")
	}
}
