# FreeCompliance v0.5.0 Preview 2

**Updated portable offline packages. Official Veterans Day service launch: November 11, 2026. Access is available now.**

This prerelease supersedes Preview 1 and includes the previously source-only CMMC objective importer and draft score calculator. The current feature set is a **readiness workbench**. It is not official CMMC certification, government acceptance, accredited assessment, or validated SPRS submission.

## What's new

- `freecompliance cmmc objectives-import`: read the independently obtained official June 2018 SP 800-171A supplemental CSV and create a source-hashed offline objective workspace. The CI importer test covered 110 requirement identifiers and 320 objective statements, but **independent reconciliation to the normative PDF remains outstanding**.
- `freecompliance cmmc objectives-assess`: review source-derived objectives and evidence inside the customer-controlled enclave. Unresolved, altered or missing observations are not marked compliant.
- `freecompliance cmmc score-template`: generate an initially unassessed worksheet for 110 requirements.
- `freecompliance cmmc score-draft`: calculate **draft-only** arithmetic from complete human-entered findings and confirmed SSP metadata, with explicit checks for the limited 3.5.3 / 3.13.11 partial cases. Official score accuracy remains subject to independent validation.

## Downloads

| Device | Archive |
| --- | --- |
| Windows x64 | `freecompliance-v0.5.0-preview.2-windows-amd64.zip` |
| Windows ARM64 | `freecompliance-v0.5.0-preview.2-windows-arm64.zip` |
| Linux x64 | `freecompliance-v0.5.0-preview.2-linux-amd64.tar.gz` |
| Linux ARM64 | `freecompliance-v0.5.0-preview.2-linux-arm64.tar.gz` |

Includes platform executable, offline CMMC demo launcher, local documentation and SHA256SUMS. No Go installation, internet connection, account, LLM, telemetry or subscription required at runtime for core commands.

## Verified release smoke tests

The release workflow runs Go tests, vet, affected-package race detection, checksum verification, and a completely offline initial CMMC workspace/objective/score-template demonstration. It also verifies incomplete scoring **fails closed**.

## Restricted environments

An archive SHA-256 checksum establishes byte integrity against a published manifest, but not publisher authenticity. These archives **are not publisher-signed** with a pretrusted offline key. Government and contractor organizations must follow their own software approval, transfer, and CUI handling controls. Never circumvent endpoint protection or air-gap procedures.

The importer requires the official CSV to be separately acquired and verified by the user in an approved staging process. It does not distribute the normative NIST PDF. See `CMMC_ASSESSMENT_NEXT.md` inside the download.

**Still not included:** independently reviewed complete assessment methodology, authoritative control effectiveness testing across customer systems, validated SPRS score, official submission, C3PAO accreditation, OSCAL schema-validated assessment export, government endorsement, or graphical installer.

[Public technical documentation](https://github.com/blocksifrdev/freecompliance/blob/main/docs/CMMC_ASSESSMENT_NEXT.md) | [Air-gapped installation](https://github.com/blocksifrdev/freecompliance/blob/main/docs/DOWNLOADS.md) | [Open source repository](https://github.com/blocksifrdev/freecompliance)
