# Install OpenCMMC

These instructions cover **[v1.0.0-rc.4](https://github.com/blocksifrdev/OpenCMMC/releases/tag/v1.0.0-rc.4)**, a public release candidate. No coding, account, subscription, or internet connection is needed **after** obtaining the package.

## 1. Download and verify

[Choose your Windows or Linux download](DOWNLOADS.md). Get the matching archive and the [SHA256SUMS file](https://github.com/blocksifrdev/OpenCMMC/releases/download/v1.0.0-rc.4/SHA256SUMS).

A SHA-256 checksum is a **file fingerprint**. Compare the fingerprint of your download with the line for that exact file in `SHA256SUMS`:

- **Windows PowerShell (example for ARM64):** `Get-FileHash .\opencmmc-v1.0.0-rc.4-windows-arm64.zip -Algorithm SHA256`
- **Linux terminal (example for x64):** `sha256sum opencmmc-v1.0.0-rc.4-linux-amd64.tar.gz`

Run the command in the folder containing your download. The two results must match. A checksum does **not** establish publisher identity; follow your organization's software-approval process. Windows packages are not claimed to be Authenticode-signed.

## 2. Unzip (extract) the downloaded file

### Windows 11 (ARM64 or x64)

1. Open **File Explorer → Downloads**.
2. Right-click the downloaded `.zip` file and choose **Extract All**. On some systems, select the ZIP first and choose **Extract all** at the top.
3. Choose an approved destination folder (or keep the suggested one) and click **Extract**.
4. Open the **newly extracted folder**. You should see `opencmmc.exe` and `START-CMMC-UI.cmd`. **Do not launch files directly from the compressed ZIP.**
5. Double-click **`START-CMMC-UI.cmd`**. Keep its command window open while you use OpenCMMC.

**No ZIP software or PowerShell commands are needed** for this Windows method. If Windows or your organization's security controls block the app, follow your software-approval process. Do not bypass security warnings.

### Linux (ARM64 or x64)

In your file manager, right-click the `.tar.gz` download and select **Extract Here**, then open the new extracted folder. Alternatively, in a terminal in your **Downloads** directory, use these commands for **Linux x64**:

```sh
mkdir -p OpenCMMC
tar -xzf opencmmc-v1.0.0-rc.4-linux-amd64.tar.gz -C OpenCMMC
cd OpenCMMC
./start-cmmc-ui.sh
```

For **Linux ARM64**, replace `linux-amd64` in the `tar` command with `linux-arm64`. The other commands are the same. This `tar` command extracts the files; it does not install a system-wide application.

### Open your local workbench

Your browser should open automatically. If not, open the **private local URL printed in the command window** in a browser on **the same computer**. Keep the window open, and do not share or bookmark the private URL.

**Important:** OpenCMMC keeps assessment data in a `private-assessment` subfolder beside the launcher. Keep the extracted folder in an approved, access-controlled location, and do not delete it if you need to retain your assessment. Local evidence is not encrypted by OpenCMMC.

## 3. Review and export

1. Create an assessment and choose **Level 1** or **Level 2**.
2. Select a requirement, add reviewer notes and approved evidence (up to **10 MiB per file**).
3. For Level 2 **objective-by-objective** review, import a separately verified official NIST SP 800-171A CSV. The default workspace is not a complete objective assessment.
4. Export your **JSON readiness report** for internal review.

**If you see a form-token or session error:** Close old app tabs and open the **new private URL** printed after the latest launch. Never disable the browser's security controls.

## Know the limits

This is a local readiness workbench, **not** certified CMMC software, an accredited assessment, verified SPRS scoring, or government submission. Local evidence is **not encrypted by the app**; use an approved device and obtain permission before introducing CUI. It is not a multi-user enterprise system.

[Browser walkthrough](LOCAL_UI.md) · [Air-gapped/CUI guidance](CMMC_AIRGAP.md) · [Security policy](../SECURITY.md)
