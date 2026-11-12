# Download OpenCMMC

**Current public version:** [v1.0.0-rc.6, local session and form diagnostics hotfix](https://github.com/blocksifrdev/OpenCMMC/releases/tag/v1.0.0-rc.6). **Not** v1.0.0 general availability or official CMMC certification.

| Computer | Download |
| --- | --- |
| Windows ARM64 | [ZIP](https://github.com/blocksifrdev/OpenCMMC/releases/download/v1.0.0-rc.6/opencmmc-v1.0.0-rc.6-windows-arm64.zip) |
| Windows x64 (Intel/AMD) | [ZIP](https://github.com/blocksifrdev/OpenCMMC/releases/download/v1.0.0-rc.6/opencmmc-v1.0.0-rc.6-windows-amd64.zip) |
| Linux ARM64 | [TAR.GZ](https://github.com/blocksifrdev/OpenCMMC/releases/download/v1.0.0-rc.6/opencmmc-v1.0.0-rc.6-linux-arm64.tar.gz) |
| Linux x64 | [TAR.GZ](https://github.com/blocksifrdev/OpenCMMC/releases/download/v1.0.0-rc.6/opencmmc-v1.0.0-rc.6-linux-amd64.tar.gz) |

**Next:** Download [SHA256SUMS](https://github.com/blocksifrdev/OpenCMMC/releases/download/v1.0.0-rc.6/SHA256SUMS) and verify your download. On **Windows**, right-click the ZIP → **Extract All** → **Extract**, then open the new folder and double-click `START-CMMC-UI.cmd`. On **Linux**, extract the `.tar.gz` file and run `./start-cmmc-ui.sh` from the extracted folder. See [step-by-step unzip instructions for both platforms](INSTALL_V1.md#2-unzip-extract-the-downloaded-file).

Packages were tested on native runners for all four listed architectures. This checks installation and basic workflows, **not** complete security or compliance. They are portable archives, not signed native installers.

**Older releases:** Earlier preview downloads remain archived under [GitHub Releases](https://github.com/blocksifrdev/OpenCMMC/releases). Their original filenames are not renamed. Use the current OpenCMMC release for new installations. See [release integrity](RELEASE_INTEGRITY.md).

**Sensitive systems:** Do not bypass endpoint controls. Follow your organization's approval and data-handling requirements before using this software with CUI. See [offline/CUI guidance](CMMC_AIRGAP.md) and [current limitations](RELEASE_GATES_2026-11-11.md).
