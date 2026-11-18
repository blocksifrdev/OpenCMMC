# Operating the protected OpenCMMC v1.0.0 GA release

**Current decision: NO-GO.** Configuring these protections does not count
as obtaining independent security, CMMC, hardware, air-gap or accessibility
acceptance. The actual reviewers must issue and sign truthful assessments.

The only automated production publication workflow is:
`.github/workflows/opencmmc-ga-publish.yml`. It runs by **manual dispatch
from main**, never by a PR, scheduled task or a branch push.

## Before the first dispatch

1. Protect `main` with required reviews and successful engineering CI
   checks. An owner must verify repository rules directly.
2. Create the `ga-production` GitHub environment with **at least two distinct
   named human required reviewers**; the workflow verifies the configured
   approval rule through GitHub's API and blocks an unprotected or unreadable
   environment. Do not approve a release for your own unreviewed commit.
3. Obtain independent external security testing and remediation/retest,
   qualified CMMC objective/methodology review, the actual Chromebook test,
   an air-gapped/recovery pilot, an accessibility audit, and owner authority.
   See [GA acceptance protocol](GA_ACCEPTANCE_PROTOCOL.md).
4. Capture the exact, frozen reviewed `main` commit SHA and the exact
   `v1.0.0` four-archive digest fingerprint printed by
   `scripts/verify_ga_signoffs.py --print-archive-fingerprint`.
5. Have every actual reviewer sign their exact JSON statement with their
   separately controlled Ed25519 key; see
   [independent signed approval protocol](GA_SIGNED_APPROVALS.md).
6. Establish **six trusted reviewer public keys out of band**, with human
   verification of each identity and role. Neither source checkout nor a
   reviewer-supplied bundle is an acceptable trust source by itself.
7. Store private, redacted external evidence in an approved controlled
   vault. The release script only receives statements, detached signatures
   and public keys. Do not put CUI, security findings, private signing keys,
   access URLs, or customer records into GitHub.

## Protected-environment inputs

The release job requires two *separate* GitHub environment secrets, accessible
only after environment review:

- `GA_SIGNOFF_RECORDS_B64`: base64 ZIP with exactly six
  `ROLE.json` files and six corresponding `ROLE.sig` detached Ed25519
  signatures, no nested directories.
- `GA_TRUSTED_PUBLIC_KEYS_B64`: separately provisioned base64 ZIP
  containing exactly six `ROLE.pub.pem` public keys; these keys must have
  been independently authenticated. Do not put private keys in the ZIP.

The role identifiers are `independent-security`, `cmmc-methodology`,
`chromebook-acceptance`, `airgap-acceptance`,
`accessibility-review`, and `release-owner`.

The archive unpacker refuses missing/extra/duplicate/traversal members,
symlinks and large files. The signature verifier refuses absent/expired
review, DENY, incorrect source or archive SHA, wrong trust keys and altered
signatures. These checks operate in a temporary private directory on the
GitHub-hosted runner.

## Dispatch procedure

Only an authorized operator should select:
**Actions → OpenCMMC v1.0.0 GA | protected and independently approved → Run
workflow**, using the `main` ref and exact `approved_sha`.

The workflow then:
- Refuses any repository or branch other than canonical public `main` and
  any SHA not exactly equal to the reviewed source.
- Rebuilds signed `v1.0.0` archives using the same reusable release pipeline
  as the proven rehearsal, including Go 1.27.4, reachable-advisory scanning,
  CodeQL tracking, native backup/restore smoke and real Chromium E2E.
- Waits for the protected environment's reviewer approval.
- Confirms at least two distinct human reviewers are configured, and fails
  if the GitHub configuration/API is inaccessible.
- Blocks any unresolved high or critical GitHub code scanning alert
  (or failure to retrieve alerts).
- Validates SHA-256 and signed Sigstore build and SBOM attestations, then
  verifies six independently trusted Ed25519 approvals tied to the exact
  source and four archive byte hashes.
- Creates **only a draft GitHub Release**, downloads every asset from
  GitHub, revalidates checksum/manifest, independent approvals and signing
  identity, and **only then** changes the draft to public `v1.0.0`.

A failed gate must leave GA unpublished. If a draft release is created but
a later check fails, the draft remains for authorized manual review; the
workflow refuses to overwrite an existing tag or release on retry.

## Independent verification after GA

Check the canonical published release, exact source commit, four archive
hashes, SHA256SUMS, SPDX SBOM, both Sigstore signature bundles, and
`gh attestation verify` with the approved repository and predicate types:

```sh
gh attestation verify opencmmc-v1.0.0-linux-amd64.tar.gz \
  --repo blocksifrdev/OpenCMMC \
  --bundle opencmmc-provenance.sigstore.json \
  --predicate-type https://slsa.dev/provenance/v1

gh attestation verify opencmmc-v1.0.0-linux-amd64.tar.gz \
  --repo blocksifrdev/OpenCMMC \
  --bundle opencmmc-sbom.sigstore.json \
  --predicate-type https://spdx.dev/Document/v2.3
```

A product release never constitutes CMMC certification, formal SPRS
submission, C3PAO accreditation or an organization's CUI system ATO.
