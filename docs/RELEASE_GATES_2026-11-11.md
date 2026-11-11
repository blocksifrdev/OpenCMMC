# OpenCMMC Veterans Day 2026 release gate register

**Date:** November 11, 2026. This register documents required evidence, not completed certification.

| Gate | Evidence | Status |
| --- | --- | --- |
| Branding | CLI, browser, launch scripts, packages | Candidate CI pending |
| Windows ARM64 | Native package plus browser CSRF regression | Candidate CI pending |
| Windows x64 | Native package and assessment report | Candidate CI pending |
| Linux ARM64 | Native package and assessment report | Candidate CI pending |
| Linux x64 | Native package and assessment report | Candidate CI pending |
| Security | Automated tests and independent review | Independent review outstanding |
| CMMC requirements | Official source reconciliation; correct Phase I status | Independent review outstanding |
| Full product workflows | SSP, POA&M, organizational controls and backup if claimed | Not independently verified |
| Integrity | Checksums, origin provenance and approval | Candidate validation only |
| Publication | Approved GA tag, release and postpublish smoke | Not complete |

The historical Go module import path, persistent `freecompliance.*` schema identifiers and `fc_session` cookie are preserved for backward compatibility. The outward product name is OpenCMMC.

Do not claim government endorsement, certification, C3PAO accreditation, validated SPRS scoring or official submissions. Do not publish a general-availability claim until its advertised capabilities have been independently verified.
