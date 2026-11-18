# OpenCMMC v1.0.0 | Offline CMMC Readiness Workbench

**This text applies only to an authorized, signed public v1.0.0 GA release
after the independent acceptance gates have passed. Until then, v1.0.0 is
a production release objective and this is draft release documentation.**

## Intended functionality

- Free/open-source, local-first CMMC readiness workbench.
- 15 Level 1 safeguarding requirements and 110 Level 2 Rev 2 requirement
  identifiers, with all unsupported claims and missing objectives unresolved.
- Local browser assessment flow, scoped reviewer observations, evidence
  references and explicit gap/unknown statuses, JSON export.
- Optional source-verified objective import, clearly labeled draft
  scoring, offline read-only CLI capabilities and opt-in service links.
- Offline evidence backup, SHA-256 verification and new-directory restore.
  Backup ZIP files are **plaintext**: approved encrypted device storage and
  organizational CUI handling remain required.
- No remote telemetry, cloud sign-in or mandatory network service for
  the core local assessment workflow.

## Safety and assurance

This software does **not** certify CMMC compliance, determine technical
control effectiveness from a document hash, submit an official SPRS score,
grant C3PAO accreditation or government authorization, or authorize
storage of CUI.

The approved GA workflow must rebuild this exact tagged source, run Go
and Chromium checks, scan reachable Go vulnerabilities, run native
Windows/Linux x64/ARM64 package tests, and publish independent signed
GitHub/Sigstore build provenance and SBOM bundles. Before promoting the
release from draft, an accountable release owner must obtain independently
signed security, CMMC methodology, Chromebook, air-gap, accessibility and
release authorizations, all matching the exact source commit and archive
digests.

Only trust a GA release whose published archive hashes and signed bundle
identity have been independently verified against the canonical
`blocksifrdev/OpenCMMC` repository.

See [GA acceptance protocol](https://github.com/blocksifrdev/OpenCMMC/blob/main/docs/GA_ACCEPTANCE_PROTOCOL.md),
[signed approval requirements](https://github.com/blocksifrdev/OpenCMMC/blob/main/docs/GA_SIGNED_APPROVALS.md),
and [security policy](https://github.com/blocksifrdev/OpenCMMC/blob/main/SECURITY.md).
