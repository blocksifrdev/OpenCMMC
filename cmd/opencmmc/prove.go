package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/blocksifrdev/opencmmc/internal/theorem"
)

func proveCommand(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("prove", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	input := fs.String("input", "", "bounded-model request JSON")
	jsonDest := fs.String("json", "", "output JSON proof report")
	htmlDest := fs.String("html", "", "output standalone HTML report")
	fail := fs.Bool("fail-on-counterexample", false, "exit nonzero if any modeled counterexample exists")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("unexpected positional argument")
	}
	if *input == "" {
		return errors.New("--input is required")
	}
	f, err := os.Open(*input)
	if err != nil {
		return err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return err
	}
	if !st.Mode().IsRegular() {
		return errors.New("input must be a regular file")
	}
	if st.Size() > 1024*1024 {
		return errors.New("input exceeds 1 MiB")
	}
	dec := json.NewDecoder(io.LimitReader(f, 1024*1024+1))
	dec.DisallowUnknownFields()
	var r theorem.Request
	if err := dec.Decode(&r); err != nil {
		return fmt.Errorf("invalid theorem request: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return errors.New("input must contain exactly one JSON object")
	}
	report, err := theorem.Evaluate(r)
	if err != nil {
		return err
	}
	if *jsonDest != "" {
		b, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		if err = writeSafe(*jsonDest, append(b, '\n')); err != nil {
			return err
		}
	}
	if *htmlDest != "" {
		if err := renderProof(*htmlDest, report); err != nil {
			return err
		}
	}
	fmt.Fprintf(out, "OpenCMMC | Theorem Intelligence (bounded, model-relative)\nScope: %s\nUniverse: %d | Consistent worlds: %d | Admitted evidence: %d | Excluded: %d\n", report.Scope, report.PossibleStates, report.ConsistentStates, report.AdmittedEvidence, len(report.IgnoredEvidence))
	counterexamples := 0
	for _, v := range report.Results {
		fmt.Fprintf(out, "%-22s %s | %d counterexample worlds\n", v.Status, v.ID, v.CounterexampleStates)
		if v.EvidencePlan != nil {
			fmt.Fprintf(out, "  Hypothetical least-cost evidence to verify: %s (cost %d)\n", strings.Join(v.EvidencePlan.CandidateIDs, ", "), v.EvidencePlan.TotalCost)
		}
		if v.Status == theorem.ModelCounterexample {
			counterexamples++
		}
	}
	fmt.Fprintln(out, "No certification, real-world safety proof, or enforcement decision is implied.")
	if *fail && counterexamples > 0 {
		return fmt.Errorf("%d model counterexample(s) found", counterexamples)
	}
	return nil
}

func renderProof(path string, r theorem.Report) error {
	type readableWitness struct{ Key, Value string }
	type result struct {
		theorem.ClaimResult
		Witness []readableWitness
	}
	model := struct {
		theorem.Report
		Claims []result
	}{Report: r}
	for _, v := range r.Results {
		x := result{ClaimResult: v}
		for k, val := range v.Counterexample {
			x.Witness = append(x.Witness, readableWitness{k, val})
		}
		sort.Slice(x.Witness, func(i, j int) bool { return x.Witness[i].Key < x.Witness[j].Key })
		model.Claims = append(model.Claims, x)
	}
	const page = `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>OpenCMMC | Theorem Intelligence</title><style>
 :root{font:16px/1.5 system-ui,sans-serif;color:#15263b;background:#f5f7fa}*{box-sizing:border-box}body{max-width:960px;margin:auto;padding:32px 18px}header{background:#12223b;color:white;padding:32px;border-radius:18px}h1{margin:0;font-size:2rem}h2{font-size:1.15rem;margin:0}main{display:grid;gap:16px;margin:20px 0}.card{background:white;border:1px solid #dce3eb;border-radius:14px;padding:22px}.status{display:inline-block;font-weight:700;font-size:.78rem;border-radius:50px;padding:5px 10px;background:#e4eaf5}.MODEL_SUPPORTED{background:#ccebd6;color:#15552b}.MODEL_COUNTEREXAMPLE{background:#ffdfd9;color:#81291d}.UNKNOWN{background:#fff2c6;color:#664e00}code{overflow-wrap:anywhere}small,.muted{color:#55647a}header .muted{color:#cad5e5}.warn{border-left:4px solid #b76c1f;padding:10px;background:#fffaed}dl{display:grid;grid-template-columns:max-content 1fr;gap:6px 14px}dt{font-weight:600}dd{margin:0}footer{font-size:.86rem}</style></head><body>
 <header><p>BLOCKSIFR / OPENCMMC</p><h1>Theorem Intelligence</h1><p>Bounded claim verification, modeled counterexamples and evidence planning.</p><p class="muted">Scope: {{.Scope}} | As of: {{.AsOf}}</p></header>
 <main><section class="card"><h2>Finite model boundary</h2><p>{{.PossibleStates}} possible worlds, {{.ConsistentStates}} evidence-consistent worlds, {{.AdmittedEvidence}} admitted observations, {{len .IgnoredEvidence}} ignored.</p><p><small>Model fingerprint: <code>{{.ModelSHA256}}</code></small></p><div class="warn">{{.Disclaimer}}</div></section>
 {{range .Claims}}<section class="card"><p><span class="status {{.Status}}">{{.Status}}</span></p><h2>{{.ID}}</h2><p>{{.Description}}</p>{{if .FrameworkRef}}<p><small>Framework annotation: {{.FrameworkRef}}</small></p>{{end}}<p>{{.Explanation}}</p><p><strong>Models satisfying claim:</strong> {{.SatisfyingStates}} | <strong>Counterexample models:</strong> {{.CounterexampleStates}}</p>{{if .Witness}}<h3>Modeled counterexample</h3><dl>{{range .Witness}}<dt>{{.Key}}</dt><dd>{{.Value}}</dd>{{end}}</dl>{{end}}{{if .EvidencePlan}}<h3>Minimum-cost candidate evidence (hypothetical)</h3><p>Cost: {{.EvidencePlan.TotalCost}}. Candidate IDs: {{range .EvidencePlan.CandidateIDs}}<code>{{.}}</code> {{end}}</p><p><small>Must be independently obtained and verified. Exact only within the supplied bounded candidate pool.</small></p>{{end}}</section>{{end}}
 </main><footer>OpenCMMC is not a certification body. Model coverage, evidence source truth, and validity require independent validation.</footer></body></html>`
	t, err := template.New("report").Parse(page)
	if err != nil {
		return err
	}
	var b strings.Builder
	if err := t.Execute(&b, model); err != nil {
		return err
	}
	return writeSafe(path, []byte(b.String()))
}
