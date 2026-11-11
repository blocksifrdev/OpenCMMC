// Package scan produces a local, evidence-linked repository readiness report.
// A verified signal is not a certification or a claim that an entire control is met.
package scan

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	Verified          = "VERIFIED"
	Partial           = "PARTIAL"
	AttestationNeeded = "ATTESTATION_NEEDED"
	Missing           = "MISSING"
	Gap               = "GAP"
	Inherited         = "INHERITED"
	maxEvidenceBytes  = 1024 * 1024
)

type Evidence struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int64  `json:"bytes"`
}
type Finding struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Status        string     `json:"status"`
	Claim         string     `json:"claim"`
	Observation   string     `json:"observation"`
	FrameworkRefs []string   `json:"framework_refs,omitempty"`
	Evidence      []Evidence `json:"evidence,omitempty"`
	NextAction    string     `json:"next_action"`
}
type Summary struct {
	Verified          int `json:"verified"`
	Partial           int `json:"partial"`
	AttestationNeeded int `json:"attestation_needed"`
	Missing           int `json:"missing"`
	Gap               int `json:"gap"`
	Inherited         int `json:"inherited"`
	Total             int `json:"total"`
}
type Report struct {
	SchemaVersion   string    `json:"schema_version"`
	Tool            string    `json:"tool"`
	GeneratedAt     string    `json:"generated_at"`
	Scope           string    `json:"scope"`
	ScopeNotice     string    `json:"scope_notice"`
	FrameworkNotice string    `json:"framework_notice"`
	Summary         Summary   `json:"summary"`
	Findings        []Finding `json:"findings"`
}

type treeState struct {
	root      string
	files     []string
	workflows []string
	tests     []string
}

