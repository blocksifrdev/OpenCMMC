// Package sourceproof binds bounded theorem evaluation to read-only GitHub
// active-ruleset metadata. It does not validate branch protection settings,
// ruleset bypass rights, deployment enforcement, or audit compliance.
package sourceproof

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/blocksifrdev/freecompliance/internal/githubscan"
	"github.com/blocksifrdev/freecompliance/internal/theorem"
)

const maxResponse = 2 << 20
const maxPages = 5
const maxSourceAge = 5 * time.Minute

// Client does not permit production callers to override an API host. BaseURL
// and Clock are provided solely for deterministic local test fixtures.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	Token      string
	Clock      func() time.Time
}

type Source struct {
	Endpoint    string `json:"endpoint"`
	SHA256      string `json:"sha256"`
	Bytes       int    `json:"bytes"`
	RetrievedAt string `json:"retrieved_at"`
	SourceDate  string `json:"source_date,omitempty"`
	ValidUntil  string `json:"valid_until,omitempty"`
	Fresh       bool   `json:"fresh"`
}

type Result struct {
	SchemaVersion  string         `json:"schema_version"`
	Repository     string         `json:"repository"`
	DefaultBranch  string         `json:"default_branch,omitempty"`
	Sources        []Source       `json:"sources"`
	SourceComplete bool           `json:"source_complete"`
	Warnings       []string       `json:"warnings"`
	Proof          theorem.Report `json:"proof"`
	Limitations    string         `json:"limitations"`
}

type metadata struct {
	FullName      string `json:"full_name"`
	DefaultBranch string `json:"default_branch"`
}
type rule struct {
	Type       string          `json:"type"`
	Parameters json.RawMessage `json:"parameters"`
}

func (c Client) now() time.Time {
	if c.Clock != nil {
		return c.Clock().UTC()
	}
	return time.Now().UTC()
}

