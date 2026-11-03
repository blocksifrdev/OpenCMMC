# OpenCMMC is available now

**Published: November 2, 2026**  
**Official launch: Veterans Day, November 11, 2026**

## Stop guessing about CMMC. Start proving your controls.

**Know what's required. See what's documented. Find what's missing.**

OpenCMMC by BlockSiFr is publicly available today, ahead of our formal Veterans Day launch.

We are making a permanent, free and open-source CMMC readiness foundation available to the Defense Industrial Base: small defense businesses, prime contractors, subcontractors, security and compliance teams, assessors, and government technical reviewers.

The project is dedicated in gratitude for military service and inspired by the founder's father, a lieutenant colonel in the U.S. Army.

**Why it matters:** Essential cybersecurity readiness should not require a paid subscription or uploading sensitive evidence to someone else's cloud.

### Available to use today

- An offline CMMC Level 1 evidence workspace with 15 safeguarding item identifiers.
- An offline CMMC Level 2 evidence workspace with 110 NIST SP 800-171 Revision 2 requirement identifiers.
- Local evidence references, SHA-256 integrity checks, configurable freshness checks, and reviewer-entered findings.
- Offline JSON readiness reports that distinguish documentation, gaps, and unresolved items.
- Detached Ed25519 bundle signature verification with independently trusted public keys.
- A separate, experimental bounded Theorem Intelligence evaluator and a source-bound GitHub proof option; these are not a complete CMMC assessment.

**No account. No subscription. No required cloud service. Open source under Apache 2.0.**

### Get started now

Source: https://github.com/blocksifrdev/OpenCMMC

```bash
git clone https://github.com/blocksifrdev/OpenCMMC.git
cd freecompliance
go test ./...
go build -o freecompliance ./cmd/freecompliance
./freecompliance cmmc init --level 2 --scope "Demo enclave" --out workspace.json
mkdir -p evidence
./freecompliance cmmc assess --input workspace.json --evidence-dir evidence --json readiness.json
```

The source checkout requires network access, but the built core CLI runs offline. In air-gapped environments, approve and transfer software through your organization's authorized procedures. Keep CUI and security-sensitive evidence in your approved boundary.

### What comes next

On **November 11**, we formally launch the Veterans Day service initiative, demonstrate a disconnected CMMC readiness workflow, and invite broader independent evaluation and open-source contributions. Access does **not** wait until November 11.

Objective-by-objective assessment under applicable official procedures, effectiveness testing across real systems, validated draft SPRS scoring, and schema-tested OSCAL export remain development milestones. Government submission authority, C3PAO accreditation, and government endorsement are not capabilities that this software can grant.

**OpenCMMC does not itself certify CMMC compliance or replace an authorized assessment. A matching file hash does not prove a control works.** There is no claim of adoption, sponsorship, or endorsement by the Department of Defense / Department of War.

### Help improve it

Try the offline quick start with **non-sensitive test data**, report issues or confusing requirements, and suggest evidence adapters that would help small defense suppliers.

**The commitment: Compliance should be proven, not presumed. And the tools to build that proof should be free.**

[README and quick start](../README.md) | [Offline deployment guide](CMMC_AIRGAP.md) | [Veterans Day commitment](VETERANS_DAY_2026.md) | [CMMC development roadmap](CMMC_PROOF_ROADMAP.md)
