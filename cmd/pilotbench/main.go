// pilotbench summarizes independently reviewed, paired compliance-assurance cases.
// It deliberately does not interpret its output as a certification or audit opinion.
package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	supported      = "SUPPORTED"
	counterexample = "COUNTEREXAMPLE"
	unknown        = "UNKNOWN"
	undetermined   = "UNDETERMINED"
)

type Record struct {
	CaseID    string
	System    string
	Minutes   float64
	Expected  string
	Observed  string
	Reviewed  bool
	Reference string
}

type Summary struct {
	System                             string   `json:"system"`
	TotalCases                         int      `json:"total_cases"`
	Minutes                            float64  `json:"total_review_and_collection_minutes"`
	AdjudicatedCases                   int      `json:"adjudicated_cases"`
	UnadjudicatedCases                 int      `json:"unadjudicated_cases"`
	ObservedSupported                  int      `json:"observed_supported"`
	ObservedCounterexample             int      `json:"observed_counterexample"`
	ObservedUnknown                    int      `json:"observed_unknown"`
	VerifiedNegativeCases              int      `json:"verified_negative_cases"`
	FalseGreen                         int      `json:"false_green"`
	FalseGreenRate                     *float64 `json:"false_green_rate,omitempty"`
	CorrectlySupported                 int      `json:"correctly_supported"`
	CorrectlyIdentifiedCounterexamples int      `json:"correctly_identified_counterexamples"`
}

type Output struct {
	Schema                string    `json:"schema"`
	Notice                string    `json:"notice"`
	AllSystemsPairedCases int       `json:"all_systems_paired_cases"`
	AllSystemsPaired      bool      `json:"all_systems_paired"`
	Systems               []Summary `json:"systems"`
}

func parse(input io.Reader) ([]Record, error) {
	rd := csv.NewReader(input)
	rd.FieldsPerRecord = -1
	rd.TrimLeadingSpace = true
	data, err := rd.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read CSV: %w", err)
	}
	if len(data) < 2 {
		return nil, errors.New("CSV requires header and at least one case")
	}
	headers := []string{"case_id", "system", "minutes", "expected", "observed", "reviewed", "reference"}
	if len(data[0]) != len(headers) {
		return nil, fmt.Errorf("expected %d columns", len(headers))
	}
	for i, h := range headers {
		if strings.TrimSpace(data[0][i]) != h {
			return nil, fmt.Errorf("column %d must be %s", i+1, h)
		}
	}
	var out []Record
	seen := map[string]bool{}
	for i, row := range data[1:] {
		line := i + 2
		if len(row) != len(headers) {
			return nil, fmt.Errorf("line %d: expected %d columns", line, len(headers))
		}
		trim := func(index int) string { return strings.TrimSpace(row[index]) }
		r := Record{CaseID: trim(0), System: trim(1), Expected: trim(3), Observed: trim(4), Reference: trim(6)}
		if r.CaseID == "" || r.System == "" || r.Reference == "" {
			return nil, fmt.Errorf("line %d: case_id, system, reference required", line)
		}
		key := r.System + "\x00" + r.CaseID
		if seen[key] {
			return nil, fmt.Errorf("line %d: duplicate system and case_id", line)
		}
		seen[key] = true
		r.Minutes, err = strconv.ParseFloat(trim(2), 64)
		if err != nil || r.Minutes < 0 || r.Minutes > 1000000 || r.Minutes != r.Minutes {
			return nil, fmt.Errorf("line %d: invalid minutes", line)
		}
		r.Reviewed, err = strconv.ParseBool(strings.ToLower(trim(5)))
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid reviewed boolean", line)
		}
		if r.Observed != supported && r.Observed != counterexample && r.Observed != unknown {
			return nil, fmt.Errorf("line %d: invalid observed status", line)
		}
		if r.Reviewed && (r.Expected != supported && r.Expected != counterexample) {
			return nil, fmt.Errorf("line %d: independently reviewed cases require SUPPORTED or COUNTEREXAMPLE expected", line)
		}
		if !r.Reviewed && r.Expected != undetermined {
			return nil, fmt.Errorf("line %d: unreviewed expected must be UNDETERMINED", line)
		}
		out = append(out, r)
	}
	// Ground-truth labels must agree across systems for every case.
	truth := map[string]string{}
	for _, r := range out {
		if !r.Reviewed {
			continue
		}
		if t, ok := truth[r.CaseID]; ok && t != r.Expected {
			return nil, fmt.Errorf("conflicting independent adjudication for case %s", r.CaseID)
		}
		truth[r.CaseID] = r.Expected
	}
	return out, nil
}

func summarize(records []Record) Output {
	bySystem := map[string]*Summary{}
	caseSystems := map[string]map[string]bool{}
	for _, r := range records {
		s := bySystem[r.System]
		if s == nil {
			s = &Summary{System: r.System}
			bySystem[r.System] = s
		}
		s.TotalCases++
		s.Minutes += r.Minutes
		switch r.Observed {
		case supported:
			s.ObservedSupported++
		case counterexample:
			s.ObservedCounterexample++
		case unknown:
			s.ObservedUnknown++
		}
		if r.Reviewed {
			s.AdjudicatedCases++
			if r.Expected == counterexample {
				s.VerifiedNegativeCases++
				if r.Observed == supported {
					s.FalseGreen++
				}
			}
			if r.Expected == supported && r.Observed == supported {
				s.CorrectlySupported++
			}
			if r.Expected == counterexample && r.Observed == counterexample {
				s.CorrectlyIdentifiedCounterexamples++
			}
		} else {
			s.UnadjudicatedCases++
		}
		if caseSystems[r.CaseID] == nil {
			caseSystems[r.CaseID] = map[string]bool{}
		}
		caseSystems[r.CaseID][r.System] = true
	}
	names := make([]string, 0, len(bySystem))
	for n := range bySystem {
		names = append(names, n)
	}
	sort.Strings(names)
	o := Output{Schema: "freecompliance.pilotbench.v1", Notice: "Operator-entered measurements, not independently verified by this tool. Rates concern only adjudicated technical cases. Not a compliance certification or a product-efficacy claim.", AllSystemsPaired: true}
	for _, n := range names {
		s := bySystem[n]
		if s.VerifiedNegativeCases > 0 {
			rate := float64(s.FalseGreen) / float64(s.VerifiedNegativeCases)
			s.FalseGreenRate = &rate
		}
		o.Systems = append(o.Systems, *s)
	}
	for _, systems := range caseSystems {
		if len(systems) == len(names) {
			o.AllSystemsPairedCases++
		} else {
			o.AllSystemsPaired = false
		}
	}
	return o
}

func run(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("pilotbench", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	input := fs.String("input", "", "CSV evidence measurements")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *input == "" || fs.NArg() != 0 {
		return errors.New("usage: go run ./cmd/pilotbench --input path/to/adjudicated.csv")
	}
	f, err := os.Open(*input)
	if err != nil {
		return err
	}
	defer f.Close()
	records, err := parse(f)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(summarize(records))
}
func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "pilotbench:", err)
		os.Exit(1)
	}
}