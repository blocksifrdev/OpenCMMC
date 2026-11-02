package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/blocksifrdev/freecompliance/internal/cmmc"
)

func cmmcScoringCommand(args []string, out io.Writer) error {
	switch args[0] {
	case "score-template":
		fs := flag.NewFlagSet("cmmc score-template", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		scope := fs.String("scope", "", "system / enclave boundary")
		dest := fs.String("out", "", "output JSON for human review")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 || *dest == "" {
			return errors.New("--out required; no positional arguments")
		}
		w, err := cmmc.NewWorkspace(2, *scope)
		if err != nil {
			return err
		}
		in := cmmc.DraftScoreInput{Baseline: w.Baseline, Scope: *scope}
		for _, c := range w.Controls {
			in.Adjudications = append(in.Adjudications, cmmc.Adjudication{ID: c.ID, Finding: "UNASSESSED"})
		}
		b, err := json.MarshalIndent(in, "", "  ")
		if err != nil {
			return err
		}
		if err = writeSafe(*dest, append(b, '\n')); err != nil {
			return err
		}
		fmt.Fprintln(out, "Draft score template created with 110 UNASSESSED requirements. No score until all human-reviewed decisions and a confirmed SSP are provided.")
		return nil
	case "score-draft":
		fs := flag.NewFlagSet("cmmc score-draft", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		input := fs.String("input", "", "human-adjudicated 110-requirement JSON")
		dest := fs.String("json", "", "draft scoring output JSON")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 || *input == "" || *dest == "" {
			return errors.New("--input and --json required; no positional arguments")
		}
		fi, err := os.Stat(*input)
		if err != nil {
			return err
		}
		if fi.Size() > 2<<20 {
			return errors.New("score input exceeds 2 MiB")
		}
		b, err := os.ReadFile(*input)
		if err != nil {
			return err
		}
		var in cmmc.DraftScoreInput
		dec := json.NewDecoder(bytes.NewReader(b))
		dec.DisallowUnknownFields()
		if err = dec.Decode(&in); err != nil {
			return err
		}
		if strings.TrimSpace(in.Scope) == "" {
			return errors.New("scope required")
		}
		result, err := cmmc.CalculateDraftScore(in)
		if err != nil {
			return err
		}
		b, err = json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		if err = writeSafe(*dest, append(b, '\n')); err != nil {
			return err
		}
		fmt.Fprintf(out, "DRAFT ONLY: arithmetic %d / 110; deductions %d. User-entered findings not independently validated. No SPRS submission or certification.\n", result.CalculatedDraftScore, result.DeductedPoints)
		return nil
	default:
		return fmt.Errorf("unknown scoring command %q", args[0])
	}
}
