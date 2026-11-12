# OpenCMMC Veterans Day 2026 release gate register

**Date:** November 11, 2026. This register documents required evidence, not completed certification.

| Gate | Evidence | Status |
| --- | --- | --- |
| Branding | CLI, browser, launch scripts, packages | PASS in [run #43070771899](https://github.com/blocksifrdev/OpenCMMC/actions/runs/43070771899) |
| Windows ARM64 | Native package plus browser CSRF regression | PASS in [run #43070771899](https://github.com/blocksifrdev/OpenCMMC/actions/runs/43070771899) |
| Windows x64 | Native package and assessment report | PASS in [run #43070771899](https://github.com/blocksifrdev/OpenCMMC/actions/runs/43070771899) |
| Linux ARM64 | Native package and assessment report | PASS in [run #43070771899](https://github.com/blocksifrdev/OpenCMMC/actions/runs/43070771899) |
| Linux x64 | Native package and assessment report | PASS in [run #43070771899](https://github.com/blocksifrdev/OpenCMMC/actions/runs/43070771899) |
| Security | Automated tests and independent review | Independent review outstanding |
| CMMC requirements | Official source reconciliation; correct Phase I status | Independent review outstanding |
| Full product workflows | SSP, POA&M, organizational controls and backup if claimed | Not independently verified |
| Integrity | Checksums, origin provenance and approval | PASS SHA-256/SPDX in [run #43070771899](https://github.com/blocksifrdev/OpenCMMC/actions/runs/43070771899); independent publisher identity pending |
| Publication | Approved GA tag, release and postpublish smoke | PASS prerelease published; GA not complete |

The active Go module, schema identifiers, command names, and package filenames use OpenCMMC. Readers still accept earlier version-1 workspace structures by matching their validated schema suffix and checking the full assessment content. Previously published release artifacts and Git history are immutable.

Do not claim government endorsement, certification, C3PAO accreditation, validated SPRS scoring or official submissions. Do not publish a general-availability claim until its advertised capabilities have been independently verified.

**November 11 update:** The full rename release [v1.0.0-rc.2](https://github.com/blocksifrdev/OpenCMMC/releases/tag/v1.0.0-rc.2) passed the native platform build and smoke tests in [run #43070771899](https://github.com/blocksifrdev/OpenCMMC/actions/runs/43070771899). An additional opt-in services interface is undergoing verification for a subsequent release. Independent CMMC methodology review, end-to-end feature completeness, and broader security assurance remain outstanding.

**Optional service development:** OpenCMMC v1.0.0-rc.3 offered explicit public service links. The next candidate adds a local, consent-gated JSON service-request download for CortexTrace, Control and BLAKE. This does not transfer evidence, activate a paid service, launch a scan or perform runtime enforcement. Real authenticated handoffs remain pending. Branding and archive-integrity checks are enforced by the current CI and candidate-build workflows. This does not establish completion of broader customer integration or an independent security review.

**ChromeOS local-session repair (rc.5 candidate):** The browser bootstrapping cookie was changed from Strict to Lax to support top-level GET redirect from an external Linux VM browser launch. POSTs retain session, CSRF, and exact-Origin enforcement, and the server retains loopback-only Host validation. Automated bootstrap and hostile-origin regression tests cover the change. Chromebook integration feedback remains necessary.

**Local browser auth hotfix (rc.6 candidate):** Prior rc.5 cookie bootstrap improvements did not settle all ChromeOS-reported form failures. Two simultaneous local instances could previously share the `oc_session` host cookie even when listening on different ports. The candidate uses port-specific cookie names, a visible running-build version, and per-cause error identifiers without reflecting sensitive request data. Exact Origin checking, CSRF validation and loopback binding remain enforced. Native CI is not a substitute for confirmation on the user's Chromebook.
