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

func cmmcCommand(args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: freecompliance cmmc init|assess|verify-bundle|objectives-init|objectives-import|objectives-assess|score-template|score-draft|ui")
	}
	switch args[0] {
	case "ui":
		return cmmcUICommand(args[1:], out)
	case "score-template", "score-draft":
		return cmmcScoringCommand(args, out)
	case "objectives-init", "objectives-import", "objectives-assess":
		return cmmcObjectivesCommand(args, out)
	case "init":
		fs := flag.NewFlagSet("cmmc init", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		level := fs.Int("level", 2, "CMMC level 1 or 2")
		scope := fs.String("scope", "", "explicit system/enclave scope")
		dest := fs.String("out", "", "workspace output JSON")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() > 0 {
			return errors.New("unexpected positional arguments")
		}
		if *dest == "" {
			return errors.New("--out is required")
		}
		w, err := cmmc.NewWorkspace(*level, *scope)
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
		fmt.Fprintf(out, "Created offline CMMC Level %d workspace with %d requirements. All are UNASSESSED. No certification implied.\n", w.Level, len(w.Controls))
		return nil
	case "assess":
		fs := flag.NewFlagSet("cmmc assess", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		input := fs.String("input", "", "workspace JSON path")
		root := fs.String("evidence-dir", "", "root of approved local evidence (never uploaded)")
		dest := fs.String("json", "", "output readiness JSON")
		days := fs.Int("max-age-days", 90, "local evidence age threshold; policy-defined")
		atValue := fs.String("at", "", "optional RFC3339 assessment instant; defaults to now")
		fail := fs.Bool("fail-on-unresolved", false, "fail CI for any non-DOCUMENTED control")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() > 0 {
			return errors.New("unexpected positional arguments")
		}
		if *input == "" || *root == "" {
			return errors.New("--input and --evidence-dir required")
		}
		b, err := os.ReadFile(*input)
		if err != nil {
			return err
		}
		if len(b) > 4<<20 {
			return errors.New("workspace exceeds 4 MiB")
		}
		var w cmmc.Workspace
		if err = json.Unmarshal(b, &w); err != nil {
			return err
		}
		at := time.Now().UTC()
		if *atValue != "" {
			at, err = time.Parse(time.RFC3339, *atValue)
			if err != nil {
				return err
			}
		}
		r, err := cmmc.Evaluate(w, *root, at, *days)
		if err != nil {
			return err
		}
		if *dest != "" {
			result, er := json.MarshalIndent(r, "", "  ")
			if er != nil {
				return er
			}
			if er = writeSafe(*dest, append(result, '\n')); er != nil {
				return er
			}
		}
		fmt.Fprintf(out, "CMMC Level %d | %s | total %d | documented %d | gaps %d | unassessed %d | unverifiable %d | scope review %d\n", r.Level, r.Scope, r.Summary.Total, r.Summary.Documented, r.Summary.Gaps, r.Summary.Unassessed, r.Summary.Unverifiable, r.Summary.ScopeReview)
		fmt.Fprintln(out, "NOT CERTIFIED. A linked artifact or matching hash does not establish control effectiveness. No SPRS submission.")
		if *fail && r.Summary.Documented != r.Summary.Total {
			return errors.New("one or more controls are unresolved")
		}
		return nil
	case "verify-bundle":
		fs := flag.NewFlagSet("cmmc verify-bundle", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		bundle := fs.String("bundle", "", "offline update package or catalog")
		sig := fs.String("signature", "", "base64 detached Ed25519 signature")
		key := fs.String("public-key", "", "trusted Ed25519 base64 public key provisioned out of band")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() > 0 || *bundle == "" || *sig == "" || *key == "" {
			return errors.New("--bundle, --signature and --public-key are required")
		}
		digest, err := cmmc.VerifySignedBundle(*bundle, *sig, *key)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "VALID Ed25519 signature | SHA-256 %s | public-key trust must be established out-of-band. No compliance attestation implied.\n", digest)
		return nil
	default:
		return fmt.Errorf("unknown CMMC command %q; use init, assess or verify-bundle", args[0])
	}
}
