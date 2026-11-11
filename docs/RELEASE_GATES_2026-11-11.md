# OpenCMMC Veterans Day 2026 release gate register

**Date:** November 11, 2026. This register documents required evidence, not completed certification.

| Gate | Evidence | Status |
| --- | --- | --- |
| Branding | CLI, browser, launch scripts, packages | PASS in [run #43064896487](https://github.com/blocksifrdev/OpenCMMC/actions/runs/43064896487) |
| Windows ARM64 | Native package plus browser CSRF regression | PASS in [run #43064896487](https://github.com/blocksifrdev/OpenCMMC/actions/runs/43064896487) |
| Windows x64 | Native package and assessment report | PASS in [run #43064896487](https://github.com/blocksifrdev/OpenCMMC/actions/runs/43064896487) |
| Linux ARM64 | Native package and assessment report | PASS in [run #43064896487](https://github.com/blocksifrdev/OpenCMMC/actions/runs/43064896487) |
| Linux x64 | Native package and assessment report | PASS in [run #43064896487](https://github.com/blocksifrdev/OpenCMMC/actions/runs/43064896487) |
| Security | Automated tests and independent review | Independent review outstanding |
| CMMC requirements | Official source reconciliation; correct Phase I status | Independent review outstanding |
| Full product workflows | SSP, POA&M, organizational controls and backup if claimed | Not independently verified |
| Integrity | Checksums, origin provenance and approval | PASS SHA-256/SPDX in [run #43064896487](https://github.com/blocksifrdev/OpenCMMC/actions/runs/43064896487); independent publisher identity pending |
| Publication | Approved GA tag, release and postpublish smoke | PASS prerelease published; GA not complete |

The active Go module, schema identifiers, command names, and package filenames use OpenCMMC. Readers still accept earlier version-1 workspace structures by matching their validated schema suffix and checking the full assessment content. Previously published release artifacts and Git history are immutable.

Do not claim government endorsement, certification, C3PAO accreditation, validated SPRS scoring or official submissions. Do not publish a general-availability claim until its advertised capabilities have been independently verified.

**November 11 update:** The prerelease branch published [v1.0.0-rc.1](https://github.com/blocksifrdev/OpenCMMC/releases/tag/v1.0.0-rc.1). All native package smoke tests passed. Independent CMMC methodology review, end-to-end feature completeness, and broader security assurance remain outstanding.
