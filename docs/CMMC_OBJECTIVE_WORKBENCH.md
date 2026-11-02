# CMMC objective-level workbench | experimental offline release

**Purpose:** inspect the evidence for individual declared CMMC assessment objectives within a customer-controlled system boundary. Works without internet or any cloud connection.

**Supported today:** offline objective workspace, reviewer assertions, examine/interview/test method labels, evidence SHA-256 and timestamp checks, conservative rollups of *listed* objectives, JSON report. This release DOES NOT provide the complete independently reviewed SP 800-171A objective catalog, an assessor determination, a validated SPRS score, or CMMC certification.

## Choose the relevant baseline

The 2018 NIST SP 800-171A publication is marked withdrawn by NIST because Rev. 3 superseded it in 2024. It remains relevant to the CMMC Level 2 NIST SP 800-171 Rev. 2 baseline, subject to the customer's current contract and official program direction. Never mix Rev. 3 objective identifiers into a Rev. 2 workspace without an explicitly reviewed crosswalk.

- CMMC program: https://business.defense.gov/cmmc/
- Official legacy assessment source: https://csrc.nist.gov/pubs/sp/800/171/a/final
- Official NIST publication PDF is normative; the supplemental CSV/XLSX files can help create a versioned catalog but must be reconciled to it.

## Run offline

Build with Go 1.23+ in a customer-approved environment. For an air gap, transfer using an approved, validated software process.

```sh
./freecompliance cmmc objectives-init --level 2 --scope "CUI enclave A" --out objectives.json
mkdir -p evidence
./freecompliance cmmc objectives-assess --input objectives.json --evidence-dir evidence --json objective-report.json
```

The first command generates a 110-control skeleton with **empty** objective lists. A reviewer must populate each control's objective list using the appropriate authoritative procedure source and retain a source/version record. The tool intentionally does not invent objectives to make a checklist look complete.

### Objective entry example (synthetic only)

For a single control in `objectives.json`, the `objectives` array can contain:

```json
{
  "id": "3.1.1[a]",
  "decision": "ASSERTED_MET",
  "method": "EXAMINE",
  "reviewer": "sample reviewer",
  "rationale": "Synthetic example only. Reviewer examined an approved local file.",
  "evidence": [{
    "path": "review.txt",
    "sha256": "<64 lowercase hex digits for actual file>",
    "captured_at": "2026-11-02T15:00:00Z"
  }]
}
```

The ID is illustrative of the format, **not validation that it appears in the precise referenced government procedure**. Add no restricted evidence, CUI, secrets, or real customer metadata to a public issue. `method` can be `EXAMINE`, `INTERVIEW`, or `TEST`, but method labels are reviewer assertions rather than automated execution of those procedures.

## Interpretation

| Report status | Meaning |
| --- | --- |
| `OBJECTIVE_LIST_MISSING` | No objectives supplied for a control. Cannot infer compliance. |
| `UNRESOLVED` | Some listed objectives are unassessed, out of scope pending review, or backed by unverifiable evidence. |
| `REVIEWER_REPORTED_GAP` | A human reviewer recorded that at least one listed objective was not met. |
| `EVIDENCE_REVIEW_READY` | Each **listed** objective is asserted met, with locally verified file integrity and timestamp. Does **not** mean objective catalog complete, control effective, or a CMMC pass. |

All reports say `NOT_INDEPENDENTLY_VALIDATED` for catalog completeness, `NOT_INDEPENDENTLY_VERIFIED` for effectiveness, `NOT_CALCULATED` for SPRS, and `NOT_PERFORMED` for government submission. `procedure_sha256` is user-supplied reference metadata, not a trusted certification or verified signature. An independently reconciled, version-pinned official objective catalog is a separate milestone.

`--max-age-days` defaults to 90 as **local policy**, not an official CMMC evidence recency requirement. Outputs may expose scope names, file paths, hashes, and reviewer identifiers; follow enclave/CUI procedures.

## Acceptance before an official assessment claim

1. Independently reconcile every objective ID against the authoritative 2018 SP 800-171A PDF and the current CMMC program baseline; test for missing and duplicate objectives.
2. Import assessment objects and methods with official source lineage, provenance hashes, and meaningful validity conditions.
3. Validate applicable exclusions with the proper assessor and contractual boundary.
4. Independently test objective effectiveness, negative paths, and bypass behavior. A document or file hash alone is insufficient.
5. Reconcile draft scoring against the official, current government methodology, and retain the authorized assessor's decisions.
6. Schema-validate OSCAL exports and independently test import/export before claiming interoperability.
7. Keep SPRS submission, accreditation, and government endorsements separate from software output.

See [the CMMC program roadmap](CMMC_PROOF_ROADMAP.md) and [air-gapped deployment guide](CMMC_AIRGAP.md).