package main

import (
	"strings"
	"testing"
)

const header = "case_id,system,minutes,expected,observed,reviewed,reference\n"

func TestPairedBenchmark(t *testing.T) {
	src := header + "c1,OpenCMMC,12,COUNTEREXAMPLE,COUNTEREXAMPLE,true,review-1\n" + "c1,Baseline,20,COUNTEREXAMPLE,SUPPORTED,true,review-1\n" + "c2,OpenCMMC,10,SUPPORTED,UNKNOWN,true,review-2\n" + "c2,Baseline,25,SUPPORTED,SUPPORTED,true,review-2\n"
	rows, err := parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	o := summarize(rows)
	if !o.AllSystemsPaired || o.AllSystemsPairedCases != 2 {
		t.Fatalf("pairing: %+v", o)
	}
	if o.Systems[0].System != "Baseline" || o.Systems[0].FalseGreen != 1 || o.Systems[0].FalseGreenRate == nil || *o.Systems[0].FalseGreenRate != 1 {
		t.Fatalf("false green: %+v", o.Systems[0])
	}
	if o.Systems[1].FalseGreen != 0 || o.Systems[1].ObservedUnknown != 1 {
		t.Fatalf("free: %+v", o.Systems[1])
	}
}
func TestFailClosedCases(t *testing.T) {
	tests := []string{
		"c1,A,1,SUPPORTED,SUPPORTED,false,ref\n",
		"c1,A,1,UNDETERMINED,SUPPORTED,true,ref\n",
		"c1,A,-1,SUPPORTED,SUPPORTED,true,ref\n",
		"c1,A,1,SUPPORTED,PASS,true,ref\n",
		"c1,A,1,SUPPORTED,SUPPORTED,true,ref\nc1,A,1,SUPPORTED,SUPPORTED,true,ref\n",
		"c1,A,1,SUPPORTED,SUPPORTED,true,ref\nc1,B,1,COUNTEREXAMPLE,UNKNOWN,true,ref\n",
		"c1,A,1,SUPPORTED,SUPPORTED,true,\n",
	}
	for _, tt := range tests {
		if _, err := parse(strings.NewReader(header + tt)); err == nil {
			t.Errorf("accepted invalid input %q", tt)
		}
	}
}
func TestUndeterminedNotScored(t *testing.T) {
	rows, err := parse(strings.NewReader(header + "c1,A,2,UNDETERMINED,SUPPORTED,false,ref\n"))
	if err != nil {
		t.Fatal(err)
	}
	v := summarize(rows).Systems[0]
	if v.AdjudicatedCases != 0 || v.FalseGreenRate != nil || v.UnadjudicatedCases != 1 {
		t.Fatal(v)
	}
}