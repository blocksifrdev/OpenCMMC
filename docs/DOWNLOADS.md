# OpenCMMC downloads

## Current public release candidate: v1.0.0-rc.1

[Official Veterans Day release page](https://github.com/blocksifrdev/OpenCMMC/releases/tag/v1.0.0-rc.1) | [Offline install guide](INSTALL_V1.md) | [Release notes](RELEASE_NOTES_v1.0.0-rc.1.md)

Download the `opencmmc-v1.0.0-rc.1` ZIP/TAR.GZ for your platform and the [SHA256SUMS](https://github.com/blocksifrdev/OpenCMMC/releases/download/v1.0.0-rc.1/SHA256SUMS) file. Public artifacts were built and smoke-tested on native Windows ARM64, Windows x64, Linux ARM64 and Linux x64 GitHub runners before release. The candidate is **not** official CMMC certification, complete enterprise assessment functionality or v1.0.0 general availability.

## Historical Preview 4 downloads

**Public preview v0.5.0-preview.4:** [GitHub Releases](https://github.com/blocksifrdev/OpenCMMC/releases).

**Compatibility note:** The published v0.5.0-preview.4 downloads and commands still bear the legacy `freecompliance` name. New v1 candidate packages use `opencmmc` and include a backward-compatible executable alias. The candidate is not a publicly verified general-availability release.

This download is a **portable Go executable with a self-contained local browser interface**, not a native MSI or EXE installer. Use START-CMMC-UI.cmd on Windows or start-cmmc-ui.sh on Linux to open the graphical workflow without editing JSON. The CLI is also included. Preview 4 includes the offline NIST SP 800-171A supplemental CSV objective importer and the **draft-only** Level 2 score worksheet. See [CMMC assessment and scoring instructions](CMMC_ASSESSMENT_NEXT.md). Windows users can double-click the browser UI launcher; the synthetic demo launcher is also included. No Go compiler, subscription, login, or internet is needed for the packaged offline tools. The browser UI is available now; a native signed installer remains a separate milestone.

**Provenance:** Preview 4 includes a released SPDX module inventory and a GitHub/Sigstore provenance attestation for each package. Follow [release-integrity instructions](RELEASE_INTEGRITY.md) and verify the specific artifact from a trusted staging machine. An attestation is not an Authenticode certificate or government authorization.

## Choose your computer

| Computer | File name | Offline demo |
| --- | --- | --- |
| Windows x64 (Intel/AMD) | `freecompliance-v0.5.0-preview.4-windows-amd64.zip` | Double-click `START-CMMC-UI.cmd` |
| Windows ARM64 | `freecompliance-v0.5.0-preview.4-windows-arm64.zip` | Double-click `START-CMMC-UI.cmd` |
| Linux x64 | `freecompliance-v0.5.0-preview.4-linux-amd64.tar.gz` | Run `./start-cmmc-ui.sh` |
| Linux ARM64 | `freecompliance-v0.5.0-preview.4-linux-arm64.tar.gz` | Run `./start-cmmc-ui.sh` |

**Windows:** Download the ZIP and SHA256SUMS, verify the ZIP using PowerShell `Get-FileHash .\freecompliance-v0.5.0-preview.4-windows-amd64.zip -Algorithm SHA256` and compare the output with the matching row in SHA256SUMS. Extract to an approved local folder, then double-click `START-CMMC-UI.cmd`. The launcher attempts to open the private local browser URL automatically. If your browser does not open, copy the URL shown in the command window. To use the CLI directly, run `.\freecompliance.exe help` from PowerShell in the extracted folder.

**Linux:** Download the TAR.GZ and SHA256SUMS; run `sha256sum` on the archive and compare it to the matching row in SHA256SUMS. Extract using `tar -xzf freecompliance-v0.5.0-preview.4-linux-amd64.tar.gz`, then run `./start-cmmc-ui.sh` or `./freecompliance help`. The script attempts to open your browser on the same computer. Review the launch script before running it.

The browser launcher creates a private assessment directory when you choose your CMMC level and system scope. Level 2 initializes with 110 requirement identifiers and no claimed objective completions. To populate assessment objectives, import the separately verified official NIST CSV. Nothing automatically passes or is submitted. The old START-CMMC-DEMO launcher remains available for synthetic CLI demonstration. No credentials or network are required.

## Air-gapped / CUI environments

1. Obtain, inspect and approve the package in a connected staging environment operated under your organization's policies.
2. Check the SHA-256 against the published manifest, inspect source commit and CI build provenance, and use your organization's **trusted software publisher verification process**.
3. **SHA256SUMS is an integrity manifest, not a publisher signature. The published Preview 3 has no independent publisher Authenticode/code-signing certificate and no GitHub build provenance attestation; future releases will add verifiable build attestations without claiming a government ATO.** Do not bypass government or organization software restrictions to install it.
4. Transfer the approved archive and verified metadata using approved removable-media procedures.
5. Recheck the digest after transfer; extract and run core commands locally. The `github` and `prove-github` subcommands are online and must not be used in a true air gap.
6. Keep CUI, filesystem paths, hashes, reports, reviewer names and technical metadata within the authorized enclave.

**Browser interface:** See [LOCAL_UI.md](LOCAL_UI.md) for the exact workflow, token and CSRF protections, privacy boundaries, and limits. The local UI is not approved multi-user authentication or encrypted-at-rest storage.

**Scope limits:** FreeCompliance provides offline readiness and evidence-integrity workspaces, a separately sourced NIST CSV objective importer, draft-only user-entered scoring, and experimental bounded theorem evaluation. It does not supply a **PDF-reconciled and independently reviewed** CMMC objective methodology, independently validated SPRS scoring, official submission, accredited CMMC assessments, or government endorsement.

Windows SmartScreen or Defender may flag an unsigned preview executable. Do **not** bypass security controls. Obtain approval from your IT or security team. Report issues with sanitized, non-sensitive samples only.

For details see [CMMC air-gapped operation](CMMC_AIRGAP.md) and [objective workbench](CMMC_OBJECTIVE_WORKBENCH.md).