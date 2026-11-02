# CMMC proof and assessment roadmap

**Baseline date:** November 2, 2026

## Purpose

Make FreeCompliance useful as a complete **offline assessment workbench** while maintaining an explicit boundary between free software and government or third-party authorization. The roadmap below is a development plan, not an announcement that the milestones have shipped.

## Current demonstrated scope

- Level 1 inventory of 15 safeguarding item identifiers.
- Level 2 inventory of 110 NIST SP 800-171 Revision 2 requirement identifiers.
- Evidence references, integrity and freshness assessment, source isolation, and JSON reporting.
- Detached Ed25519 bundle signature validation using pretrusted public keys.
- Separate experimental theorem engine for bounded user-supplied models.
- None of these functions produce a valid SPRS score, assessor accreditation, or official certification.

## Milestone A: Complete version-pinned objective library

- Generate a curated objective index from the official NIST SP 800-171A assessment procedures and the applicable government CMMC program baseline.
- Preserve exact publication title, edition, revision, date, source URI, extraction process, and source digest.
- Represent objective IDs, assessment methods (examine, interview, test), assessment objects, parameters, scope, exceptions, and evidence requirements.
- Preserve source documents and provenance; follow applicable copyright and distribution requirements.
- Validate objective coverage against official published counts and independently review all mappings.
- Never mark a requirement satisfied if some applicable objectives remain unresolved.

**Acceptance:** independently reconciled objectives, repeatable import, schema tests, zero silently omitted objectives.

## Milestone B: Actual control effectiveness

- Add local read-only adapters for approved Windows, Linux, directory, cloud configuration export, identity, and security platforms.
- All adapters must operate from offline evidence snapshots, and optional connected collectors must be excluded from physical air-gap binaries.
- Define deterministic pass/fail/unknown conditions for precisely scoped predicates.
- Record collection method, identity, time, artifact fingerprint, reviewer, and authorization. Require human review for interview and procedural evidence.
- Add mutation/negative-path test fixtures and independent reviewer adjudication.
- Keep testing non-invasive and separately approved; no network probing or active exploitation by default.

**Acceptance:** verified applicability and strong negative tests for a defined initial set; no automatic promotion from document presence to effectiveness.

## Milestone C: Score readiness and SPRS handoff

- Version-pin the official DoD NIST SP 800-171 assessment methodology and evaluation scope.
- Model system security plan boundaries, applicable requirements, deficiencies, and customer-assessed determinations.
- Provide a **draft** score and per-item calculation trace only after official methodology validation and human review.
- Test against government-published examples and independent qualified assessors.
- Output a user-reviewable submission preparation package; do not automate SPRS submission without approved authority and an explicitly supported government interface.

**Acceptance:** zero disagreement on reviewed published score fixtures and explicit decision-maker signoff. Until achieved, no numeric SPRS score.

## Milestone D: Portable assessor package

- Version-pinned OSCAL assessment results, plans, findings, evidence references, and POA&M interoperability.
- Validate generated JSON/XML against the exact NIST OSCAL model and version-specific schemas.
- Include redaction, encryption-at-rest guidance, signatures, reviewer history, export approval, and reproducible provenance.
- Describe outputs as **schema-validated OSCAL exports**, not "OSCAL-certified" unless an actual official certification program and approval exist.

**Acceptance:** schema validation and independent system-to-system import tests on redacted fixtures.

## Milestone E: Independent assessment and government adoption

- Invite CMMC practitioners, qualified assessors, defense suppliers, and government technical reviewers to evaluate correctness and usability.
- C3PAO authorization/accreditation is an organizational designation under the applicable program, not a toggle in software.
- Government procurement, authority to operate, CUI boundary approvals, and deployment decisions require separate official processes.
- Do not imply official use, endorsement, accreditation, or accepted donation without records establishing it.

## Delivery and risk gates

Each release requires Go unit tests, race tests, static analysis, adversarial test fixtures, documented source lineage, reproducible sample output, and clear warnings for unsupported claims. Air-gapped transfer requires separate key trust and approved media handling. Hold release if any evidence path can falsely elevate missing or contradictory observations to positive control findings.

## Relevant sources

- https://dodcio.defense.gov/cmmc/About/
- https://csrc.nist.gov/pubs/sp/800/171/a/final
- https://www.dcma.mil/DIBCAC/
- https://pages.nist.gov/OSCAL/
