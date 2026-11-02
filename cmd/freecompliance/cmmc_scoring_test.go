package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScoreTemplateAndFailClosed(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "score.json")
	output := filepath.Join(dir, "draft.json")
	var printed bytes.Buffer
	if err := cmmcCommand([]string{"score-template", "--scope", "test-enclave", "--out", input}, &printed); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(b, []byte(`"finding": "UNASSESSED"`)) {
		t.Fatal("template not unresolved")
	}
	err = cmmcCommand([]string{"score-draft", "--input", input, "--json", output}, &printed)
	if err == nil {
		t.Fatal("unreviewed input yielded score")
	}
	if _, err = os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("created score despite missing decisions")
	}
	if !strings.Contains(printed.String(), "110 UNASSESSED") {
		t.Fatal("template misleading")
	}
}
