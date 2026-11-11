package theorem

import "testing"

func TestNamespaceCompatibleTheorem(t *testing.T) {
    for _, tc := range []struct { value string; accepted bool }{
        {Schema, true}, {"previous.theorem.v1", true},
        {"previous.theorem.v2", false}, {"other.version.v1", false}, {"", false},
    } {
        if got:=compatibleTheoremSchema(tc.value); got!=tc.accepted {
            t.Errorf("compatible(%q)=%v; want %v", tc.value,got,tc.accepted)
        }
    }
}