// Scan inspects local metadata and CI declarations; it does not call cloud APIs.
func Scan(root string) (Report, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return Report{}, err
	}
	st, err := os.Stat(abs)
	if err != nil {
		return Report{}, err
	}
	if !st.IsDir() {
		return Report{}, fmt.Errorf("not a directory: %s", root)
	}
	t := treeState{root: abs}
	err = filepath.WalkDir(abs, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(abs, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel != "." && skipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		t.files = append(t.files, rel)
		if strings.HasPrefix(rel, ".github/workflows/") && (strings.HasSuffix(rel, ".yml") || strings.HasSuffix(rel, ".yaml")) {
			t.workflows = append(t.workflows, rel)
		}
		name := d.Name()
		if strings.HasSuffix(name, "_test.go") || strings.HasSuffix(name, ".test.ts") || strings.HasSuffix(name, ".spec.ts") || strings.HasSuffix(name, ".test.js") || strings.HasSuffix(name, ".spec.js") || strings.HasPrefix(rel, "tests/") || strings.HasPrefix(rel, "test/") {
			t.tests = append(t.tests, rel)
		}
		if len(t.files) > 50000 {
			return errors.New("repository exceeds 50,000 inspected files; select a smaller scope")
		}
		return nil
	})
	if err != nil {
		return Report{}, err
	}
	sort.Strings(t.files)
	sort.Strings(t.workflows)
	sort.Strings(t.tests)
	now := time.Now().UTC().Format(time.RFC3339)
	r := Report{
		SchemaVersion: "1.0", Tool: "BlockSiFr OpenCMMC", GeneratedAt: now, Scope: "LOCAL_REPOSITORY",
		ScopeNotice:     "This scan inspects local repository artifacts and GitHub Actions declarations only. It does not check deployed configuration, IAM, branch protection, whether CI ran successfully, or organizational controls. VERIFIED means a specific scanner assertion has supporting local evidence, not that a framework control is satisfied.",
		FrameworkNotice: "Framework references are informative alignment annotations, not control-by-control determinations, attestations, audit opinions, or certifications.",
		Findings:        []Finding{},
	}
	r.Findings = append(r.Findings,
		t.fileCheck("SECURITY_POLICY", "Public vulnerability reporting policy", "A SECURITY.md file with nonempty content is present", []string{".github/SECURITY.md", "SECURITY.md"}, "GV.PO-01", "Add SECURITY.md describing responsible vulnerability reporting. "),
		t.fileCheck("CODEOWNERS", "Declared code owners", "A nonempty CODEOWNERS file is present", []string{".github/CODEOWNERS", "docs/CODEOWNERS", "CODEOWNERS"}, "GV.RR-02", "Define code ownership; verify enforcement using GitHub branch protection outside this offline scan."),
		t.dependabotCheck(),
		t.workflowCheck(),
		t.actionPinCheck(),
		t.permissionsCheck(),
		t.testCheck(),
		t.lockCheck(),
		t.sbomCheck(),
		t.fileCheck("INCIDENT_RESPONSE_PLAN", "Incident response plan artifact", "A nonempty incident response document is present", []string{"docs/incident-response.md", "docs/INCIDENT_RESPONSE.md", "INCIDENT_RESPONSE.md", "security/incident-response.md"}, "RS.MA-01", "Create a response plan; test the plan with a real exercise."),
		t.codeScanningCheck(),
		manual("BRANCH_PROTECTION", "Protected branches", "Repository branch protection and required reviews are active", "GV.RR-02", "Connect a read-only GitHub API assessor or supply an independently verifiable branch protection export."),
		manual("PRODUCTION_ACCESS", "Production identity and access", "Production access follows policy and least privilege", "PR.AA-05", "Provide deployment-environment IAM evidence; a source repository cannot demonstrate production access controls."),
		manual("RESPONSE_EFFECTIVENESS", "Response control effectiveness", "Incident response controls have been exercised and measured", "RS.MA-01", "Document a dated tabletop or incident-response exercise and outcome."),
	)
	for _, f := range r.Findings {
		r.Summary.Total++
		switch f.Status {
		case Verified:
			r.Summary.Verified++
		case Partial:
			r.Summary.Partial++
		case AttestationNeeded:
			r.Summary.AttestationNeeded++
		case Missing:
			r.Summary.Missing++
		case Gap:
			r.Summary.Gap++
		case Inherited:
			r.Summary.Inherited++
		}
	}
	return r, nil
}
func skipDir(name string) bool {
	switch name {
	case ".git", "node_modules", "vendor", ".venv", "venv", ".next", "dist", "build", "coverage", "target", ".terraform", ".cache", ".idea":
		return true
	}
	return false
}
func (t treeState) has(rel string) bool {
	n := sort.SearchStrings(t.files, rel)
	return n < len(t.files) && t.files[n] == rel
}
func (t treeState) getEvidence(rel string) (Evidence, error) {
	if !t.has(rel) {
		return Evidence{}, os.ErrNotExist
	}
	file := filepath.Join(t.root, filepath.FromSlash(rel))
	info, err := os.Lstat(file)
	if err != nil {
		return Evidence{}, err
	}
	if !info.Mode().IsRegular() {
		return Evidence{}, errors.New("evidence is not regular file")
	}
	if info.Size() > maxEvidenceBytes {
		return Evidence{}, fmt.Errorf("evidence file over %d bytes", maxEvidenceBytes)
	}
	f, err := os.Open(file)
	if err != nil {
		return Evidence{}, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, maxEvidenceBytes+1))
	if err != nil {
		return Evidence{}, err
	}
	if n > maxEvidenceBytes {
		return Evidence{}, errors.New("evidence too large")
	}
	return Evidence{Path: rel, SHA256: hex.EncodeToString(h.Sum(nil)), Bytes: n}, nil
}
func (t treeState) read(rel string) (string, error) {
	if !t.has(rel) {
		return "", os.ErrNotExist
	}
	p := filepath.Join(t.root, filepath.FromSlash(rel))
	i, e := os.Lstat(p)
	if e != nil {
		return "", e
	}
	if !i.Mode().IsRegular() || i.Size() > maxEvidenceBytes {
		return "", errors.New("not a regular readable artifact under size limit")
	}
	b, e := os.ReadFile(p)
	return string(b), e
}
func (t treeState) first(paths []string) (Evidence, bool) {
	for _, p := range paths {
		if !t.has(p) {
			continue
		}
		e, err := t.getEvidence(p)
		if err == nil && e.Bytes > 0 {
			return e, true
		}
	}
	return Evidence{}, false
}
func finding(id, title, claim, status, observation, framework, next string, evidence ...Evidence) Finding {
	f := Finding{ID: id, Title: title, Claim: claim, Status: status, Observation: observation, NextAction: next}
	if framework != "" {
		f.FrameworkRefs = []string{"NIST-CSF-2.0:" + framework}
	}
	if len(evidence) > 0 {
		f.Evidence = evidence
	}
	return f
}
func (t treeState) fileCheck(id, title, claim string, paths []string, framework, next string) Finding {
	if ev, ok := t.first(paths); ok {
		return finding(id, title, claim, Verified, "Artifact exists and its SHA-256 digest was recorded. Its organizational effectiveness was not evaluated.", framework, "Validate document ownership, adoption, review cadence and enforcement.", ev)
	}
	return finding(id, title, claim, Missing, "No nonempty readable artifact found at the expected locations.", framework, next)
}
func manual(id, title, claim, framework, next string) Finding {
	return finding(id, title, claim, AttestationNeeded, "Cannot substantiate this claim from offline repository files.", framework, next)
}
func (t treeState) dependabotCheck() Finding {
	id, title, claim := "DEPENDENCY_UPDATE_CONFIG", "Automated dependency update configuration", "Dependabot update configuration declares at least one update rule"
	for _, path := range []string{".github/dependabot.yml", ".github/dependabot.yaml"} {
		if !t.has(path) {
			continue
		}
		ev, er := t.getEvidence(path)
		contents, err := t.read(path)
		if er != nil || err != nil {
			return finding(id, title, claim, AttestationNeeded, "Configuration could not be safely inspected.", "PR.PS-02", "Review the Dependabot configuration and file size.")
		}
		if regexp.MustCompile(`(?m)^\s*updates\s*:`).MatchString(contents) && regexp.MustCompile(`(?m)^\s*-\s*package-ecosystem\s*:`).MatchString(contents) {
			return finding(id, title, claim, Partial, "Dependabot update declarations found; execution and coverage have not been verified.", "PR.PS-02", "Verify dependency updates actually run, open PRs, and receive review.", ev)
		}
		return finding(id, title, claim, Gap, "Dependabot configuration exists without detectable update ecosystem rules.", "PR.PS-02", "Add a valid updates list and package-ecosystem entries.", ev)
	}
	return finding(id, title, claim, Missing, "No .github/dependabot.yml or .yaml found.", "PR.PS-02", "Configure automated dependency update PRs or supply evidence of an alternative service.")
}
func (t treeState) workflowCheck() Finding {
	f := finding("CI_DECLARED", "CI workflow declarations", "At least one GitHub Actions workflow is present", Missing, "No GitHub Actions workflow found.", "PR.PS-06", "Add a CI test workflow, or document an independently verified external CI system.")
	if len(t.workflows) == 0 {
		return f
	}
	for _, path := range t.workflows {
		ev, err := t.getEvidence(path)
		if err == nil {
			f.Status = Partial
			f.Observation = fmt.Sprintf("%d GitHub Actions workflow file(s) declared. Execution success and enforcement have not been observed.", len(t.workflows))
			f.Evidence = []Evidence{ev}
			f.NextAction = "Inspect GitHub run history, required-status-check settings and workflow privileges."
			return f
		}
	}
	f.Status = AttestationNeeded
	f.Observation = "Workflow files could not be safely read."
	return f
}

