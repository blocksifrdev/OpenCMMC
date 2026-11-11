// Package githubscan verifies selected GitHub repository metadata with read-only GET requests.
// It does not claim a full control, framework, or audit determination.
package githubscan

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
	"regexp"
	"strings"
	"time"

	"github.com/blocksifrdev/OpenCMMC/internal/scan"
)

const maxBody = 2 << 20

var namePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	Token      string
}
type repository struct {
	FullName      string `json:"full_name"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
	Archived      bool   `json:"archived"`
	Disabled      bool   `json:"disabled"`
}
type branch struct {
	Name      string `json:"name"`
	Protected bool   `json:"protected"`
}

// ParseRepo accepts owner/repo, not URLs or alternate HTTP hosts.
func ParseRepo(s string) (string, string, error) {
	p := strings.Split(s, "/")
	if len(p) != 2 || !namePattern.MatchString(p[0]) || !namePattern.MatchString(p[1]) || p[0] == "." || p[1] == "." || p[0] == ".." || p[1] == ".." {
		return "", "", errors.New("repository must be owner/name (not a URL)")
	}
	return p[0], p[1], nil
}

func (c Client) fetch(ctx context.Context, path string, dest any) (scan.Evidence, error) {
	origin := c.BaseURL
	if origin == "" {
		origin = "https://api.github.com"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(origin, "/")+path, nil)
	if err != nil {
		return scan.Evidence{}, errors.New("invalid GitHub API endpoint")
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "BlockSiFr-OpenCMMC/0.2")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	hc := c.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 12 * time.Second}
	}
	resp, err := hc.Do(req)
	if err != nil {
		return scan.Evidence{}, errors.New("GitHub API request failed; verify network availability")
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized, http.StatusForbidden:
		return scan.Evidence{}, fmt.Errorf("GitHub API denied read access (HTTP %d); use a read-only token if needed", resp.StatusCode)
	case http.StatusNotFound:
		return scan.Evidence{}, errors.New("GitHub resource not found or not accessible")
	default:
		return scan.Evidence{}, fmt.Errorf("GitHub metadata unavailable (HTTP %d)", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil || len(data) > maxBody {
		return scan.Evidence{}, errors.New("GitHub response too large or unreadable")
	}
	if json.Unmarshal(data, dest) != nil {
		return scan.Evidence{}, errors.New("GitHub metadata malformed")
	}
	digest := sha256.Sum256(data)
	return scan.Evidence{Path: "github-api:" + path, SHA256: hex.EncodeToString(digest[:]), Bytes: int64(len(data))}, nil
}

func finding(id, title, claim, status, observation, next string, ev ...scan.Evidence) scan.Finding {
	return scan.Finding{ID: id, Title: title, Claim: claim, Status: status, Observation: observation, NextAction: next, Evidence: ev}
}

// Assess reads GitHub's repository metadata and default branch protected flag.
func (c Client) Assess(ctx context.Context, name string) (scan.Report, error) {
	owner, repo, err := ParseRepo(name)
	if err != nil {
		return scan.Report{}, err
	}
	path := "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(repo)
	var meta repository
	ev, err := c.fetch(ctx, path, &meta)
	if err != nil {
		return scan.Report{}, err
	}
	if meta.FullName == "" || meta.DefaultBranch == "" || !strings.EqualFold(meta.FullName, name) {
		return scan.Report{}, errors.New("GitHub response missing expected repository identity or default branch")
	}
	var br branch
	branchEv, branchErr := c.fetch(ctx, path+"/branches/"+url.PathEscape(meta.DefaultBranch), &br)
	r := scan.Report{SchemaVersion: "1.0", Tool: "BlockSiFr OpenCMMC", GeneratedAt: time.Now().UTC().Format(time.RFC3339), Scope: "GITHUB_REPOSITORY:" + meta.FullName, ScopeNotice: "This read-only scan examines repository metadata and a default branch protected flag. It does not establish required reviews, CI checks, bypass restrictions, security effectiveness or framework compliance.", FrameworkNotice: "API response hashes are integrity references, not signed audit attestations. Framework coverage is not determined."}
	visibility := "public"
	if meta.Private {
		visibility = "private"
	}
	r.Findings = append(r.Findings, finding("GITHUB_REPOSITORY_METADATA", "Repository metadata observable", "Read-only metadata was retrieved", scan.Verified, fmt.Sprintf("GitHub reports %s repository %s; default branch %s.", visibility, meta.FullName, meta.DefaultBranch), "Preserve access-controlled response evidence for auditors.", ev))
	if meta.Archived || meta.Disabled {
		r.Findings = append(r.Findings, finding("GITHUB_REPO_ACTIVE", "Repository active status", "Repository not archived or disabled", scan.Gap, "GitHub reports repository archived or disabled.", "Review repository lifecycle status.", ev))
	} else {
		r.Findings = append(r.Findings, finding("GITHUB_REPO_ACTIVE", "Repository active status", "Repository not archived or disabled", scan.Verified, "GitHub reports archived=false and disabled=false.", "Verify organizational ownership separately.", ev))
	}
	switch {
	case branchErr != nil:
		r.Findings = append(r.Findings, finding("GITHUB_BRANCH_PROTECTED", "Default branch protected flag", "Default branch reports protected=true", scan.AttestationNeeded, "Default branch metadata could not be read; status unknown.", "Review permissions and rerun."))
	case br.Name != meta.DefaultBranch:
		r.Findings = append(r.Findings, finding("GITHUB_BRANCH_PROTECTED", "Default branch protected flag", "Default branch reports protected=true", scan.AttestationNeeded, "Branch metadata did not match the default branch.", "Verify branch identity."))
	case br.Protected:
		r.Findings = append(r.Findings, finding("GITHUB_BRANCH_PROTECTED", "Default branch protected flag", "Default branch reports protected=true", scan.Verified, "GitHub reports protected=true; enforcement details are unverified.", "Independently inspect reviews, required checks and bypasses.", branchEv))
	default:
		r.Findings = append(r.Findings, finding("GITHUB_BRANCH_PROTECTED", "Default branch protected flag", "Default branch reports protected=true", scan.Gap, "GitHub reports protected=false for its default branch.", "Enable and test branch protection or rulesets.", branchEv))
	}
	r.Findings = append(r.Findings,
		finding("GITHUB_REQUIRED_REVIEWS", "Required review enforcement", "Pull request reviews required", scan.AttestationNeeded, "Branch summary API does not establish review enforcement.", "Inspect repository ruleset details and test enforcement."),
		finding("GITHUB_REQUIRED_CHECKS", "Required CI enforcement", "Passing checks required", scan.AttestationNeeded, "Branch summary API does not establish status check enforcement.", "Inspect branch ruleset details and test enforcement."))
	for _, f := range r.Findings {
		r.Summary.Total++
		switch f.Status {
		case scan.Verified:
			r.Summary.Verified++
		case scan.Partial:
			r.Summary.Partial++
		case scan.AttestationNeeded:
			r.Summary.AttestationNeeded++
		case scan.Missing:
			r.Summary.Missing++
		case scan.Gap:
			r.Summary.Gap++
		case scan.Inherited:
			r.Summary.Inherited++
		}
	}
	return r, nil
}
