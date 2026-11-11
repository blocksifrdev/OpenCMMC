package cmmc

import "strings"

// Earlier release workspaces keep the same v1 structure but may use an older
// namespace. Match the exact version and structure suffix, then apply normal
// control, baseline, and evidence validation.
func compatibleCMMCSchema(actual, expected string) bool {
    if actual == expected { return true }
    prefix, suffix, ok := strings.Cut(actual, ".")
    _, expectedSuffix, expectedOK := strings.Cut(expected, ".")
    if !ok || !expectedOK || suffix != expectedSuffix || len(prefix) < 4 || len(prefix) > 32 { return false }
    for _, c := range prefix { if c < 'a' || c > 'z' { return false } }
    return true
}
