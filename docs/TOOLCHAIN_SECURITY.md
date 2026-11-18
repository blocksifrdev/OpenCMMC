# Compiler and runtime security baseline

As of November 18, 2026, Go 1.23 is past its upstream security support
window. Go 1.27 and 1.26 are the actively patched release lines. Go 1.27.4
was released on November 10, 2026 with security fixes.

All future OpenCMMC binaries must be **compiled using a supported patched
Go toolchain**. The `go.mod` minimum compiler version is now `go 1.27.4`.
Native Windows x64 and ARM64, Linux x64 and ARM64, Chromium browser
acceptance, unit/race tests, and packaging must be rerun after the upgrade.
The version string of the shipped binary alone is not a security attestation;
inspect the Go binary build metadata with `go version -m <binary>`.

A standalone no-network OpenCMMC assessment requires no Go installation on
the customer's computer. CI may fetch Go and the advisory database. The
application does not perform vulnerability scanning or network telemetry.

References:
- https://go.dev/doc/devel/release
- https://go.dev/doc/go1.27

Fail release for any reachable vulnerabilities reported by an up-to-date
`govulncheck` run, and record the output or reviewed exception in the
release evidence. Static analysis is not a penetration test and does not
approve handling CUI.
