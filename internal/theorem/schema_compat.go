package theorem

import "strings"

// v1 models saved before the namespace was standardized remain readable.
// All other model content is subject to the existing strict validator.
func compatibleTheoremSchema(actual string) bool {
    if actual == Schema { return true }
    prefix, suffix, ok := strings.Cut(actual, ".")
    if !ok || suffix != "theorem.v1" || len(prefix) < 4 || len(prefix) > 32 { return false }
    for _, c := range prefix { if c < 'a' || c > 'z' { return false } }
    return true
}
