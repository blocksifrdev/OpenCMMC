# Use the OpenCMMC browser workbench

OpenCMMC runs **on your computer**. The browser screen is served locally at `127.0.0.1`; it does not require a cloud account or send evidence to BlockSiFr.

## Open it

- **Windows:** In Downloads, right-click the ZIP → **Extract All** → **Extract**. Open the new folder and double-click `START-CMMC-UI.cmd`. Do not run the launcher inside the ZIP.
- **Linux:** Right-click the `.tar.gz` file → **Extract Here**, open the extracted folder, then run `./start-cmmc-ui.sh` in a terminal. For `tar` command examples, see the [installation guide](INSTALL_V1.md#2-unzip-extract-the-downloaded-file).

The browser usually opens for you. Otherwise, open the **private URL printed in the command window** using a browser **on the same computer**. Keep that window open. Do **not** share or bookmark the private URL.

[Download](DOWNLOADS.md) · [Installation and file verification](INSTALL_V1.md)

## Make your first assessment

1. Choose **Level 1** (15 requirement identifiers) or **Level 2** (110 requirement identifiers). Name the system you're reviewing.
2. For **Level 2 objective-level** review, obtain the official NIST SP 800-171A supplemental CSV through an approved source, verify its SHA-256 fingerprint, and import it in the workbench. A new workspace does **not** include complete objectives or pass by default.
3. Choose a requirement, record the reviewer decision and explanation, and attach approved local evidence. Maximum file size: **10 MiB each**.
4. Review the status and export the **JSON report**. This is an internal readiness record, **not** official CMMC certification or an SPRS submission.

No coding or editing JSON is needed for these browser steps. Command-line scoring remains **draft-only**; it is not part of the browser workbench.

## Optional next steps (only after you create an assessment)

The bottom of the workbench offers **CortexTrace** for AI coding-session change review, **BlockSiFr Control** for runtime enforcement, and **BLAKE** for deeper vulnerability discovery. These are separate services. **Prepare a private request** opens a local confirmation screen and downloads a JSON planning file only after you agree. The file contains no evidence, scope, report, file paths, findings, cookies or CUI. A separate information link opens an external website only if you click; that action needs internet access. Neither step launches a coding session, policy enforcement or a scan. The free workflow stays available offline.

[How to prepare a private service request](SERVICE_HANDOFF.md) · [Optional services](OPTIONAL_SERVICES.md)

## If the browser rejects a form

An older tab may have an expired session token after the app restarts.

ChromeOS may rewrite the Host header when forwarding a browser request into the Linux container. Starting in rc.8, OpenCMMC also accepts browser-attested same-origin HTML form navigations with exact session and CSRF validation, even when the forwarded Host differs from the browser-visible Origin. Requests with cross-site Fetch Metadata remain forbidden.

**rc.9 note:** `Sec-Fetch-Site: same-site` and `none` are not grounds by themselves to reject a fully authenticated form from an explicitly allowlisted local Origin. The error page now shows safe category labels for Origin and Fetch Metadata when `ORIGIN_MISMATCH` persists, without reflecting raw Origin text, private access values or evidence. Never share a HAR or request body.

If a form is rejected, check the displayed **diagnostic label**. `FORM_TOKEN_EXPIRED` means reload the current assessment and retry; `ORIGIN_MISMATCH` means the browser submitted from an origin outside the allowlisted local HTTP addresses or the expected port; ChromeOS `localhost` and `penguin.linux.test` are accepted in rc.7; `UNSUPPORTED_FORM_TYPE` means the browser did not use a supported HTML form encoding. Never disable security checks. If a tab instead says **Reconnect OpenCMMC**, the browser has no valid session cookie. Use the entire new private URL printed in the active terminal; it must start with `http://127.0.0.1:8765/?access=`. The access code is a secret. Never share it.

1. Close older OpenCMMC browser tabs.
2. Open the **new private URL** from the currently running command window.
3. If a **Refresh your assessment session** page appears, use its reload option and retry. The rejected form was not saved.

If it still fails, restart the launcher, confirm only one app is using port 8765, and report the browser/Windows version and error message. **Never share the private URL, cookies, or real assessment evidence. Do not disable CSRF protection.**

**Chromebook Linux workaround:** If ChromeOS host Chrome still gives `ORIGIN_MISMATCH`, use [Firefox ESR installed directly in the Crostini Linux environment](INSTALL_V1.md#chromebook-open-opencmmc-without-chromeos-browser-forwarding). This avoids the ChromeOS forwarding path; do not change CSRF validation or delete your assessment folder. The Chromebook-specific host-browser defect is [still open](https://github.com/blocksifrdev/OpenCMMC/issues/49).

## Protect your information

- Evidence stays in a local private folder but is **not encrypted at rest by OpenCMMC**. Use approved device/storage encryption.
- The session URL is a secret. The workbench is **not** a multi-user authentication system or an approved CUI security boundary.
- A file hash checks integrity; it cannot establish that a security requirement has been met.
- Use synthetic evidence until your organization approves the software, storage, and handling of any controlled data.

[Air-gapped/CUI guidance](CMMC_AIRGAP.md) · [Technical threat model](THREAT_MODEL.md) · [Current limitations](RELEASE_GATES_2026-11-11.md)
