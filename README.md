# FreeCompliance

**Free compliance readiness. Evidence you can inspect. No false green.**

FreeCompliance is an Apache-2.0 open-source utility from BlockSiFr. It inspects a local software repository for independently reproducible security and governance signals, records SHA-256 evidence digests, and creates a structured JSON assessment plus a polished self-contained HTML evidence report. It does not ask for credentials, upload source code, or charge for its core functionality.

> Compliance requirements shouldn't be hidden behind a subscription. The work of validating that controls actually function still matters. We make the first part free.

## 60-second quick start

Requires Go 1.23+ for building from this source tree. No third-party Go modules.

Unpack the included source archive, then run:

```bash
cd freecompliance
go test ./...
go run ./cmd/freecompliance scan --path . --json readiness.json --html readiness.html
```

Once the public standalone repository is published, it can be cloned from `https://github.com/BlockSiFr/freecompliance`. This is a planned location, not a claim that publication has happened. Open `readiness.html` in any browser.

Assess another local repository:

```bash
go build -o freecompliance ./cmd/freecompliance
./freecompliance scan --path ../my-app --json ../my-app-readiness.json --html ../my-app-readiness.html
```

Check drift from a saved baseline:

```bash
./freecompliance compare --before first.json --after latest.json
```

Gate a pipeline only on *observed* gaps:

```bash
./freecompliance scan --path . --fail-on-gap
```

`--fail-on-gap` will not fail on unknown, unobserved, or merely undocumented controls. Choose stricter gates through your organization's own policy, not by interpreting unknown as passing.

## Inspect actual GitHub branch metadata (v0.2.0)

The optional GitHub assessor performs **GET-only** calls against the GitHub REST API. The local scan never makes network requests. The GitHub assessor retrieves repository visibility, lifecycle metadata, and the default branch's `protected` flag. It does **not** conclude that review approvals or required checks are enforced.

```bash
# Public repository (up to GitHub's unauthenticated rate limit):
go run ./cmd/freecompliance github --repo BlockSiFr/freecompliance --json github.json --html github.html

# For private repositories, use a read-only token with Contents: read and Metadata: read:
export GITHUB_TOKEN="<token supplied securely outside this repository>"
go run ./cmd/freecompliance github --repo YOUR-ORG/YOUR-REPO --fail-on-gap
unset GITHUB_TOKEN
```

Never put tokens in a CLI argument, issue, screenshot, GitHub Actions log, report, or committed file. The JSON and HTML reports include response **hashes** and endpoint paths but never the credential. For API failures, the command produces an error rather than a false green. If a branch query is unavailable after successfully retrieving repository metadata, branch protection is marked `ATTESTATION_NEEDED`.

`VERIFIED` for `GITHUB_BRANCH_PROTECTED` means only that the API reported the flag as true. It is not evidence that *specific* required-review or CI rules work. This is intentionally narrower than a full compliance claim.

## Drop-in GitHub Action (after public release)

This repository includes a composite GitHub Action. Once a public release is tagged, any GitHub repository can use it without credentials or a subscription. First check out the target repo and install Go, then run:

```yaml
- name: Free repository evidence scan
  uses: BlockSiFr/freecompliance@v0.2.0
  with:
    fail-on-gap: 'true'
```

The action writes a summary into the GitHub Actions job UI and exposes paths to local JSON and HTML reports as outputs. It does not auto-publish evidence or leave comments. For supply-chain hardening, replace the tag with the reviewed immutable action commit SHA when pinning it in production.

This `uses:` reference will work **only after the repository and release exist**.

## What it does today

