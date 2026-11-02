# FreeCompliance v0.5.0 Preview 4 | Supply-chain hardening

**Release candidate:** Public preview. Official Veterans Day service initiative launch is November 11, 2026.

## This update

- Retains the existing offline CMMC browser workbench and all four portable platforms (Windows/Linux, amd64/arm64).
- Corrects stale and conflicting documentation for the Preview 3 interface.
- Adds a substantive security disclosure policy and enterprise evaluation checklist.
- Generates and publishes a source-commit-linked **SPDX 2.3 Go-module inventory**, with negative tests and an SHA-256 manifest entry.
- Creates **signed GitHub/Sigstore build provenance attestations** for each release archive and the software composition inventory, with online verification gating publication.

**Publication and provenance claims apply only if the actual Preview 4 GitHub Release exists and its attestation verifier passed.** Publishing these notes or landing the workflow does not constitute a signed release.

## What this does not grant

- Windows Authenticode code signing or an independently controlled offline release signing root.
- A government authorization to operate or authority to host CUI.
- An independently performed security assessment or penetration test.
- Complete official CMMC assessment procedure coverage, validated SPRS scoring/submission, third-party C3PAO accreditation, or government endorsement.

Go module SPDX metadata is a supply-chain *inventory*, not a vulnerability-free certificate, full product SBOM, or a NIST compliance assessment.

For approval and offline verification use [Release integrity guidance](https://github.com/blocksifrdev/freecompliance/blob/main/docs/RELEASE_INTEGRITY.md). Keep CUI within approved customer controls.

**Free forever. Built in gratitude. Given in service.**
