# Security policy

**Scope:** OpenCMMC is an open-source, offline-first readiness workbench. It does **not** come with government authorization to operate, CUI accreditation, official CMMC certification, or a warranty that an organization's controls are effective.

## Report vulnerabilities privately

Do **not** submit proof-of-concept exploits, sensitive logs, passwords, CUI, customer artifacts, or private infrastructure details to public GitHub issues or PRs.

1. On the [OpenCMMC Security page](https://github.com/blocksifrdev/OpenCMMC/security), use **Report a vulnerability** if the repository has enabled private vulnerability reporting.
2. If that private channel is not available, request a secure contact route from repository maintainers **without disclosing exploit details publicly**. Do not assume an email address or bug-bounty program exists.
3. When a private channel is established, provide the affected version and OS, a minimal sanitized reproduction, the security impact, and whether there is evidence of exploitation. Never send CUI into GitHub.
4. Maintainers should acknowledge, triage, reproduce, develop a regression test, fix, and coordinate publication with the reporter. No specific response-time SLA is offered by this open-source preview.

## Supported release policy

Only the newest published [GitHub Release](https://github.com/blocksifrdev/OpenCMMC/releases) is recommended for evaluation. Older preview binaries are not security-maintained enterprise branches unless explicitly announced. Users must check for fixes manually; the offline app does not auto-update.

## Threat boundaries and mitigations

- The browser UI listens only on loopback 127.0.0.1 with private session token, Host/Origin and CSRF checks. It has **no managed multi-user identity, authorization, trusted device posture or transport encryption**.
- Evidence is confined to a user-controlled directory and hashed. Files are **not application-encrypted**, and file integrity is not evidence that technical controls operate effectively.
- The user's operating-system permissions, device encryption, antivirus, approved software supply chain, backup policies and enclave boundary must protect the evidence.
- The on-device preview is not approved for storing or processing CUI by virtue of being offline. Each organization must independently approve its deployment boundary and handling controls.
- `github` and `prove-github` are explicitly network-dependent optional CLI operations; they must not be used inside a real air gap.
- SHA256SUMS alone cannot establish publisher identity. Only builds with *verified* release provenance attestations should be described as having that capability. A Sigstore/GitHub attestation is not Windows Authenticode signing, and neither is government authorization.

## Secure development and releases

Changes require regression tests and passing GitHub CI before merge. Release packages include a manifest and checksum; Preview 4 includes GitHub/Sigstore build-provenance attestations and an SPDX Go-module inventory, both requiring independent verification against a customer's trust policy. See [enterprise evaluation gates](docs/ENTERPRISE_EVALUATION.md) and [release integrity](docs/RELEASE_INTEGRITY.md).

This project follows selected secure-software-development practices informed by NIST SP 800-218, **not** a NIST/DoD-certified SSDF conformance program.
