# OpenCMMC v1.0.0-rc.1 | Veterans Day 2026 Release Candidate

**Status: Public prerelease, not v1.0 general availability.**

On Veterans Day 2026, BlockSiFr is publishing the OpenCMMC-branded candidate of its free, Apache-2.0, local-first CMMC evidence-readiness toolkit.

## What is included

- OpenCMMC-branded command-line utility and loopback-only browser UI.
- Offline Level 1 and Level 2 readiness workspaces, local evidence references and SHA-256 integrity checks.
- Separately sourced NIST SP 800-171A CSV assessment-objective ingestion and draft-only scoring utilities.
- Portable Windows ARM64, Windows x64, Linux ARM64, and Linux x64 packages.
- This initial rename-stage candidate contained transitional command aliases. Future OpenCMMC packages use one executable name.
- SHA-256 manifests and SPDX Go module inventory.

## Validation basis

This prerelease is published only when CI package generation, unit tests, race checks, static analysis, checksum validation, and native-architecture smoke tests pass. Test evidence is in the GitHub Actions run for the release branch. Tests exercise packaged binaries, workspace initialization and readiness export. These results do not establish comprehensive software security, government approval or CMMC certification.

## Limitations

This is a **candidate**, not a feature-complete enterprise CMMC assessment platform. It does not claim complete SSP/POA&M workflows, enterprise multi-user isolation, encrypted-at-rest evidence storage, independently reconciled current official assessment methodology, approved OSCAL export, validated SPRS scoring, automatic official submission, accredited C3PAO assessment or government endorsement. Evidence and reviewer assertions require independent evaluation.

The existing `v0.5.0-preview.4` artifact names and schema identities remain unchanged for existing users.

Read [installation instructions](docs/INSTALL_V1.md) and [current release gates](docs/RELEASE_GATES_2026-11-11.md). Do not bypass organizational software deployment or CUI security controls.
