package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blocksifrdev/freecompliance/internal/cmmc"
)

func TestObjectiveCLIWorkflow(t *testing.T) {
	d := t.TempDir()
	ws := filepath.Join(d, "objectives.json")
	e := filepath.Join(d, "evidence")
	if err := os.Mkdir(e, 0700); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := run([]string{"cmmc", "objectives-init", "--level", "2", "--scope", "Demo", "--out", ws}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "EMPTY") {
		t.Fatalf("inaccurate banner %q", out.String())
	}
	out.Reset()
	report := filepath.Join(d, "objective-report.json")
	if err := run([]string{"cmmc", "objectives-assess", "--input", ws, "--evidence-dir", e, "--json", report}, &out); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	var r cmmc.ObjectiveReport
	if err = json.Unmarshal(b, &r); err != nil {
		t.Fatal(err)
	}
	if r.Summary.MissingObjectiveLists != 110 || r.SPRSScore != "NOT_CALCULATED" {
		t.Fatal("invalid empty workspace report")
	}
}