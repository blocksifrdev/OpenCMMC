package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/blocksifrdev/freecompliance/internal/cmmc"
)

func cmmcObjectivesCommand(args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: cmmc objectives-init|objectives-assess")
	}
	switch args[0] {
	case "objectives-init":
		fs := flag.NewFlagSet("cmmc objectives-init", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		level := fs.Int("level", 2, "CMMC Level 1 or 2")
		scope := fs.String("scope", "", "system boundary / enclave")
		dest := fs.String("out", "", "offline objective workspace JSON")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 || *dest == "" {
			return errors.New("--out required; no positional arguments")
		}
		w, err := cmmc.NewObjectiveWorkspace(*level, *scope)
		if err != nil {
			return err
		}
		b, err := json.MarshalIndent(w, "", "  ")
		if err != nil {
			return err
		}
		if err = writeSafe(*dest, append(b, '\n')); err != nil {
			return err
		}
		fmt.Fprintf(out, "Created %d-control objective workspace. Objective lists EMPTY until populated from independently reviewed NIST procedures. No compliance or SPRS score.\n", len(w.Controls))
		return nil
	case "objectives-assess":
		fs := flag.NewFlagSet("cmmc objectives-assess", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		input := fs.String("input", "", "objective workspace JSON")
		evidence := fs.String("evidence-dir", "", "approved local evidence directory")
		dest := fs.String("json", "", "JSON findings output")
		days := fs.Int("max-age-days", 90, "customer policy evidence age, not official CMMC recency")
		clock := fs.String("at", "", "optional RFC3339 evaluation time")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 || *input == "" || *evidence == "" {
			return errors.New("--input and --evidence-dir required; no positional arguments")
		}
		fi, err := os.Stat(*input)
		if err != nil {
			return err
		}
		if fi.Size() > 4<<20 {
			return errors.New("objective workspace exceeds 4 MiB")
		}
		b, err := os.ReadFile(*input)
		if err != nil {
			return err
		}
		var w cmmc.ObjectiveWorkspace
		if err = json.Unmarshal(b, &w); err != nil {
			return err
		}
		t := time.Now().UTC()
		if *clock != "" {
			t, err = time.Parse(time.RFC3339, *clock)
			if err != nil {
				return err
			}
		}
		r, err := cmmc.AssessObjectives(w, *evidence, t, *days)
		if err != nil {
			return err
		}
		if *dest != "" {
			j, e := json.MarshalIndent(r, "", "  ")
			if e != nil {
				return e
			}
			if e = writeSafe(*dest, append(j, '\n')); e != nil {
				return e
			}
		}
		fmt.Fprintf(out, "CMMC objective intake | controls %d | empty lists %d | reviewer reported gaps %d | evidence review ready %d | unresolved %d\n", r.Summary.TotalControls, r.Summary.MissingObjectiveLists, r.Summary.GapsReported, r.Summary.ReviewReady, r.Summary.Unresolved)
		fmt.Fprintln(out, "Catalog completeness NOT validated. Evidence hashes do not prove operating effectiveness. No SPRS score, CMMC certification or government submission.")
		return nil
	default:
		return fmt.Errorf("unknown objective command %q", args[0])
	}
}