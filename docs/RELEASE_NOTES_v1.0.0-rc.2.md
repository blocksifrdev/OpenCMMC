# OpenCMMC v1.0.0-rc.2 | Single-Name Release Candidate

**Status:** Public prerelease candidate, not v1.0 general availability.

## What's new
- The sole public product name, CLI executable, archive filename, Go module, browser identity, and package metadata are OpenCMMC.
- CLI source is located at `cmd/opencmmc`.
- Legacy command aliases and retired preview publishing workflow were removed from the active source tree.
- Version-1 assessment/workspace structure remains readable across namespace changes; output uses OpenCMMC schema identifiers.
- Windows ARM64, Windows x64, Linux ARM64, and Linux x64 archives are built and smoke-tested before publication.
- SHA-256 file manifests and an SPDX module inventory accompany the download.

## Installation
Download the appropriate `opencmmc-v1.0.0-rc.2` archive, verify it with `SHA256SUMS`, extract it and run `START-CMMC-UI.cmd` (Windows) or `./start-cmmc-ui.sh` (Linux).

The new package contains **only** `opencmmc` as its CLI executable. Update local shell scripts that invoke an older executable name before using this release. Previously published archives are not modified.

## Scope and limitations
OpenCMMC provides local readiness and evidence workspaces, not government certification, an accredited assessment, or a validated SPRS submission. Official objective-source mapping requires independent review. The software does not supply fully validated enterprise SSP/POA&M, multi-user security boundaries or application-level storage encryption. Do not handle CUI without organization-specific approval.

This release is conditioned on successful pipeline checks and has no implied government endorsement.
