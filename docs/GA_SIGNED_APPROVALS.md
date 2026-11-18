# OpenCMMC v1.0: independent cryptographic approval gate

**GA is still NO-GO.** This is a release-verification mechanism, not an
authorization that has already been issued. Real security/CMMC specialists,
test operators and the accountable release owner must conduct their work.

## Required individually signed decisions

Six separately identified, dated reviewers/operators must supply
`APPROVE` decisions under these exact roles:

1. `independent-security` - external adversarial testing, high/critical remediation
2. `cmmc-methodology` - qualified Rev 2 objective/methodology/scoring reconciliation
3. `chromebook-acceptance` - real Chromebook browser workflow acceptance
4. `airgap-acceptance` - actual isolated device installation, recovery and rollback
5. `accessibility-review` - independently documented keyboard/assistive-technology audit
6. `release-owner` - accountable organizational publishing authorization

The role names are machine-verifiable; **identity, qualifications, independence
and report accuracy require human verification**. A signature from a tester
whose report is not valid does not establish production readiness.

## Exact source and artifact identity

Build the **exact proposed** `v1.0.0` commit on a trusted runner with the
versioned workflow. Do not add commits after reviewer approval without
issuing new statements.

Run:

```bash
python3 scripts/verify_ga_signoffs.py \
  --version v1.0.0 --source-sha EXACT40CHARLOWERCASECOMMITSHA \
  --dist dist --print-archive-fingerprint
```

This prints one stable SHA-256 fingerprint derived from four independently
validated portable archive hashes. Unlike the full `SHA256SUMS` file, this
fingerprint is unchanged if the signed SBOM/provenance bundles differ between
otherwise byte-identical builds. The verifier does not assume two builds are
reproducible; it compares the final four archive hashes with the signatures.

Store each **private review report** in separately authorized storage, not
GitHub, and compute the report's SHA-256. Do not transfer CUI, security
reports, private workbench URLs or credentials into the repository.

## Signed statement format

For each role, save a UTF-8 JSON file as `ROLE.json`. It must have exactly
these keys, with no extra keys or duplicate names:

```json
{
  "schema": "opencmmc.ga.signoff.v1",
  "role": "independent-security",
  "release_version": "v1.0.0",
  "source_sha": "EXACT40CHARLOWERCASECOMMITSHA",
  "archives_sha256": "EXACT64CHARARCHIVEFINGERPRINT",
  "report_sha256": "EXACT64CHARPRIVATEREPORTDIGEST",
  "reviewer": "Independent reviewer's verified identity",
  "reviewed_at": "2026-11-18T20:00:00Z",
  "decision": "APPROVE"
}
```

The above is a **format example, not an approval**. Do not reuse or invent
these values. Record real source/sha/reviewer/time/report values.

Each reviewer retains their own Ed25519 private key and signs the **exact
unchanged bytes** of their JSON file:

```bash
openssl pkeyutl -sign -inkey /trusted/private/reviewer-ed25519.pem \
  -rawin -in ROLE.json -out ROLE.sig
```

The security/release owner provisions the reviewer public key
`ROLE.pub.pem` in an independent **trusted** directory. Never accept a
reviewer's claimed key merely because it accompanied a signed record.
Verifying six signatures with six self-supplied or unverified keys proves
nothing about reviewer independence.

## Fail-closed prepublication verification

```bash
python3 scripts/verify_ga_signoffs.py \
  --version v1.0.0 \
  --source-sha EXACT40CHARLOWERCASECOMMITSHA \
  --dist dist \
  --records-dir /approved/review-statements \
  --trust-dir /independently/provisioned/trusted-public-keys
```

The command fails unless **all six** signatures validate under the trusted
role-specific Ed25519 keys and match the exact GA source and archive hashes.
It rejects absent approvals, DENY, expired (>90 days), future timestamps,
bad signatures, duplicate JSON fields and manipulated binaries. It is
entirely local/offline and logs no report content.

The verifier does **not** issue an independent report, verify its technical
quality, confirm reviewer qualification, or authorize CUI. It is one
cryptographic condition of publication, not the entire GA decision.
