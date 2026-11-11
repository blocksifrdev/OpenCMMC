package cmmc

import "testing"

func TestNamespaceCompatibility(t *testing.T) {
    for _, tc := range []struct { actual, expected string; accepted bool }{
        {Schema, Schema, true},
        {ObjectiveSchema, ObjectiveSchema, true},
        {"previous.cmmc.workspace.v1", Schema, true},
        {"previous.cmmc.objectives.v1", ObjectiveSchema, true},
        {"previous.cmmc.workspace.v2", Schema, false},
        {"previous.cmmc.objectives.v2", ObjectiveSchema, false},
        {"other.cmmc.readiness.v1", Schema, false},
        {"x.cmmc.workspace.v1", Schema, false},
        {"", Schema, false},
    } {
        if got := compatibleCMMCSchema(tc.actual, tc.expected); got != tc.accepted {
            t.Errorf("compatible(%q,%q)=%v, want %v", tc.actual,tc.expected,got,tc.accepted)
        }
    }
}
