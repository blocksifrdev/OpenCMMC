# Verify OpenCMMC downloads

Use the official [OpenCMMC Releases page](https://github.com/blocksifrdev/OpenCMMC/releases) and your organization's approved acquisition process.

## Check file integrity

1. Download the archive for your device and the matching `SHA256SUMS` file **from the same release**.
2. On Linux, run `sha256sum ARCHIVE`. On Windows PowerShell, run `Get-FileHash ARCHIVE -Algorithm SHA256`.
3. Compare the hash to the matching archive line in the released `SHA256SUMS` file. Reject mismatches.
4. Review the released source commit, GitHub Actions build results, and included `SBOM.spdx.json` before use.

**Limits:** SHA-256 detects changes in bytes but is not a publisher signature or proof that software is safe. GitHub build provenance should be claimed only for specific artifacts when independently verified. The current portable Windows archives are not asserted to carry Windows Authenticode signatures or government authorizations.

For an air gap, verify the release on an approved staging machine, transfer through authorized media, then recheck the archive hash after transfer. Never upload customer CUI, secrets, or access credentials for troubleshooting.

[Installation](INSTALL_V1.md) · [Security policy](../SECURITY.md) · [Air-gapped operation](CMMC_AIRGAP.md)
