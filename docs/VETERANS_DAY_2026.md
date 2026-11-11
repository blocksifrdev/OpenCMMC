# Veterans Day 2026 | A OpenCMMC Service Commitment

**November 11, 2026**

## Built in gratitude. Given in service.

This Veterans Day, OpenCMMC by BlockSiFr honors the men and women who have served in the United States Armed Forces. The initiative is personal to our founder, the son of a U.S. Army lieutenant colonel.

We are giving back by making essential defense cybersecurity readiness tools free, open source, and usable on disconnected systems. This is a permanent commitment, not a time-limited coupon or a promise of government adoption.

> **Stop guessing about CMMC. Start proving your controls.**
>
> Know what's required. See what's documented. Find what's missing.

## Who this is for

- Small and midsize defense suppliers that cannot justify another compliance subscription.
- Prime contractors and subcontractors operating FCI or CUI systems.
- Government cybersecurity and acquisition-support teams that need inspectable, portable tooling.
- Internal reviewers and authorized independent assessors who want reproducible evidence inventories.

## What you can use now

OpenCMMC is Apache 2.0 licensed and can run its core CMMC workbench without a cloud service, account, telemetry, or paid subscription:

| Function | Actual capability |
| --- | --- |
| CMMC Level 1 | Offline workspace with 15 FAR safeguarding item identifiers |
| CMMC Level 2 | Offline workspace with 110 NIST SP 800-171 Rev. 2 requirement identifiers |
| Evidence inventory | Local, per-item reviewer assertions and evidence references |
| Evidence integrity | SHA-256 file checks and configurable freshness checks |
| Assessment report | Portable JSON with documented, gap, unassessed, unverifiable, or scope-review-needed findings |
| Offline transfer check | Detached Ed25519 signature verification with a separately trusted public key |
| Theorem Intelligence | Experimental bounded proof evaluator; not yet a complete CMMC objective evaluator |

The identifier catalogs are **not** the complete CMMC assessment procedures. A file hash is evidence of integrity, not a finding of effective security control operation.

### Start without an account

Build from reviewed source in an approved environment:

```bash
git clone https://github.com/blocksifrdev/OpenCMMC.git
cd OpenCMMC
go test ./...
go build -o opencmmc ./cmd/opencmmc
./opencmmc cmmc init --level 2 --scope "Approved CUI enclave" --out workspace.json
mkdir -p evidence
./opencmmc cmmc assess --input workspace.json --evidence-dir evidence --json readiness.json
```

For an air-gapped environment, build and transfer only through your organization's approved software and removable-media processes. The source checkout requires a connected staging environment; the installed core binary does not. See [air-gapped deployment](CMMC_AIRGAP.md).

## Why now

On July 13, 2026, the Department announced suspension of CMMC Phase II implementation while keeping Phase I self-assessment obligations in place. Contractors should confirm applicable current contracts and directives rather than infer requirements from a software score.

Official sources:
- [DoD / DoW CMMC program](https://dodcio.defense.gov/cmmc/About/)
- [NIST SP 800-171A assessment procedures](https://csrc.nist.gov/pubs/sp/800/171/a/final)
- [DCMA DIBCAC assessment resources and SPRS links](https://www.dcma.mil/DIBCAC/)

## Our commitments

1. **Free forever:** Core open-source assessment, proof evaluation, and local evidence export will have no subscription requirement.
2. **Offline first:** No mandatory cloud, account, AI inference API, or outbound telemetry for core operations.
3. **Evidence over theater:** Do not present a listed document, passing hash, or modeled proof as a complete security assessment.
4. **Source transparency:** Public code, reproducible tests, documented assumptions, and versioned catalogs.
5. **Government interoperability:** Work toward validated NIST SP 800-171A objective mapping and OSCAL-compatible exports with independently checked schemas.
6. **No endorsement claims:** This is a private, independent service initiative available for government evaluation, not an accepted gift, procurement award, official endorsement, or deployment by any government organization.

## Next releases, subject to testing

See [CMMC capability roadmap](CMMC_PROOF_ROADMAP.md). We are building complete objective-level assessment workspaces, source-bound effectiveness tests, defensible draft score calculations, review workflows, and interoperable evidence packages. Those features must be independently validated before any public effectiveness or conformance claim.

**Official CMMC assessment results, C3PAO authorization or accreditation, and authorized SPRS submission cannot be conferred by software alone.** Government users must follow their own security authorization, acquisition, records, and CUI requirements.

## Dedication

To our veterans and military families, and to the contractors whose work sustains national defense: **Thank you for your service. This is ours.**

Copyright 2026 BlockSiFr. Released under [Apache 2.0](../LICENSE).