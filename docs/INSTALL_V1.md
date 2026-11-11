# Install OpenCMMC

These instructions cover **[v1.0.0-rc.4](https://github.com/blocksifrdev/OpenCMMC/releases/tag/v1.0.0-rc.4)**, a public release candidate. No coding, account, subscription, or internet connection is needed **after** obtaining the package.

## 1. Download and verify

[Choose your Windows or Linux download](DOWNLOADS.md). Get the matching archive and the [SHA256SUMS file](https://github.com/blocksifrdev/OpenCMMC/releases/download/v1.0.0-rc.4/SHA256SUMS).

A SHA-256 checksum is a **file fingerprint**. Compare the fingerprint of your download with the line for that exact file in `SHA256SUMS`:

- **Windows PowerShell (example for ARM64):** `Get-FileHash .\opencmmc-v1.0.0-rc.4-windows-arm64.zip -Algorithm SHA256`
- **Linux terminal (example for x64):** `sha256sum opencmmc-v1.0.0-rc.4-linux-amd64.tar.gz`

Run the command in the folder containing your download. The two results must match. A checksum does **not** establish publisher identity; follow your organization's software-approval process. Windows packages are not claimed to be Authenticode-signed.

## 2. Open the app

**Windows:** Extract the ZIP, open the extracted folder, and double-click `START-CMMC-UI.cmd`.

**Linux:** Extract the TAR.GZ, open a terminal in the extracted folder, and run:

```sh
./start-cmmc-ui.sh
```

Leave the terminal/command window open. Open the **private local URL it prints** in a browser on **the same computer** if the browser does not open automatically. Do not copy that URL to another device or share it.

## 3. Review and export

1. Create an assessment and choose **Level 1** or **Level 2**.
2. Select a requirement, add reviewer notes and approved evidence (up to **10 MiB per file**).
3. For Level 2 **objective-by-objective** review, import a separately verified official NIST SP 800-171A CSV. The default workspace is not a complete objective assessment.
4. Export your **JSON readiness report** for internal review.

**If you see a form-token or session error:** Close old app tabs and open the **new private URL** printed after the latest launch. Never disable the browser's security controls.

## Know the limits

This is a local readiness workbench, **not** certified CMMC software, an accredited assessment, verified SPRS scoring, or government submission. Local evidence is **not encrypted by the app**; use an approved device and obtain permission before introducing CUI. It is not a multi-user enterprise system.

[Browser walkthrough](LOCAL_UI.md) · [Air-gapped/CUI guidance](CMMC_AIRGAP.md) · [Security policy](../SECURITY.md)