// get only requests official fixed-path read-only endpoints, rejects redirects,
// and verifies freshness based on the server's HTTP Date and Age headers.
func (c Client) get(ctx context.Context, path string, dest any) (Source, error) {
	host := c.BaseURL
	if host == "" {
		host = "https://api.github.com"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(host, "/")+path, nil)
	if err != nil {
		return Source{}, errors.New("invalid request")
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("User-Agent", "FreeCompliance-sourceproof/0.4")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	safe := *client
	safe.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := safe.Do(req)
	if err != nil {
		return Source{}, errors.New("GitHub GET failed (network or timeout)")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Source{}, fmt.Errorf("GitHub GET unavailable: HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponse+1))
	if err != nil || len(data) > maxResponse {
		return Source{}, errors.New("GitHub response unreadable or over limit")
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return Source{}, errors.New("GitHub response is not valid JSON for expected endpoint")
	}
	now := c.now()
	digest := sha256.Sum256(data)
	out := Source{Endpoint: "github-api:" + path, SHA256: hex.EncodeToString(digest[:]), Bytes: len(data), RetrievedAt: now.Format(time.RFC3339)}
	dated, err := http.ParseTime(resp.Header.Get("Date"))
	if err != nil {
		return out, nil
	}
	out.SourceDate = dated.UTC().Format(time.RFC3339)
	age := time.Duration(0)
	if raw := resp.Header.Get("Age"); raw != "" {
		seconds, e := strconv.ParseInt(raw, 10, 64)
		if e != nil || seconds < 0 || seconds > 86400 {
			return out, nil
		}
		age = time.Duration(seconds) * time.Second
	}
	skew := now.Sub(dated)
	if skew < -30*time.Second || skew > maxSourceAge || age > maxSourceAge {
		return out, nil
	}
	out.Fresh = true
	out.ValidUntil = now.Add(maxSourceAge).Format(time.RFC3339)
	return out, nil
}

func prRequired(r rule) bool {
	if r.Type != "pull_request" {
		return false
	}
	var p struct {
		RequiredApprovingReviewCount *int `json:"required_approving_review_count"`
	}
	if json.Unmarshal(r.Parameters, &p) != nil || p.RequiredApprovingReviewCount == nil {
		return false
	}
	return *p.RequiredApprovingReviewCount > 0
}
func checksRequired(r rule) bool {
	if r.Type != "required_status_checks" {
		return false
	}
	var p struct {
		RequiredStatusChecks []struct {
			Context string `json:"context"`
		} `json:"required_status_checks"`
	}
	if json.Unmarshal(r.Parameters, &p) != nil {
		return false
	}
	for _, check := range p.RequiredStatusChecks {
		if strings.TrimSpace(check.Context) != "" {
			return true
		}
	}
	return false
}

// Assess collects evidence from a GitHub repository and evaluates two narrowly
// defined claims: active ruleset declarations, and whether those declarations
// suffice for a no-bypass change-control claim. MODEL_SUPPORTED is bounded.
func (c Client) Assess(ctx context.Context, name string) (Result, error) {
	owner, repo, err := githubscan.ParseRepo(name)
	if err != nil {
		return Result{}, err
	}
	now := c.now()
	out := Result{SchemaVersion: "freecompliance.sourceproof.v1", Repository: name, Sources: []Source{}, Warnings: []string{}, Limitations: "Evaluates active GitHub ruleset declarations only, not classic branch protection, bypass actor rights, actual GitHub enforcement or production deployment. MODEL_SUPPORTED is model-relative and does not certify an organization."}
	prefix := "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(repo)
	var meta metadata
	source, err := c.get(ctx, prefix, &meta)
	if err != nil {
		return Result{}, err
	}
	out.Sources = append(out.Sources, source)
	if !strings.EqualFold(meta.FullName, name) || meta.DefaultBranch == "" || len(meta.DefaultBranch) > 255 || strings.Contains(meta.DefaultBranch, "?") {
		return Result{}, errors.New("GitHub identity or branch mismatch")
	}
	out.DefaultBranch = meta.DefaultBranch
	sourceOK := source.Fresh
	if !source.Fresh {
		out.Warnings = append(out.Warnings, "GitHub repository metadata lacks fresh Date/Age provenance")
	}
	reviewFound, checksFound := false, false
	complete := false
	for page := 1; page <= maxPages; page++ {
		path := prefix + "/rules/branches/" + url.PathEscape(meta.DefaultBranch) + "?per_page=100&page=" + strconv.Itoa(page)
		var rules []rule
		ev, e := c.get(ctx, path, &rules)
		if e != nil {
			out.Warnings = append(out.Warnings, "Active-ruleset endpoint unavailable; source completeness UNKNOWN: "+e.Error())
			sourceOK = false
			break
		}
		out.Sources = append(out.Sources, ev)
		if !ev.Fresh {
			out.Warnings = append(out.Warnings, "Rule response is stale or lacks trusted HTTP freshness metadata")
			sourceOK = false
		}
		for _, r := range rules {
			if prRequired(r) {
				reviewFound = true
			}
			if checksRequired(r) {
				checksFound = true
			}
		}
		if len(rules) < 100 {
			complete = true
			break
		}
	}
	if !complete {
		out.Warnings = append(out.Warnings, "Active-ruleset pagination not proven complete")
	}
	out.SourceComplete = complete && sourceOK
	evidence := []theorem.Observation{}
	// These are evidence of *rule declarations*, not proof of enforcement.
	// Only admit results from complete fresh snapshots, never partial pages.
	if out.SourceComplete {
		ref := ""
		for _, s := range out.Sources {
			ref += s.Endpoint + "#sha256=" + s.SHA256 + ";"
		}
		emit := func(id, variable string) {
			evidence = append(evidence, theorem.Observation{ID: id, Atom: theorem.Atom{Variable: variable, Equals: "present"}, SourceRef: ref, Status: "VERIFIED", ObservedAt: now.Format(time.RFC3339), ValidUntil: now.Add(maxSourceAge).Format(time.RFC3339)})
		}
		if reviewFound {
			emit("github-active-review-rule", "reviews")
		}
		if checksFound {
			emit("github-active-check-rule", "checks")
		}
	}
	request := theorem.Request{SchemaVersion: theorem.Schema, Scope: "GITHUB_RULESET_DECLARATIONS:" + meta.FullName + ":" + meta.DefaultBranch, AsOf: now.Format(time.RFC3339), Variables: map[string][]string{"reviews": {"present", "unproven"}, "checks": {"present", "unproven"}, "bypass": {"possible", "none"}}, Evidence: evidence, Obligations: []theorem.Obligation{
		{ID: "RULESET_REVIEW_AND_CHECK_DECLARATIONS", Description: "The active ruleset API contains a pull-request rule with at least one required approval and a status-check rule with at least one context (does not establish effectiveness)", All: []theorem.Atom{{Variable: "reviews", Equals: "present"}, {Variable: "checks", Equals: "present"}}},
		{ID: "CHANGE_CONTROL_WITH_NO_BYPASS", Description: "Required review and CI declarations and absence of all bypass paths (bypass information is intentionally unverified)", All: []theorem.Atom{{Variable: "reviews", Equals: "present"}, {Variable: "checks", Equals: "present"}, {Variable: "bypass", Equals: "none"}}},
	}, Candidates: []theorem.Candidate{{ID: "verify-review-policy", Atom: theorem.Atom{Variable: "reviews", Equals: "present"}, Cost: 2, SourceToVerify: "Confirm effective review rules and permitted paths"}, {ID: "verify-check-policy", Atom: theorem.Atom{Variable: "checks", Equals: "present"}, Cost: 2, SourceToVerify: "Confirm effective check rules and contexts"}, {ID: "verify-all-bypass-paths", Atom: theorem.Atom{Variable: "bypass", Equals: "none"}, Cost: 4, SourceToVerify: "Authorized ruleset/branch protection and deployment-path investigation"}}}
	proof, err := theorem.Evaluate(request)
	if err != nil {
		return Result{}, err
	}
	if !out.SourceComplete {
		for i := range proof.Results {
			proof.Results[i].Status = theorem.Unknown
			proof.Results[i].Explanation = "Source unavailable, stale, or incompletely paginated. Modeled worlds cannot be presented as source-bound evidence."
			proof.Results[i].Counterexample = nil
			proof.Results[i].EvidencePlan = nil
		}
	}
	out.Proof = proof
	if out.SourceComplete && !reviewFound {
		out.Warnings = append(out.Warnings, "No active ruleset with a positive required approval count was observed; classic branch protection not examined")
	}
	if out.SourceComplete && !checksFound {
		out.Warnings = append(out.Warnings, "No active ruleset with required status check contexts was observed; classic branch protection not examined")
	}
	out.Warnings = append(out.Warnings, "Read-only metadata cannot prove absence of bypass actors or successful enforcement")
	return out, nil
}