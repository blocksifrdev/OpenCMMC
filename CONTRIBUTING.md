# Contributing to OpenCMMC

OpenCMMC should become useful to anyone with a repository, not just customers of BlockSiFr.

## Principles

1. **No false green.** A positive result requires independently observable evidence supporting exactly the claim stated in the rule. A repository artifact cannot prove production IAM, runtime safety or certification.
2. **Open standards and portable evidence.** Prefer NIST OSCAL and published framework references. Do not copy controlled or licensed standards without the appropriate redistribution permission.
3. **Offline by default.** No hidden network calls, no telemetry, no credential harvesting. Remote adapters must be explicit, read-only where possible and opt-in.
4. **Proof and provenance.** Every automated rule needs positive, negative, missing-evidence and malicious-input tests. Include the evidence scope and any limitations.
5. **Fail on observed unsafe state, not silence.** Unknown does not mean success or failure; it means investigation.
6. **Community-first.** No subscription, account registration or proprietary service dependency for core checks or evidence export.

## Development

```bash
go test ./...
go vet ./...
go run ./cmd/opencmmc scan --path . --json /tmp/opencmmc.json --html /tmp/opencmmc.html
```

## Submitting a rule

Document the assertion, source of evidence, false-positive/false-negative limits, test cases, license of referenced control text, exact NIST CSF 2.0 code if applicable, and remediation guidance. Explicitly distinguish a declaration from effective enforcement.

## Reporting security concerns

Report privately to the maintainers via GitHub's private vulnerability reporting feature when configured. Do not include secrets or real sensitive infrastructure configuration in issues.
