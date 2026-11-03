# Release provenance, software composition, and offline transfer

## Current status

The published [Preview 4](https://github.com/blocksifrdev/OpenCMMC/releases/tag/v0.5.0-preview.4) contains four platform archives, SHA-256 integrity manifests, an SPDX 2.3 Go module inventory, and GitHub/Sigstore **build-provenance attestations**. The [release CI run #41532097431](https://github.com/blocksifrdev/OpenCMMC/actions/runs/41532097431) completed successfully, including on-runner verification of release attestation subjects before publication. Independent customers must repeat verification against their own trust policy.

Preview 4 **does not** have Windows Authenticode signing, an independently provisioned offline root-of-trust, a government ATO, or third-party CMMC accreditation. The Go module inventory includes the project module, listed Go modules, and toolchain metadata. It **does not** inventory operating systems, firmware, loaded external applications, or identify all vulnerabilities. Older [Preview 3](https://github.com/blocksifrdev/OpenCMMC/releases/tag/v0.5.0-preview.3) did not carry these signed provenance attestations.

## Before importing software into a restricted environment

1. In an approved connected staging area, open the GitHub Release page and verify the release version, source commit, and successful release workflow.
2. Download **the specific archive for your approved OS/architecture**, `SHA256SUMS`, and `SBOM.spdx.json` where available.
3. Run `sha256sum -c SHA256SUMS` on Linux or compare `Get-FileHash` on Windows. Checksums detect changes **relative to the selected manifest**, not a trusted publisher.
4. For a release marked as having provenance attestations, run `gh attestation verify ARCHIVE -R blocksifrdev/freecompliance` in staging. This must succeed; otherwise the release cannot be treated as provenance-verified. Inspect repository identity, commit and workflow. See [GitHub's official verification documentation](https://docs.github.com/en/actions/how-tos/secure-your-work/use-artifact-attestations/use-artifact-attestations).
5. For offline verification, obtain the artifact attestation bundle with `gh attestation download ARCHIVE -R blocksifrdev/freecompliance` and approved Sigstore trust roots using `gh attestation trusted-root` on staging. Transport both with the approved installer and `gh` CLI into the enclave, keeping the trust root separately authenticated through an organization-approved path.
6. Verify offline via `gh attestation verify ARCHIVE -R blocksifrdev/freecompliance --bundle ATTESTATION.jsonl --custom-trusted-root trusted_root.jsonl`. Independently verify the repository ID and expected source commit; verify the CLI and trust roots themselves under your organization's approved chain of custody. See [official offline verification guide](https://docs.github.com/en/actions/how-tos/secure-your-work/use-artifact-attestations/verify-attestations-offline).
7. Apply your own malware analysis, code signing requirements, removable-media handling, authorization to operate, endpoint security and CUI procedures **before** execution. The open-source release does not bypass these gates.

**Important:** GitHub/Sigstore build provenance is a signed claim about the build workflow and source. It does not prove code is safe, establish your organization's authorization, provide Windows Authenticode publisher trust, or grant any government accreditation. An offline trust root must be obtained and authenticated independently. Verify the live Release's attestation before using these instructions.

## Release process

A release is created from a new `release/vX` branch only after PR checks pass; release branch creation alone does not publish. A special maintainer-reviewed release marker triggers packaging. A guarded job performs Go tests and race detection, `go vet`, package checksums, offline smoke tests, complete human-adjudication fail-closed scoring tests, SPDX module inventory generation, and release provenance attestation. A publisher job uploads packages only if the package job succeeds. A failed attestation or verification must block the release.

## Non-guarantees

No self-attested score, documented evidence, PDF identifier count, SPDX inventory, or passing CI establishes CMMC certification or a valid SPRS submission. Customer CUI should not be placed in GitHub or public support issues.
