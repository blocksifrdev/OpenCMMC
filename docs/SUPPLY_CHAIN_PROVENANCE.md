# Signed release provenance and offline verification

OpenCMMC's distribution pipeline now requests keyless, signed Sigstore
attestations through GitHub Actions' short-lived OIDC identity.

For every candidate build the workflow:
- Creates Windows x64/ARM64 and Linux x64/ARM64 portable archives and a
  machine-readable SPDX SBOM.
- Writes a SHA-256 manifest and validates every archive and SBOM digest.
- Requests a SLSA build provenance attestation over those exact files and a
  separate SBOM attestation over the portable archives.
- Retains the signed Sigstore bundles as
  `opencmmc-provenance.sigstore.json` and
  `opencmmc-sbom.sigstore.json`, alongside the archives and checksums.
- Blocks publishing if any verification or signing action fails.

A **checksum alone does not establish publisher identity**. A cryptographically
verified provenance bundle does not establish absence of vulnerabilities,
assessor approval, CMMC compliance, government endorsement, or an approved
environment for CUI. Only trust bundles bound to the expected **public**
`blocksifrdev/OpenCMMC` repository and exact intended build ref, tag,
commit SHA, archive digest and signed builder identity.

## Normal verification before transfer to an offline computer

From an internet-connected, trusted verification workstation with GitHub CLI
installed, first verify the release's SHA256SUMS and the authenticated
attestation. For example:

```sh
sha256sum -c SHA256SUMS
sha256sum -c SHA256SUMS.sha256
gh attestation verify opencmmc-v1.0.0-rc.11-linux-amd64.tar.gz \
  --repo blocksifrdev/OpenCMMC
```

Substitute the actual release version and platform filename. Check that the
attested source commit and builder are the expected release production
workflow. Preserve that evidence with your signed software approval.

## Offline verification

Before air-gapping, transfer (using an approved media-introduction process):
the archive, SHA256SUMS, both signed Sigstore bundles and a separately trusted
Sigstore root for the relevant public/enterprise instance. Do **not** rely on
a trust root packaged inside the same archive being validated; provision
it out of band according to the organization's software assurance policy.

GitHub supports offline verification with a downloaded attestation bundle and
`gh attestation verify` using `--bundle` and
`--custom-trusted-root`. Verify the exact artifact digest and approved
repository, builder, source commit and certificate identity. Do not describe
a successful offline checksum alone as signature verification.

Official implementation and verification references:
- https://github.com/actions/attest
- https://docs.github.com/en/actions/how-tos/secure-your-work/use-artifact-attestations/verify-attestations-offline

## Remaining gates

External software assurance review, real target hardware and air-gapped
deployment acceptance, qualified CMMC assessment-methodology review and
customer-specific CUI authorization are tracked in
[the GA gate register](RELEASE_GATES_2026-11-11.md).

This signing workflow must itself pass CI. A code change that describes
attestations but has not generated and verified a real signature does not
establish provenance.
