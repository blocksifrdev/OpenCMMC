# Enterprise and Government evaluation readiness: evidence gates

**Current assessment as of November 2, 2026:** **8/10 internal technical evaluation gates verified** against release Preview 4 and CI run [41532097431](https://github.com/blocksifrdev/freecompliance/actions/runs/41532097431). This is a software-engineering evaluation rubric, **not** an 8/10 federal/government production approval or CMMC compliance score. The two independent-assessment gates remain open. See [P0 issue #24](https://github.com/blocksifrdev/freecompliance/issues/24).

| # | Gate | Acceptance evidence | Verified status (Preview 4) |
| --- | --- | --- | --- |
| 1 | Public, inspectable, permanent free core | Apache-2.0 source, LICENSE, public repo | Verified |
| 2 | Portable deployable downloads | Windows/Linux amd64/arm64 GitHub Release, checksum manifest | Verified |
| 3 | Offline baseline functionality | No cloud needed for local CMMC workbench; isolated smoke demonstration | Verified in CI with Docker --network none; actual customer enclave test pending |
| 4 | Locally usable interface | Loopback UI creates and reviews CMMC workspaces | CI-verified |
| 5 | Evidence fail-closed semantics | Missing objective/evidence remains unresolved; incomplete score rejected | CI-verified |
| 6 | Engineering regression gates | Tests, race detection, vet and release CI green | CI-verified |
| 7 | Publisher provenance | Actual release archives bound to signed GitHub/Sigstore build attestations and successful verification | Verified for Preview 4, run 41532097431 |
| 8 | Machine-readable composition inventory | Version/commit-linked SPDX module inventory packaged, hash-checked | Not delivered in Preview 3 |
| 9 | Independent adversarial deployment/security evaluation | Signed independent test report, findings remediated and rerun in isolated VM | Not verified |
| 10 | Qualified independent CMMC assessment review | NIST normative PDF objectives and CMMC scoring independently reconciled, with reviewer attribution | Not verified |

**Quantitative interpretation:** The first eight engineering gates have project CI and release evidence. Gate 3 includes network-disabled container testing, **not** a physically air-gapped/customer-authorized installation. Gates 3-8 are validated by the project's own automated checks only. Independent external validation is needed for stronger government/customer assurance. An eight-of-ten engineering checklist score does **not** mean "80% CMMC compliant", "8/10 government approved", approved for CUI, or ready for production use by an agency.

## Required before treating the product as a government-ready solution

- Obtain organization-specific software approval, CUI enclave system authorization, incident response and data retention policies.
- Confirm current contract and CMMC program baseline and have a qualified independent expert reconcile NIST SP 800-171A assessment objectives and assessment guide semantics.
- Run an independent penetration/security review of the loopback browser and evidence handling, including negative-path tests, OS user isolation, removable-media, file symlink/TOCTOU handling, backups and malware protection.
- Deploy supported automatic or customer-managed software update verification and documented support channel with a tracked vulnerability disclosure response.
- Document government submission boundaries; never claim an SPRS API integration without authorized credentials and a supported interface.
- Establish signed, auditable release provenance, an inventory suitable for customer intake, and test reproducibility independently.
- Obtain government endorsement/authorization or C3PAO accreditation **only through the applicable external official process**, never by self-declaration.

## Truthful market language

Acceptable: "FreeCompliance is free, open source, offline-first CMMC readiness software, with downloadable Windows/Linux browser workbench and an evidence review workflow."

Not acceptable: "Government-approved", "CMMC certified", "8/10 government compliance", "official SPRS ready", "secure enough for CUI", "NIST certified", or "the software tests every control automatically".

Links: [GitHub release](https://github.com/blocksifrdev/freecompliance/releases) | [Release verification](RELEASE_INTEGRITY.md) | [Security policy](../SECURITY.md) | [NIST assessment source importer](CMMC_ASSESSMENT_NEXT.md).