var usesRE = regexp.MustCompile(`(?m)^\s*(?:-\s*)?uses\s*:\s*['"]?([^\s#'"}]+)`)
var pinRE = regexp.MustCompile(`(?i)@[0-9a-f]{40}$`)

func (t treeState) actionPinCheck() Finding {
	id, title, claim := "ACTION_PINNING", "Immutable external action references", "All detected external GitHub Action references are pinned to full commit SHAs"
	if len(t.workflows) == 0 {
		return finding(id, title, claim, AttestationNeeded, "No GitHub Actions workflows to evaluate.", "PR.PS-06", "Add and inspect CI workflows or document external CI.")
	}
	unpinned := []string{}
	total := 0
	proof := []Evidence{}
	for _, path := range t.workflows {
		body, err := t.read(path)
		ev, evErr := t.getEvidence(path)
		if err != nil || evErr != nil {
			return finding(id, title, claim, AttestationNeeded, "At least one workflow was unreadable; cannot assess all action references.", "PR.PS-06", "Review workflow files for links, symlinks, or size violations.")
		}
		proof = append(proof, ev)
		for _, m := range usesRE.FindAllStringSubmatch(body, -1) {
			ref := strings.TrimSpace(m[1])
			if strings.HasPrefix(ref, "./") || strings.HasPrefix(ref, "docker://") {
				continue
			}
			total++
			if !pinRE.MatchString(ref) {
				unpinned = append(unpinned, ref)
			}
		}
	}
	if len(unpinned) > 0 {
		sort.Strings(unpinned)
		if len(unpinned) > 4 {
			unpinned = append(unpinned[:4], "...")
		}
		return finding(id, title, claim, Gap, fmt.Sprintf("%d of %d external action references lack a full commit SHA: %s", len(unpinned), total, strings.Join(unpinned, ", ")), "PR.PS-06", "Pin external actions to verified full 40-character commit SHAs. Review third-party action provenance.", proof...)
	}
	if total == 0 {
		return finding(id, title, claim, AttestationNeeded, "No external action references detected; this check is not applicable or syntax could not be recognized.", "PR.PS-06", "Review workflow references and supply relevant CI supply-chain evidence.", proof...)
	}
	return finding(id, title, claim, Verified, fmt.Sprintf("%d external action reference(s) use full commit SHA pins. SHA provenance and running code were not validated.", total), "PR.PS-06", "Verify action provenance and update pinned references regularly.", proof...)
}
func (t treeState) permissionsCheck() Finding {
	id, title, claim := "WORKFLOW_PERMISSIONS", "Workflow permission declarations", "Workflows declare permissions without a write-all grant"
	if len(t.workflows) == 0 {
		return finding(id, title, claim, AttestationNeeded, "No workflow declarations.", "PR.AA-05", "Inspect CI token permissions in your CI provider.")
	}
	declared := 0
	proof := []Evidence{}
	writeAll := false
	for _, path := range t.workflows {
		body, err := t.read(path)
		ev, eErr := t.getEvidence(path)
		if err != nil || eErr != nil {
			return finding(id, title, claim, AttestationNeeded, "Unable to inspect every workflow.", "PR.AA-05", "Review unreadable workflows.")
		}
		proof = append(proof, ev)
		if regexp.MustCompile(`(?m)^\s*permissions\s*:`).MatchString(body) {
			declared++
		}
		if regexp.MustCompile(`(?mi)^\s*permissions\s*:\s*['"]?write-all\b`).MatchString(body) {
			writeAll = true
		}
	}
	if writeAll {
		return finding(id, title, claim, Gap, "At least one workflow explicitly declares permissions: write-all.", "PR.AA-05", "Replace write-all with minimum per-job GitHub token permissions.", proof...)
	}
	if declared < len(t.workflows) {
		return finding(id, title, claim, AttestationNeeded, fmt.Sprintf("Explicit permissions found in %d of %d workflows; default or inherited permissions are unknown.", declared, len(t.workflows)), "PR.AA-05", "Add explicit least-privilege permissions to every workflow and verify job-level grants.", proof...)
	}
	return finding(id, title, claim, Partial, "Every workflow contains a permissions declaration; granular least-privilege and effective settings need review.", "PR.AA-05", "Inspect job-level permissions and actual GitHub configuration.", proof...)
}
func (t treeState) testCheck() Finding {
	if len(t.tests) == 0 {
		return finding("TEST_ARTIFACTS", "Automated test artifacts", "Repository contains test sources", Missing, "No recognized test files found.", "PR.PS-06", "Add automated tests; examine tests run and required CI status checks.")
	}
	path := t.tests[0]
	ev, err := t.getEvidence(path)
	if err != nil {
		return finding("TEST_ARTIFACTS", "Automated test artifacts", "Repository contains test sources", AttestationNeeded, "Test source found but could not be read safely.", "PR.PS-06", "Verify test sources and CI execution.")
	}
	return finding("TEST_ARTIFACTS", "Automated test artifacts", "Repository contains test sources", Partial, fmt.Sprintf("Found %d test-related file(s). Test execution, coverage and quality are not verified.", len(t.tests)), "PR.PS-06", "Provide test-run results and demonstrate failure-path coverage.", ev)
}
func (t treeState) lockCheck() Finding {
	paths := []string{"go.sum", "package-lock.json", "npm-shrinkwrap.json", "pnpm-lock.yaml", "yarn.lock", "poetry.lock", "uv.lock", "Pipfile.lock", "Cargo.lock", "Gemfile.lock", "composer.lock", "gradle.lockfile", "bun.lock", "bun.lockb"}
	for _, p := range paths {
		if ev, ok := t.first([]string{p}); ok {
			return finding("DEPENDENCY_LOCKFILE", "Dependency lockfile", "Repository root has a dependency lockfile", Verified, "Found a dependency lockfile; resolved dependency safety or freshness is not established.", "PR.PS-02", "Review vulnerable or unmaintained dependencies and establish update policy.", ev)
		}
	}
	return finding("DEPENDENCY_LOCKFILE", "Dependency lockfile", "Repository root has a dependency lockfile", AttestationNeeded, "No standard root-level lockfile detected. This may be appropriate for a dependency-free project.", "PR.PS-02", "Determine whether the project has external dependencies; supply lockfile if supported.")
}
func (t treeState) sbomCheck() Finding {
	names := []string{"sbom.cdx.json", "sbom.spdx.json", "sbom.json", "sbom.xml", "bom.json", "bom.xml"}
	if ev, ok := t.first(names); ok {
		return finding("SBOM_ARTIFACT", "Software bill of materials", "Repository contains a nonempty root-level SBOM", Partial, "A nonempty SBOM file was found. Syntax, completeness, currency and provenance were not evaluated.", "ID.AM-02", "Validate SBOM against its declared schema and regenerate on releases.", ev)
	}
	return finding("SBOM_ARTIFACT", "Software bill of materials", "Repository contains a nonempty root-level SBOM", AttestationNeeded, "No recognized root-level SBOM artifact detected. It may be generated only in CI.", "ID.AM-02", "Generate a CycloneDX or SPDX SBOM and collect signed release build evidence.")
}
func (t treeState) codeScanningCheck() Finding {
	id, title, claim := "CODE_SCAN_DECLARED", "Code scanning declaration", "A GitHub Actions workflow declares a known code scanning action"
	for _, p := range t.workflows {
		body, err := t.read(p)
		if err != nil {
			continue
		}
		if strings.Contains(body, "github/codeql-action/") || strings.Contains(body, "semgrep/semgrep-action") {
			ev, err := t.getEvidence(p)
			if err == nil {
				return finding(id, title, claim, Partial, "A recognized code scanning action is declared, but execution and alerts are not verified.", "PR.PS-06", "Inspect completed code scanning runs and unresolved critical results.", ev)
			}
		}
	}
	return finding(id, title, claim, AttestationNeeded, "No recognized scanning action declaration found. Other tools or external CI may be used.", "PR.PS-06", "Connect existing scanning results or add a scanner and verify results.")
}
