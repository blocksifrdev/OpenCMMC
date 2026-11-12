# Install OpenCMMC

These instructions cover **[v1.0.0-rc.6](https://github.com/blocksifrdev/OpenCMMC/releases/tag/v1.0.0-rc.6)**, a public release candidate. No coding, account, subscription, or internet connection is needed **after** obtaining the package.

## 1. Download and verify

[Choose your Windows or Linux download](DOWNLOADS.md). Get the matching archive and the [SHA256SUMS file](https://github.com/blocksifrdev/OpenCMMC/releases/download/v1.0.0-rc.6/SHA256SUMS).

A SHA-256 checksum is a **file fingerprint**. Compare the fingerprint of your download with the line for that exact file in `SHA256SUMS`:

- **Windows PowerShell (example for ARM64):** `Get-FileHash .\opencmmc-v1.0.0-rc.6-windows-arm64.zip -Algorithm SHA256`
- **Linux terminal (example for x64):** `sha256sum opencmmc-v1.0.0-rc.6-linux-amd64.tar.gz`

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
tar -xzf opencmmc-v1.0.0-rc.6-linux-amd64.tar.gz -C OpenCMMC
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

**If the browser says “Reconnect OpenCMMC,” or “Open the private local URL”:** Open the full private URL printed by the current terminal session. It starts with `http://127.0.0.1:8765/?access=` and contains a unique access code. The browser will establish a local session and remove the code from the address bar. Do not use a bookmark, switch to localhost, or share the code. If the terminal stopped, restart `./start-cmmc-ui.sh` from the extracted folder and use its newly printed URL. Do not delete the `private-assessment` directory.

**If a form-token or origin error persists:** Close all earlier OpenCMMC tabs; use the fresh private URL from the same computer. Never disable the browser's security controls. Version rc.6 also isolates session cookies by port and provides safe form diagnostics. Browser responses distinguish `FORM_TOKEN_EXPIRED`, `ORIGIN_MISMATCH` and `UNSUPPORTED_FORM_TYPE`, rather than presenting one generic error. If your error persists, report only the diagnostic label and the local browser address up to (not including) any `?` character. Do not share session secrets, file paths, CUI or evidence.

**To verify which binary is installed:** Run `cd ~/OpenCMMC && ./opencmmc version` in the Linux terminal. The browser workbench also shows the build version. If port 8765 is already in use, first identify the listener with `ss -ltnp '( sport = :8765 )'`. Stop only a confirmed old OpenCMMC process, and launch the new version. These operations do not require deleting your assessment directory.

## Know the limits

This is a local readiness workbench, **not** certified CMMC software, an accredited assessment, verified SPRS scoring, or government submission. Local evidence is **not encrypted by the app**; use an approved device and obtain permission before introducing CUI. It is not a multi-user enterprise system.

[Browser walkthrough](LOCAL_UI.md) · [Air-gapped/CUI guidance](CMMC_AIRGAP.md) · [Security policy](../SECURITY.md)
