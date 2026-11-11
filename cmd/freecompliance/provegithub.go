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
	"strings"

	"github.com/blocksifrdev/freecompliance/internal/sourceproof"
	"github.com/blocksifrdev/freecompliance/internal/theorem"
)

func proveGithubCommand(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("prove-github", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	repo := fs.String("repo", "", "repository owner/name")
	jsonDest := fs.String("json", "", "JSON output")
	htmlDest := fs.String("html", "", "HTML output")
	fail := fs.Bool("fail-on-unproven", false, "fail if any claim is not model-supported")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 || *repo == "" {
		return errors.New("--repo owner/name is required; no positional arguments")
	}
	c := sourceproof.Client{Token: os.Getenv("GITHUB_TOKEN")}
	r, err := c.Assess(context.Background(), *repo)
	if err != nil {
		return err
	}
	if *jsonDest != "" {
		b, e := json.MarshalIndent(r, "", "  ")
		if e != nil {
			return e
		}
		if e = writeSafe(*jsonDest, append(b, '\n')); e != nil {
			return e
		}
	}
	if *htmlDest != "" {
		if err := renderGithubProof(*htmlDest, r); err != nil {
			return err
		}
	}
	fmt.Fprintf(out, "OpenCMMC | source-bound theorem demo\nRepo: %s | Branch: %s | Source complete: %t\n", r.Repository, r.DefaultBranch, r.SourceComplete)
	unproven := false
	for _, c := range r.Proof.Results {
		fmt.Fprintf(out, "%-22s %s\n", c.Status, c.ID)
		if c.Status != theorem.ModelSupported {
			unproven = true
		}
	}
	for _, w := range r.Warnings {
		fmt.Fprintln(out, "LIMITATION:", w)
	}
	if *fail && unproven {
		return errors.New("one or more claims are not model-supported")
	}
	return nil
}

func renderGithubProof(path string, r sourceproof.Result) error {
	const doc = `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>OpenCMMC | Live source-bound proofs</title><style>body{max-width:900px;margin:35px auto;padding:0 18px;font:16px/1.5 system-ui;color:#182538;background:#f6f8fc}section{background:white;border:1px solid #d9e3f0;padding:22px;border-radius:14px;margin:14px 0}h1{margin-bottom:3px}small{color:#52647a}code{overflow-wrap:anywhere}.UNKNOWN{color:#855b04}.MODEL_COUNTEREXAMPLE{color:#a32e21}.MODEL_SUPPORTED{color:#1d682e}</style></head><body><h1>OpenCMMC | Source-bound theorem</h1><p>Scope: {{.Repository}} / {{.DefaultBranch}} | Source complete: {{.SourceComplete}}</p><section><h2>Bounded results</h2>{{range .Proof.Results}}<p><strong class="{{.Status}}">{{.Status}}</strong> | {{.ID}}</p><p>{{.Explanation}}</p>{{end}}</section><section><h2>Source hashes and freshness</h2>{{range .Sources}}<p><code>{{.Endpoint}}</code><br><small>SHA256: {{.SHA256}} | Retrieved: {{.RetrievedAt}} | API Date: {{.SourceDate}} | Fresh: {{.Fresh}}</small></p>{{end}}</section><section><h2>Limits and next evidence</h2>{{range .Warnings}}<p>{{.}}</p>{{end}}<p>{{.Limitations}}</p></section></body></html>`
	t, err := template.New("sourceproof").Parse(doc)
	if err != nil {
		return err
	}
	var b strings.Builder
	if err := t.Execute(&b, r); err != nil {
		return err
	}
	return writeSafe(path, []byte(b.String()))
}