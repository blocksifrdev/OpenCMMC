package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/blocksifrdev/OpenCMMC/internal/githubscan"
	"github.com/blocksifrdev/OpenCMMC/internal/scan"
)

var releaseVersion = "v0.4.0-dev"

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "opencmmc:", err)
		os.Exit(1)
	}
}
func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return usage(out)
	}
	switch args[0] {
	case "scan":
		return scanCommand(args[1:], out)
	case "github":
		return githubCommand(args[1:], out)
	case "compare":
		return compareCommand(args[1:], out)
	case "prove":
		return proveCommand(args[1:], out)
	case "prove-github":
		return proveGithubCommand(args[1:], out)
	case "cmmc":
		return cmmcCommand(args[1:], out)
	case "version", "--version", "-v":
		fmt.Fprintf(out, "OpenCMMC %s\n", releaseVersion)
		return nil
	case "help", "-h", "--help":
		return usage(out)
	default:
		return fmt.Errorf("unknown command %q; use 'opencmmc help'", args[0])
	}
}
func usage(w io.Writer) error {
	_, err := fmt.Fprint(w, `OpenCMMC by BlockSiFr | free, local, evidence-first readiness

Usage:
  opencmmc scan --path . [--json report.json] [--html report.html] [--fail-on-gap]
  opencmmc github --repo owner/name [--json report.json] [--html report.html] [--fail-on-gap]
  opencmmc compare --before old.json --after new.json
  opencmmc prove --input obligations.json [--json proof.json] [--html proof.html] [--fail-on-counterexample]
  opencmmc prove-github --repo owner/name [--json github-proof.json] [--html github-proof.html] [--fail-on-unproven]
  opencmmc cmmc ui --data-dir ./private-assessment --port 8765
  opencmmc cmmc init --level 2 --scope "enclave-1" --out workspace.json
  opencmmc cmmc assess --input workspace.json --evidence-dir ./evidence --json readiness.json
  opencmmc cmmc objectives-init --level 2 --scope "enclave-1" --out objectives.json
  opencmmc cmmc objectives-import --csv nist-171a.csv --sha256 EXPECTED_SHA --scope "enclave-1" --out objectives.json
  opencmmc cmmc objectives-assess --input objectives.json --evidence-dir evidence --json objective-report.json
  opencmmc cmmc score-template --scope "enclave-1" --out score-input.json
  opencmmc cmmc score-draft --input score-reviewed.json --json score-draft.json
  opencmmc cmmc verify-bundle --bundle update.tar --signature update.sig --public-key trusted.pub
  opencmmc version

A local scan DOES NOT certify SOC 2, CMMC, ISO 27001 or any NIST framework.
VERIFIED means a specific check passed using local artifacts, not organizational compliance.
Local browser UI binds only to 127.0.0.1 and serves no remote scripts or telemetry.
Local scan needs no network, secrets, cloud credentials or telemetry.
GitHub scan makes read-only requests and optionally uses GITHUB_TOKEN.
The theorem command evaluates bounded caller-defined models, never certifies compliance.
`)
	return err
}
func scanCommand(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	path := fs.String("path", ".", "repository root")
	jsonDest := fs.String("json", "", "JSON output file")
	htmlDest := fs.String("html", "", "HTML output file")
	fail := fs.Bool("fail-on-gap", false, "exit with error if an observed GAP exists")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected positional argument: %s", fs.Arg(0))
	}
	r, err := scan.Scan(*path)
	if err != nil {
		return err
	}
	if *jsonDest != "" {
		b, er := json.MarshalIndent(r, "", "  ")
		if er != nil {
			return er
		}
		if er = writeSafe(*jsonDest, append(b, '\n')); er != nil {
			return er
		}
	}
	if *htmlDest != "" {
		if er := renderHTML(*htmlDest, r); er != nil {
			return er
		}
	}
	fmt.Fprintf(out, "OpenCMMC | repository evidence readiness\nVerified: %d | Partial: %d | Missing: %d | Gaps: %d | Attestation needed: %d | Total: %d\n\n", r.Summary.Verified, r.Summary.Partial, r.Summary.Missing, r.Summary.Gap, r.Summary.AttestationNeeded, r.Summary.Total)
	for _, f := range r.Findings {
		fmt.Fprintf(out, "%-21s %-25s %s\n", f.Status, f.ID, f.Observation)
	}
	fmt.Fprintln(out, "\nEvidence assertions only. No compliance, certification, or audit opinion is implied.")
	if *fail && r.Summary.Gap > 0 {
		return fmt.Errorf("%d observed gap(s) and --fail-on-gap was set", r.Summary.Gap)
	}
	return nil
}

