package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProveCommandCounterexample(t *testing.T) {
	path := "../../examples/theorem/branch-controls.json"
	var out bytes.Buffer
	j := filepath.Join(t.TempDir(), "out.json")
	h := filepath.Join(t.TempDir(), "out.html")
	if err := run([]string{"prove", "--input", path, "--json", j, "--html", h}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "MODEL_COUNTEREXAMPLE") || !strings.Contains(out.String(), "inspect-effective-bypass-rules") {
		t.Fatal(out.String())
	}
	for _, p := range []string{j, h} {
		b, err := os.ReadFile(p)
		if err != nil || len(b) == 0 {
			t.Fatalf("missing output %s: %v", p, err)
		}
	}
	out.Reset()
	if err := run([]string{"prove", "--input", path, "--fail-on-counterexample"}, &out); err == nil {
		t.Fatal("expected explicit gate failure")
	}
}
func TestProveRejectsTrailingJSON(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bad.json")
	os.WriteFile(p, []byte(`{"a":1} {"b":2}`), 0600)
	if err := run([]string{"prove", "--input", p}, &bytes.Buffer{}); err == nil {
		t.Fatal("expected failure")
	}
}
