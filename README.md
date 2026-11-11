# OpenCMMC

**Free, open-source CMMC readiness.** Organize requirements, collect evidence, and see what still needs review, even on a computer without internet access.

Built by BlockSiFr for defense contractors and security teams. No subscription, account, or cloud service required for the local workbench.

**Latest download:** [OpenCMMC v1.0.0-rc.3](https://github.com/blocksifrdev/OpenCMMC/releases/tag/v1.0.0-rc.3), for the November 11, 2026 release. This is a **release candidate**, not a finished or certified CMMC assessment product.

## Get started

**1. Download the ZIP or TAR.GZ for your computer.**

| Your computer | Download |
| --- | --- |
| Windows 11 ARM64 | [Windows ARM64 ZIP](https://github.com/blocksifrdev/OpenCMMC/releases/download/v1.0.0-rc.3/opencmmc-v1.0.0-rc.3-windows-arm64.zip) |
| Windows x64 (Intel or AMD) | [Windows x64 ZIP](https://github.com/blocksifrdev/OpenCMMC/releases/download/v1.0.0-rc.3/opencmmc-v1.0.0-rc.3-windows-amd64.zip) |
| Linux ARM64 | [Linux ARM64 TAR.GZ](https://github.com/blocksifrdev/OpenCMMC/releases/download/v1.0.0-rc.3/opencmmc-v1.0.0-rc.3-linux-arm64.tar.gz) |
| Linux x64 | [Linux x64 TAR.GZ](https://github.com/blocksifrdev/OpenCMMC/releases/download/v1.0.0-rc.3/opencmmc-v1.0.0-rc.3-linux-amd64.tar.gz) |

[Check your download's SHA-256 fingerprint](docs/INSTALL_V1.md) before running it. Your organization's software-approval rules still apply.

**2. Extract the download and start the workbench.**

- **Windows:** Double-click `START-CMMC-UI.cmd`.
- **Linux:** Run `./start-cmmc-ui.sh` in the extracted folder.

Your browser should open. If it does not, open the **private URL shown in the terminal on that same computer**. Keep the terminal open while using the app. Never share that URL.

**3. Create a workspace and review your evidence.**

Choose CMMC Level 1 or Level 2, name the system being reviewed, record findings, add approved local evidence, and **export a JSON readiness report**. Level 2 objective-level review requires a separately verified official NIST assessment-objective CSV. New controls do not automatically pass.

[Detailed installation](docs/INSTALL_V1.md) · [Browser walkthrough](docs/LOCAL_UI.md)

## What it does

- **Level 1:** Track the 15 basic safeguarding requirements.
- **Level 2:** Track 110 NIST SP 800-171 Rev. 2 requirement identifiers; import assessment objectives from an independently verified NIST CSV.
- **Evidence:** Record reviewer decisions, attach local files, and check file integrity.
- **Reports:** Export findings and unresolved gaps for internal review.
- **Additional tools:** Optional command-line repository scans, experimental proof checks, and draft-only scoring. See [advanced assessment commands](docs/CMMC_ASSESSMENT_NEXT.md).

## Important before you use it

OpenCMMC is a **readiness and evidence-organizing tool**, not a CMMC certification, official assessment, validated SPRS score, or government submission. A document or passing file hash **does not prove** that a security requirement is met.

Evidence is stored locally **without application-level encryption at rest**. Use an approved encrypted device and obtain your organization's authorization **before** handling controlled unclassified information (CUI). The local browser is designed for one computer, not enterprise multi-user access. Packages are portable downloads, not signed Windows installers.

For current CMMC obligations, consult [official DoD CMMC guidance](https://business.defense.gov/cmmc/) and your contract. See [release limitations](docs/RELEASE_GATES_2026-11-11.md) and [security policy](SECURITY.md).

## Optional next steps

After a local assessment, OpenCMMC offers three **optional** ways to go further: [CortexTrace](https://exchange.blocksifr.com/solutions/software-delivery-governance) for AI coding-session governance, [BlockSiFr Control](https://www.blocksifr.com/control) for runtime authority and protected actions, and [BLAKE](https://www.blocksifr.com/) for deeper vulnerability discovery. These are separate services, **not required for the free readiness workflow**. Opening a link never uploads your assessment or starts a scan. [Privacy and service boundaries](docs/OPTIONAL_SERVICES.md).

## Project and support

New OpenCMMC releases use the `opencmmc` command and distribution filenames exclusively.

[Contribute](CONTRIBUTING.md) · [Report a security issue](SECURITY.md) · [Air-gapped use](docs/CMMC_AIRGAP.md) · [Previous releases](https://github.com/blocksifrdev/OpenCMMC/releases)

Free and open source under [Apache 2.0](LICENSE). © 2026 BlockSiFr.
