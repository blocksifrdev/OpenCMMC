# FreeCompliance

## Download FreeCompliance

**Windows and Linux portable preview packages:** [GitHub Releases](https://github.com/blocksifrdev/freecompliance/releases) | [Download and offline installation guide](docs/DOWNLOADS.md)

**Download now:** [Windows x64](https://github.com/blocksifrdev/freecompliance/releases/download/v0.5.0-preview.1/freecompliance-v0.5.0-preview.1-windows-amd64.zip) | [Windows ARM64](https://github.com/blocksifrdev/freecompliance/releases/download/v0.5.0-preview.1/freecompliance-v0.5.0-preview.1-windows-arm64.zip) | [Linux x64](https://github.com/blocksifrdev/freecompliance/releases/download/v0.5.0-preview.1/freecompliance-v0.5.0-preview.1-linux-amd64.tar.gz) | [Linux ARM64](https://github.com/blocksifrdev/freecompliance/releases/download/v0.5.0-preview.1/freecompliance-v0.5.0-preview.1-linux-arm64.tar.gz) | [SHA256SUMS](https://github.com/blocksifrdev/freecompliance/releases/download/v0.5.0-preview.1/SHA256SUMS)

No Go compiler is required to run the packaged executable. The Windows ZIP includes a double-click **offline demo launcher**. **The v0.5.0-preview.1 binaries predate the newly merged CMMC objective importer and draft scoring code.** For those commands, build the latest reviewed source from `main` until a new release is verified. Neither build is an independently validated SPRS scoring system, CMMC certification, or graphical installer. Follow your organization's software approval and CUI handling rules.

> **AVAILABLE NOW | OFFICIAL VETERANS DAY LAUNCH: NOVEMBER 11, 2026**
>
> Free, open-source CMMC readiness for defense contractors, assessors, and government technical teams. No subscription, login, or mandatory cloud service. Core tools work offline.
>
> **[Get started today](#get-started-in-three-steps)** | **[Read the November 2 early-access announcement](docs/AVAILABLE_NOW_2026_11_02.md)** | **[Veterans Day service commitment](docs/VETERANS_DAY_2026.md)**

FreeCompliance is **publicly accessible today, November 2, 2026**. Our formal Veterans Day launch is **Wednesday, November 11, 2026**. Early access is for use, testing, and community feedback, not a claim of completed CMMC assessment or government approval.

This initiative honors U.S. military service and is inspired by the founder's father, a lieutenant colonel in the U.S. Army. The core project is and will remain free and open source.

## Stop guessing about CMMC. Start proving your controls.

**Know what's required. See what's documented. Find what's missing.**

**FreeCompliance** is a free, open-source cybersecurity assurance toolkit for defense contractors, subcontractors, assessors, and government teams. Start a CMMC readiness workspace, organize evidence for each requirement, check whether files are intact and current, and create a report that clearly separates documented claims from gaps and unknowns.

**No subscription. No login. No required cloud service. Runs on an offline or air-gapped computer.**

> **What you get today:** An offline CMMC evidence workbench, source-code security checks, and experimental, bounded theorem evaluation. This is a readiness and evidence-integrity tool, **not an official CMMC assessment, SPRS submission, certification, or proof that a control operates effectively**.

### Who is this for?

| You are... | FreeCompliance helps you... |
| --- | --- |
| A small defense contractor | Start a no-cost CMMC evidence inventory without buying a SaaS platform. |
| A prime or subcontractor | Organize evidence within an explicitly named system or enclave. |
| A security or compliance lead | Identify undocumented, unresolved, or missing evidence and assign follow-up reviews. |
| An assessor or internal auditor | Review consistent, traceable files and see which conclusions require independent testing. |
| A restricted or disconnected organization | Run local assessments without sending CUI, source code, or evidence to BlockSiFr. |

## Complete NIST assessment-objective source import (new in source)

The latest Go source includes an **offline importer** for the official NIST SP 800-171A (June 2018) supplemental CSV. In GitHub CI, it imported all **110 Level 2 requirement identifiers and 320 objective statements**. This establishes validated CSV-format ingestion, **not** independent reconciliation against the normative NIST PDF or a passing CMMC assessment.

```bash
# First obtain the official CSV through an approved, independently verified staging process.
./freecompliance cmmc objectives-import --csv sp800-171a-assessment-procedures.csv --sha256 YOUR_VERIFIED_SHA256 --scope "Approved CUI enclave" --out objectives.json
mkdir -p evidence
./freecompliance cmmc objectives-assess --input objectives.json --evidence-dir evidence --json objective-report.json
```

A separate `cmmc score-template` command creates a draft 110-item worksheet. The `cmmc score-draft` calculator will **reject incomplete human determinations** and labels every result as unvalidated. See [source imports and draft scoring instructions](docs/CMMC_ASSESSMENT_NEXT.md). Neither function submits to SPRS.

## Get started in three steps

**Requirements:** Go 1.23+ to build from source. The core Go module has no third-party dependencies. Build the binary in an approved connected environment and transfer it under your organization's software-approval process if the target machine is air-gapped. Validate the software and transferred media before use.

```bash
git clone https://github.com/blocksifrdev/freecompliance.git
cd freecompliance
go build -o freecompliance ./cmd/freecompliance
```

**1. Create a CMMC evidence workspace.** Choose Level 1 (15 items) or Level 2 (110 requirement identifiers), and name the actual system being assessed.

```bash
./freecompliance cmmc init --level 2 --scope "CUI enclave A" --out workspace.json
```

**2. Keep evidence locally.** Create a folder and edit `workspace.json` to add reviewer assertions and references to approved evidence files. New controls start as `UNASSESSED`. FreeCompliance never automatically promotes file existence to control compliance.

```bash
mkdir -p evidence
```

**3. Check your evidence and generate an offline report.**

```bash
./freecompliance cmmc assess --input workspace.json --evidence-dir evidence --json readiness.json
```

The result shows which items are `DOCUMENTED`, `GAP`, `UNASSESSED`, `UNVERIFIABLE`, or `SCOPE_REVIEW_REQUIRED`. `DOCUMENTED` means a reviewer made a claim and the referenced files passed the available integrity and freshness checks. **It does not mean the CMMC requirement is satisfied.** Read [`docs/CMMC_AIRGAP.md`](docs/CMMC_AIRGAP.md) for how to record evidence safely and work in disconnected environments.

## What is included today?

| Capability | Current state | Network needed? |
| --- | --- | --- |
| CMMC Level 1 workspace | 15 safeguarding item identifiers and local evidence tracking | No |
| CMMC Level 2 workspace | 110 NIST SP 800-171 Rev. 2 requirement identifiers and local evidence tracking | No |
| Evidence integrity and freshness | SHA-256 comparison, age checks, bounded local file reading | No |
| Offline assessment report | JSON findings; no invented compliance score | No |
| Offline update authenticity check | Ed25519 signature check using an independently trusted public key | No |
| Repository security readiness scan | 14 local evidence assertions with HTML and JSON output | No |
| Bounded theorem evaluator | Experimental finite-world proof, counterexamples, and unknown outcomes | No |
| GitHub repository and source-bound rules assessment | Read-only GitHub API collection and narrow theorem checks | **Yes**, optional |
| Enterprise comparison benchmark | Reproducible benchmark calculator using independently adjudicated case data | No |

**What is not yet included:** Complete CMMC assessment objectives and procedures, authoritative control effectiveness testing across systems, validated SPRS scoring, official government submission, C3PAO accreditation, OSCAL-certified export, or a government endorsement. The local CMMC catalog is an **identifier index**, not a replacement for the authoritative requirement text or an approved assessment methodology.

## Objective-level CMMC review (experimental)

The offline objective workbench goes one level deeper than the requirement index. It records individual reviewer-assessed objectives, methods (examine/interview/test), source references, and evidence hashes without sending information outside the enclave.

```bash
./freecompliance cmmc objectives-init --level 2 --scope "CUI enclave A" --out objectives.json
mkdir -p evidence
./freecompliance cmmc objectives-assess --input objectives.json --evidence-dir evidence --json objective-report.json
```

**Important:** The generated 110-control workspace contains **empty objective lists**. Reviewers must populate those from the applicable, version-pinned authoritative NIST assessment procedures. The tool cannot yet validate that all official objectives have been included. Even an evidence-review-ready finding describes **listed objectives only**, not a passed CMMC requirement. No SPRS score, government submission, or CMMC certification is produced. See [the objective workbench guide](docs/CMMC_OBJECTIVE_WORKBENCH.md).

## Air-gapped deployment

FreeCompliance's `cmmc`, local `scan`, and local `prove` commands need **no network access, API keys, account, cloud tenant, telemetry, or LLM**. The optional `github` and `prove-github` commands **do** call GitHub and must not be used inside an air-gapped enclave.

Evidence stays on storage you control. No automatic export occurs. If you choose to move an evidence report outside a protected boundary, follow your organization's CUI handling, export review, and removable-media procedures. File paths, hashes, and metadata can themselves be sensitive.

To verify an offline package's detached Ed25519 signature with a pretrusted public key:

```bash
./freecompliance cmmc verify-bundle \
  --bundle approved-update.tar \
  --signature approved-update.sig \
  --public-key pretrusted-key.pub
```

Verification checks the package bytes against that key. **It does not establish that the key is trusted, that the contents are safe, or that the update should be installed.** See the [air-gap deployment guide](docs/CMMC_AIRGAP.md).

## What makes this different?

**Collecting evidence is not the same as proving a control.** FreeCompliance is building toward **Proof-Driven Compliance**: each technical claim has an explicit scope, source evidence, proof obligation, possible counterexamples, and a visible `UNKNOWN` outcome when the evidence is insufficient.

The theorem engine is experimental. It evaluates bounded models, **not every possible real-world state**. It does not yet automatically prove CMMC control effectiveness. To see a clearly labeled synthetic demonstration:

```bash
./freecompliance prove --input examples/theorem/branch-controls.json \
  --json theorem.json --html theorem.html
```

Open `theorem.html` locally. A model counterexample is a finding **within the model**, not proof of a real exploit or a failed government assessment. Read the [theorem documentation](docs/THEOREM_INTELLIGENCE.md).

## Other commands

Run a local repository scan (does not assess the whole organization):

```bash
./freecompliance scan --path . --json repository.json --html repository.html
```

Compare two saved scans:

```bash
./freecompliance compare --before first.json --after latest.json
```

Run a **connected, read-only** GitHub ruleset proof (not for an air gap):

```bash
./freecompliance prove-github --repo blocksifrdev/freecompliance \
  --json github-proof.json --html github-proof.html
```

For teams comparing the effort and reliability of assurance workflows, see the [30-day enterprise challenge](docs/ENTERPRISE_30_DAY_CHALLENGE.md) and its [measurement protocol](docs/MEASUREMENT_PROTOCOL.md). The included benchmark sample is synthetic and is **not** an independently measured comparison with Vanta, Symetri, or customers.

## CMMC requirements and important boundaries

- CMMC Level 1 references the 15 basic safeguarding requirements in FAR 52.204-21. Level 2 references the 110 security requirements in NIST SP 800-171 Revision 2, subject to applicable contracts and current government direction.
- The Department of War announced in July 2026 that CMMC Phase II rollout requirements were suspended while Phase I self-assessment requirements remained. **Check current official guidance and your contract** before making compliance or procurement decisions.
- The actual assessment scope, inherited controls, external service providers, independent assessor decisions, any permitted POA&M, and required SPRS submissions remain outside this tool's authority.
- A cryptographic file hash demonstrates an integrity comparison, not the truth of a document or successful control operation. A digital signature validates bytes against a key, not government approval.

Official references: [DoD CMMC program](https://business.defense.gov/cmmc/), [CMMC Phase II announcement](https://business.defense.gov/Engage/News/Newsfeed-Repository/Article/4542563/forging-the-arsenal-of-freedom-department-of-war-suspends-cmmc-phase-ii-require/), [FAR 52.204-21](https://www.acquisition.gov/far/52.204-21), [NIST SP 800-171](https://csrc.nist.gov/pubs/sp/800/171/r2/upd1/final), and [OSCAL](https://pages.nist.gov/OSCAL/).

**FreeCompliance is independent open-source software. It is not a DoD/Department of War product, an authorized certification body, or an official government compliance determination.**

## Free forever and open to contributors

The core assessment, offline evidence processing, theorem work, and export formats are Apache-2.0 open source. No BlockSiFr subscription is required. [BlockSiFr Control](https://blocksifr.com) is an optional separate commercial platform for consequence-aware runtime authority and execution control, not a dependency of FreeCompliance.

- [Contribute](CONTRIBUTING.md) | [Security policy](SECURITY.md) | [Threat model](docs/THREAT_MODEL.md) | [Release guidance](docs/RELEASE.md)
- Build and test: `go test ./...` and `go vet ./...`
- License: [Apache 2.0](LICENSE). Copyright 2026 BlockSiFr.

**Make compliance evidence accessible. Make uncertainty visible. Keep control of your data.**
