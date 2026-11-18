# OpenCMMC @OPENCMMC_VERSION@

An offline-first CMMC readiness and evidence workbench, created by BlockSiFr.
This is the **exact packaged build version**; check `opencmmc version` after
extracting the archive. Free and open source under the bundled LICENSE.

OpenCMMC organizes 15 Level 1 safeguarding requirements and 110 Level 2
NIST SP 800-171 Rev 2 requirement identifiers. Level 2 assessment
objectives require separate authoritative import and human review; the
initial set of identifiers is not a complete verified objective assessment.
Evidence paths and matching hashes do not establish effective controls.

The local workbench runs only on this computer, without cloud sign-in,
mandatory network access, telemetry, third-party JavaScript or account fees.
There is no managed multiuser authentication or app-level encryption.

**Not CMMC-certified, not an accredited assessment, not official SPRS
submission, not government-endorsed, and not an authorization to handle
CUI.** Your organization must approve the host, software supply chain,
encryption, evidence handling, retention and security boundary.

## Start

See `INSTALL-OFFLINE.txt` in this archive for both Windows and Linux
installation and privacy-safe local browser guidance. Keep the assessment
directory on approved encrypted storage; never delete it during an upgrade.

The portable binary can create and assess offline workspaces, import
independently verified NIST procedures, export local readiness reports, and
perform local evidence backup/verification/restoration. Backup ZIP archives
are **not encrypted**; store them only on approved encrypted media.

## Release authenticity

A published SHA256SUMS file detects changed bytes but cannot authenticate
the publisher. From the separately trusted published release, verify both
the exact archive digest and the signed GitHub/Sigstore provenance and SBOM
bundle against the canonical `blocksifrdev/OpenCMMC` repository and intended
release tag, source commit and workflow identity. Do not trust a public key
obtained only from an unverified ZIP archive.

Source, current security policy and supported download details:
https://github.com/blocksifrdev/OpenCMMC

For confidential vulnerability reports, follow that repository's
`SECURITY.md`. Never post credentials, actual CUI, private workbench URLs or
customer evidence in GitHub issues.