- Finds a declared vulnerability reporting channel (`SECURITY.md`) and code ownership (`CODEOWNERS`).
- Finds Dependabot configuration, workflow declarations, automated test sources, lockfiles, and incident response documents.
- Identifies **external GitHub Actions references that aren't pinned to full immutable commit SHAs**.
- Flags **explicit `permissions: write-all`** in workflow YAML.
- Identifies declared code scanning and SBOM files.
- Treats branch protection, production IAM and control effectiveness as **ATTESTATION_NEEDED** until a suitable evidence source is connected.
- Produces 14 local evidence claims and 5 optional GitHub metadata assertions with a status, observation, recommended next step, and NIST CSF 2.0 *informative* alignment annotation.
- Generates offline HTML and machine-readable JSON with content SHA-256 digests. `compare` shows changed finding statuses.

This is a **repository-scope** assessment, not an environment or organization-wide scan. Artifact presence is evidence of existence, not proof of implementation or effectiveness.

## Status semantics

| Status | Meaning |
|---|---|
| `VERIFIED` | The specific stated scanner assertion has direct observed local evidence |
| `PARTIAL` | Some evidence exists, but the broader claim still needs verification |
| `ATTESTATION_NEEDED` | The offline scan cannot determine the claim |
| `MISSING` | An expected local artifact wasn't found |
| `GAP` | A specific undesirable condition was observed |
| `INHERITED` | Reserved for future externally validated inherited-control evidence; not currently emitted |

No “compliance percentage” appears, because a repository-only count would falsely imply an auditable whole-organization score. The counts are **scanner assertions**, not regulatory coverage.

## Trust and privacy model

- Local, read-only scanning by default. No cloud service, model API, analytics, telemetry, or credential collection.
- Report output contains **paths, sizes and file digests**, not raw file contents.
- Hidden `.git`, dependency caches, vendored directories and common build outputs are skipped.
- Evidence only accepts regular files up to 1 MiB; symlinks are not followed.
- Reports created by the CLI use mode `0600` on Unix-like filesystems. If you publish a report, you choose to disclose those paths and hashes.
- The report records SHA-256 digests for integrity comparison, **not a digital signature, trusted timestamp, or proof that any control executed**.
- The source is inspectable and portable. There is no SaaS lock-in.

## Standards and legal boundaries

The initial framework tags point to relevant **NIST Cybersecurity Framework 2.0** subcategories as informative references. They do not reproduce third-party protected control texts. See [NIST CSF 2.0](https://www.nist.gov/cyberframework) and [OSCAL](https://pages.nist.gov/OSCAL/).

**This utility does not certify or establish compliance with SOC 2, CMMC, ISO 27001, HIPAA, FedRAMP, PCI DSS or any other regime.** An external audit, assessor, legal interpretation, contract requirement or government submission may still be necessary. “VERIFIED” is **never** equivalent to passing an entire requirement.

## Community roadmap

The open-source roadmap, with no planned paywall on core assessment or export:

1. Extend the read-only GitHub integration to independently verify required status checks, review rules, bypass rights and code scanning results, using narrowly scoped permissions.
2. User-defined control policy and truthful `NOT_APPLICABLE`/scope exclusions with reviewer identity.
3. OSCAL profile imports, assessment results, POA&M export and community-maintained versioned mappings.
4. Reproducible control tests (including negative-path verification) with independently signed evidence receipts.
5. AWS, Azure, GCP, Kubernetes and identity-provider evidence adapters with explicit permission boundaries.
6. Public auditor and assessor handoff packages with redaction controls and continuous drift review.

BlockSiFr Control remains a **separate optional enterprise product** for runtime authority, consequential-action enforcement and signed execution receipts. FreeCompliance will never require purchasing it.

## Release and project governance

See [docs/RELEASE.md](docs/RELEASE.md) for phone-first publication, immutable release tags, checksum generation, and operational steps. Read the [threat model](docs/THREAT_MODEL.md) before deploying into sensitive repositories.

## Contributing

Run `go test ./...` and `go vet ./...`. See [CONTRIBUTING.md](CONTRIBUTING.md). Community mappings must include source version, rationale, scope, license check and tests. Do not add credentials, secrets or proprietary framework text to the repository.

## License

[Apache License 2.0](LICENSE). Copyright 2026 BlockSiFr.
