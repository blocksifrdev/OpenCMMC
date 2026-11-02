# FreeCompliance portable downloads

**Public preview v0.5.0-preview.2:** [GitHub Releases](https://github.com/blocksifrdev/freecompliance/releases).

This download is a **portable command-line application**, not a graphical installer. Preview 2 includes the new offline NIST SP 800-171A objective CSV importer and the **draft-only** Level 2 score worksheet. See [CMMC assessment and scoring instructions](CMMC_ASSESSMENT_NEXT.md). Windows users can double-click the included demo launcher. No Go compiler, subscription, login, or internet is required to run the packaged core tools. A local visual interface remains a separate product milestone.

## Choose your computer

| Computer | File name | Offline demo |
| --- | --- | --- |
| Windows x64 (Intel/AMD) | `freecompliance-v0.5.0-preview.2-windows-amd64.zip` | Double-click `START-CMMC-DEMO.cmd` |
| Windows ARM64 | `freecompliance-v0.5.0-preview.2-windows-arm64.zip` | Double-click `START-CMMC-DEMO.cmd` |
| Linux x64 | `freecompliance-v0.5.0-preview.2-linux-amd64.tar.gz` | Run `./start-cmmc-demo.sh` |
| Linux ARM64 | `freecompliance-v0.5.0-preview.2-linux-arm64.tar.gz` | Run `./start-cmmc-demo.sh` |

**Windows:** Download the ZIP and SHA256SUMS, verify the ZIP using PowerShell `Get-FileHash .\freecompliance-v0.5.0-preview.2-windows-amd64.zip -Algorithm SHA256` and compare the output with the matching row in SHA256SUMS. Extract to an approved local folder, then double-click `START-CMMC-DEMO.cmd`. To use the CLI directly, run `.\freecompliance.exe help` from PowerShell in the extracted folder.

**Linux:** Download the TAR.GZ and SHA256SUMS; run `sha256sum` on the archive and compare it to the matching row in SHA256SUMS. Extract using `tar -xzf freecompliance-v0.5.0-preview.2-linux-amd64.tar.gz`, then run `./start-cmmc-demo.sh` or `./freecompliance help`. Review the launch script before running it.

The launcher creates a synthetic CMMC Level 2 workspace in the extracted directory and a local readiness report. All 110 requirements remain unassessed initially. The launchers require no credentials or network.

## Air-gapped / CUI environments

1. Obtain, inspect and approve the package in a connected staging environment operated under your organization's policies.
2. Check the SHA-256 against the published manifest, inspect source commit and CI build provenance, and use your organization's **trusted software publisher verification process**.
3. **SHA256SUMS is an integrity manifest, not a publisher signature. This preview has no independently trusted offline signing key.** Do not bypass government or organization software restrictions to install it.
4. Transfer the approved archive and verified metadata using approved removable-media procedures.
5. Recheck the digest after transfer; extract and run core commands locally. The `github` and `prove-github` subcommands are online and must not be used in a true air gap.
6. Keep CUI, filesystem paths, hashes, reports, reviewer names and technical metadata within the authorized enclave.

**Scope limits:** FreeCompliance provides offline readiness and evidence-integrity workspaces, a separately sourced NIST CSV objective importer, draft-only user-entered scoring, and experimental bounded theorem evaluation. It does not supply a **PDF-reconciled and independently reviewed** CMMC objective methodology, independently validated SPRS scoring, official submission, accredited CMMC assessments, or government endorsement.

Windows SmartScreen or Defender may flag an unsigned preview executable. Do **not** bypass security controls. Obtain approval from your IT or security team. Report issues with sanitized, non-sensitive samples only.

For details see [CMMC air-gapped operation](CMMC_AIRGAP.md) and [objective workbench](CMMC_OBJECTIVE_WORKBENCH.md).