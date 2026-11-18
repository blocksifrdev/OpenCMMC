# OpenCMMC v1.0.0 release rehearsal (no publication)

This is a **production-artifact rehearsal**, not a general-availability release.

The branch `ga/validation/v1.0.0-20261118` triggers the native release
workflow with **the exact `v1.0.0` version string**. It does not create a
public GitHub Release or tag. The existing prerelease publication condition
remains restricted to the retired `release/v1.0.0-rc.11` branch.

This rehearsal requires:
- A supported Go 1.27.4 compiler, `go test`, race tests and vet.
- `govulncheck` with no reachable disclosed vulnerabilities.
- Portable Windows x64/ARM64 and Linux x64/ARM64 builds with exact
  `OpenCMMC v1.0.0` executable-version assertions.
- Level 1/2 synthetic assessment and safe backup/restore testing against the
  actual packaged executables.
- A Chromium browser assessment, file upload, export and forwarded-origin
  security regression.
- SPDX software bill of materials, SHA-256 file verification, Sigstore
  build provenance and SBOM attestations.
- **Independent `gh attestation verify` of all four archives against the
  expected GitHub repository identity using both stored signature bundles.**

A successful rehearsal means the v1.0.0 packaging and security pipeline is
technically sound. It does **not** mean:
- An independent security penetration test occurred.
- A qualified assessor reconciled CMMC objectives, methods and scoring.
- The actual Chromebook or offline pilot passed acceptance.
- A repository owner signed an authorization to release.
- A `v1.0.0` tag or public GA binary was published.

The release is **NO-GO** until the applicable external evidence in
[GA_ACCEPTANCE_PROTOCOL.md](GA_ACCEPTANCE_PROTOCOL.md) is signed and a
protected, deliberately authorized publication workflow rechecks the
final SHA, signatures and public artifacts.

## Source and artifact identity

Rehearsal CI archives are temporary CI artifacts. Their signatures bind to
the **rehearsal commit**, not to a future production release tag. Never
relabel a rehearsal archive as a published GA build.

Only use synthetic, non-CUI data in rehearsal tests.
