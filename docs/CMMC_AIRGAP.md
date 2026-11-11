# CMMC offline and air-gapped quick start

OpenCMMC is an open-source **evidence readiness** workbench. It is not an official CMMC assessment, an SPRS scoring or submission tool, an approved CUI handling system, or a substitute for a qualified assessor.

## One-time preparation

1. Identify the actual system or enclave scope and responsible owners. Keep system security plans, customer data, and CUI within your approved boundary.
2. In an approved build/staging environment, obtain the public source from `https://github.com/blocksifrdev/OpenCMMC`, inspect it, run `go test ./...` and `go vet ./...`, and build the binary with `go build -o opencmmc ./cmd/opencmmc`.
3. Inspect and approve the binary under your organization's software supply-chain procedures. Transfer only through authorized media and processes; verify bytes and provenance on the destination. Do not presume a signed bundle is trustworthy unless its public key was provisioned and verified out of band.
4. No runtime network connection is needed for the `cmmc`, `scan`, `compare`, or `prove` commands. The `github` and `prove-github` commands require internet and are not appropriate for a truly air-gapped installation.

## Run completely offline

```bash
./opencmmc cmmc init --level 2 --scope "Approved CUI enclave" --out workspace.json
mkdir -p evidence
./opencmmc cmmc assess --input workspace.json --evidence-dir evidence --json readiness.json
```

The initialized workspace contains **110 requirement identifiers** from NIST SP 800-171 Revision 2. For Level 1 use `--level 1` to initialize **15 FAR 52.204-21 safeguarding item identifiers**. All start `UNASSESSED`. The catalog intentionally does not reproduce licensed or authoritative control content, assessment objectives, or make applicability judgments.

## How evidence is represented

A reviewer editing `workspace.json` may set a control's `claim` to `IMPLEMENTED`, identify themselves with `reviewer`, and reference an evidence file stored **under** `--evidence-dir`. Each `evidence` entry needs a relative `path`, lowercase SHA-256 `sha256`, and RFC3339 `captured_at` timestamp. A matching SHA-256 digest, acceptable file age, and file safety checks result in `DOCUMENTED` for that item. That status is **only evidence integrity with a human claim**, not proof of operating effectiveness.

Control claims:

| Input claim | Result | Interpretation |
| --- | --- | --- |
| `UNASSESSED` | `UNASSESSED` | No determination or evidence review has been recorded. |
| `GAP` | `GAP` | Reviewer reports an unresolved requirement. |
| `IMPLEMENTED` with reviewer and valid linked evidence | `DOCUMENTED` | File integrity and recency match; control operation remains unproven. |
| `IMPLEMENTED` with missing, stale, or changed evidence | `UNVERIFIABLE` | Do not infer that the requirement is satisfied. |
| `NOT_APPLICABLE` with reviewer and rationale | `SCOPE_REVIEW_REQUIRED` | An authorized assessor must independently validate exclusion. |

Default maximum age is 90 days, configurable with `--max-age-days`. This is a **local policy parameter**, not an official CMMC recency rule. Evidence paths are checked to avoid directory traversal and symlinks, and files have a size limit. The tool does not execute evidence files.

## Update and export policy

- To verify a detached Ed25519 signature, run `cmmc verify-bundle --bundle <file> --signature <base64-signature-file> --public-key <base64-key-file>`. Trust in the public key is an **out-of-band decision**.
- The signature verification command validates the supplied bytes; it does not install, parse, or approve the update bundle.
- Store `workspace.json`, evidence files, `readiness.json`, and reports inside the customer-approved security boundary. Local report outputs can expose names, scopes, paths, hashes, and reviewer metadata.
- No report is uploaded by the CMMC CLI, no telemetry is sent, and **no SPRS submission is made**. Export only after the organization's authorized CUI and security review.
- Ensure an independently reviewed SSP, assessment procedures, scope and contract requirements. Consult official CMMC requirements and the current acquisition clauses.

## Where this product stands

Current: offline evidence tracking, 15/110 identifier inventories, SHA-256 integrity and freshness checks, signature validation, explicit unresolved states, portable JSON report.

Not currently provided: validated CMMC scores, objective-by-objective assessment, complete assessor workflows, direct SCAP/OSCAL results certification, automated POA&M governance, evidence authenticity attestation, proof of operational effectiveness or government adoption. The separate experimental theorem engine is not automatically mapped to these CMMC requirements.

Official guidance: https://business.defense.gov/cmmc/ and https://csrc.nist.gov/pubs/sp/800/171/r2/upd1/final.
