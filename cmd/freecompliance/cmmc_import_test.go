package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIImportRejectsMissingSource(t *testing.T) {
	var out bytes.Buffer
	e := cmmcCommand([]string{"objectives-import", "--csv", filepath.Join(t.TempDir(), "absent.csv"), "--sha256", strings.Repeat("0", 64), "--scope", "enclave", "--out", filepath.Join(t.TempDir(), "import.json")}, &out)
	if e == nil {
		t.Fatal("missing file accepted")
	}
}
func TestCLIImportRejectsCorruptWithoutOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fake.csv")
	dest := filepath.Join(dir, "output.json")
	if e := os.WriteFile(path, []byte("Requirement,Assessment Objective\n3.1.1,[a] mock\n"), 0600); e != nil {
		t.Fatal(e)
	}
	var out bytes.Buffer
	e := cmmcCommand([]string{"objectives-import", "--csv", path, "--sha256", strings.Repeat("0", 64), "--scope", "enclave", "--out", dest}, &out)
	if e == nil {
		t.Fatal("corrupt digest accepted")
	}
	if _, er := os.Stat(dest); !os.IsNotExist(er) {
		t.Fatal("unsafe output created")
	}
}
