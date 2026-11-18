# OpenCMMC v1.0.0-rc.10 | Crash-Safe Assessment Recovery

**Status:** Free, open-source prerelease for local CMMC readiness. Not a
general-availability certification product, validated SPRS scoring tool,
official assessment, authorized government submission, or CUI security boundary.

## What changed

- Preserve a valid Level 1 or Level 2 assessment when OpenCMMC stops between
  writing its workspace file and writing `state.json`. The browser can open
  and export the original, and a second create attempt will not overwrite it.
- Detect unsafe, corrupt, and ambiguous orphaned workspace files and fail
  closed. The app never guesses which workspace to overwrite.
- Add automated recovery tests for Level 1 and Level 2 as well as corruption
  and ambiguity negative cases.
- Add [manual backup/recovery guidance](DATA_RECOVERY.md) for local offline
  evidence; encryption, backup, and recovery remain the organization's duty.
- Retain the rc.9 ChromeOS origin metadata fix and privacy-safe diagnostics
  without relaxing local loopback session or CSRF requirements.
- Update regulatory documentation to reflect the July 13, 2026 CMMC Phase II
  suspension. Current DoD guidance retains NIST SP 800-171 Revision 2
  self-assessment obligations and select government-led assessments. Confirm
  the current program requirements and each contract with DoD.

## Installation and verification

Download the Windows x64/ARM64 or Linux x64/ARM64 portable archive. Verify
its SHA-256 against the release manifest and your approved publisher-provenance
policy. Extract it into a separate release directory; preserve the complete
`private-assessment` folder and verify `opencmmc version` reports
`v1.0.0-rc.10` before opening the local private URL.

Keep session URLs, request headers containing secrets, cookies, evidence, and
CUI out of public logs and GitHub issues. Do not bypass endpoint security
controls.

## Unresolved release gates

- A successful form submission on the customer's real Chromebook has not yet
  been independently confirmed. Issue #49 remains open.
- Independent security and CMMC assessment-methodology evaluations remain
  necessary before any broader production compliance claims.
- Local evidence is not application-encrypted. Use approved encrypted storage,
  access control, backups, retention and organizational authorization.
- The optional paid BlockSiFr service links and local request downloads are
  not active scans, policy enforcement or authenticated service integrations.

See [current gate register](RELEASE_GATES_2026-11-11.md) and
[installation guidance](INSTALL_V1.md).