// githubCommand reads repository metadata using GitHub's REST API and never mutates settings.
func githubCommand(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("github", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	repo := fs.String("repo", "", "GitHub repository in owner/name format")
	jsonDest := fs.String("json", "", "JSON report output")
	htmlDest := fs.String("html", "", "HTML report output")
	fail := fs.Bool("fail-on-gap", false, "exit with error on an observed GAP")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected positional argument")
	}
	if *repo == "" {
		return errors.New("--repo owner/name is required")
	}
	client := githubscan.Client{Token: os.Getenv("GITHUB_TOKEN")}
	r, err := client.Assess(context.Background(), *repo)
	if err != nil {
		return err
	}
	if *jsonDest != "" {
		b, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			return err
		}
		if err = writeSafe(*jsonDest, append(b, '\n')); err != nil {
			return err
		}
	}
	if *htmlDest != "" {
		if err = renderHTML(*htmlDest, r); err != nil {
			return err
		}
	}
	fmt.Fprintf(out, "OpenCMMC | read-only GitHub metadata\nScope: %s\nVerified: %d | Gaps: %d | Needs attestation: %d\n", r.Scope, r.Summary.Verified, r.Summary.Gap, r.Summary.AttestationNeeded)
	for _, f := range r.Findings {
		fmt.Fprintf(out, "%-21s %-27s %s\n", f.Status, f.ID, f.Observation)
	}
	fmt.Fprintln(out, "Repository metadata assertions only. Not compliance certification.")
	if *fail && r.Summary.Gap > 0 {
		return fmt.Errorf("%d observed gap(s) with --fail-on-gap", r.Summary.Gap)
	}
	return nil
}
func writeSafe(path string, data []byte) error {
	if path == "" {
		return errors.New("missing output path")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if st, err := os.Lstat(abs); err == nil {
		if !st.Mode().IsRegular() {
			return fmt.Errorf("refusing to overwrite nonregular path %s", path)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	// Permission 0600: do not accidentally expose local evidence to other users.
	return os.WriteFile(abs, data, 0600)
}
func readReport(path string) (scan.Report, error) {
	f, err := os.Open(path)
	if err != nil {
		return scan.Report{}, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return scan.Report{}, err
	}
	if st.Size() > 5*1024*1024 {
		return scan.Report{}, errors.New("report exceeds 5 MiB")
	}
	dec := json.NewDecoder(io.LimitReader(f, 5*1024*1024+1))
	dec.DisallowUnknownFields()
	var r scan.Report
	if err := dec.Decode(&r); err != nil {
		return r, err
	}
	if r.SchemaVersion != "1.0" {
		return r, fmt.Errorf("unsupported schema version %q", r.SchemaVersion)
	}
	return r, nil
}
func compareCommand(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	before := fs.String("before", "", "prior report")
	after := fs.String("after", "", "later report")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *before == "" || *after == "" {
		return errors.New("both --before and --after are required")
	}
	a, err := readReport(*before)
	if err != nil {
		return fmt.Errorf("before: %w", err)
	}
	b, err := readReport(*after)
	if err != nil {
		return fmt.Errorf("after: %w", err)
	}
	prev := map[string]string{}
	next := map[string]string{}
	for _, f := range a.Findings {
		prev[f.ID] = f.Status
	}
	for _, f := range b.Findings {
		next[f.ID] = f.Status
	}
	names := map[string]bool{}
	for id := range prev {
		names[id] = true
	}
	for id := range next {
		names[id] = true
	}
	keys := make([]string, 0, len(names))
	for id := range names {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	changed := 0
	for _, id := range keys {
		if prev[id] != next[id] {
			old := prev[id]
			newVal := next[id]
			if old == "" {
				old = "(not assessed)"
			}
			if newVal == "" {
				newVal = "(not assessed)"
			}
			fmt.Fprintf(out, "%s: %s -> %s\n", id, old, newVal)
			changed++
		}
	}
	if changed == 0 {
		fmt.Fprintln(out, "No finding status changes.")
	}
	fmt.Fprintf(out, "Changed assertions: %d. Compare reflects local scan assertions, not framework compliance.\n", changed)
	return nil
}

const htmlTemplate = `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>OpenCMMC | Evidence status</title>
<style>
:root{color-scheme:light dark;font-family:system-ui,-apple-system,sans-serif;font-size:16px}*{box-sizing:border-box}body{max-width:1080px;margin:0 auto;padding:38px 20px;background:#f8fafc;color:#152332}header{padding:30px;border-radius:18px;background:#101e33;color:#eef4ff}h1{font-size:2.2rem;margin:0 0 10px}.muted{opacity:.78}.stats{display:grid;grid-template-columns:repeat(auto-fit,minmax(150px,1fr));gap:12px;margin:24px 0}.stat{background:white;border-radius:12px;padding:19px;border:1px solid #dce4eb}.stat strong{display:block;font-size:2.0rem}.stat span{color:#526070;font-size:.85rem}.findings{display:grid;gap:12px}.finding{background:white;border:1px solid #dce4eb;border-radius:12px;padding:20px}.finding h2{font-size:1.12rem;margin:10px 0}.status{display:inline-block;padding:4px 10px;font-size:.75rem;font-weight:700;border-radius:100px;letter-spacing:.04em;background:#e8eff6;color:#273b54}.status.VERIFIED{background:#d8f5e3;color:#166534}.status.GAP{background:#ffe0dc;color:#972a1a}.status.MISSING{background:#fff2cd;color:#815500}.status.ATTESTATION_NEEDED{background:#e7eafe;color:#39468b}.status.PARTIAL{background:#e0f2fd;color:#164e70}.row{font-size:.88rem;color:#4a5865}.evidence{padding:10px;background:#f1f5f9;border-radius:8px;overflow-wrap:anywhere}.small{font-size:.78rem}footer{padding:26px 0;color:#556575}code{font-size:.85rem} @media(prefers-color-scheme:dark){body{background:#0c1523;color:#edf3fb}.stat,.finding{background:#182638;border-color:#2c405b}.stat span,.row,footer{color:#bccadb}.evidence{background:#24384f}} </style>
</head><body>
<header><div class="muted">BLOCKSIFR / OPEN SOURCE</div><h1>OpenCMMC</h1><div>Evidence readiness for a local software repository</div><p class="muted">Generated {{.GeneratedAt}} | Scope: {{.Scope}}</p></header>
<section class="stats"><div class="stat"><strong>{{.Summary.Verified}}</strong><span>VERIFIED ASSERTIONS</span></div><div class="stat"><strong>{{.Summary.Partial}}</strong><span>PARTIAL</span></div><div class="stat"><strong>{{.Summary.Gap}}</strong><span>OBSERVED GAPS</span></div><div class="stat"><strong>{{.Summary.Missing}}</strong><span>MISSING ARTIFACTS</span></div><div class="stat"><strong>{{.Summary.AttestationNeeded}}</strong><span>NEEDS ATTESTATION</span></div></section>
<p><strong>Meaning:</strong> {{.ScopeNotice}}</p>
<main class="findings">{{range .Findings}}<article class="finding"><span class="status {{.Status}}">{{.Status}}</span><h2>{{.Title}}</h2><div class="row">{{.ID}} | {{range .FrameworkRefs}}{{.}} {{end}}</div><p><strong>Assertion:</strong> {{.Claim}}</p><p>{{.Observation}}</p><p><strong>Next action:</strong> {{.NextAction}}</p>{{range .Evidence}}<div class="evidence small">FILE: <code>{{.Path}}</code><br>SHA-256: <code>{{.SHA256}}</code><br>Size: {{.Bytes}} bytes</div>{{end}}</article>{{end}}</main>
<footer>{{.FrameworkNotice}}<p>OpenCMMC is a free assessment aid. Its findings are not legal advice, official attestations, or compliance certifications.</p></footer>
</body></html>`

func renderHTML(dest string, r scan.Report) error {
	t, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return err
	}
	var b strings.Builder
	if err = t.Execute(&b, r); err != nil {
		return err
	}
	return writeSafe(dest, []byte(b.String()))
}
