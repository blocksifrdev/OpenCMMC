# Use the OpenCMMC browser workbench

OpenCMMC runs **on your computer**. The browser screen is served locally at `127.0.0.1`; it does not require a cloud account or send evidence to BlockSiFr.

## Open it

- **Windows:** Extract the ZIP and double-click `START-CMMC-UI.cmd`.
- **Linux:** Extract the TAR.GZ and run `./start-cmmc-ui.sh`.

The browser usually opens for you. Otherwise, open the **private URL printed in the command window** using a browser **on the same computer**. Keep that window open. Do **not** share or bookmark the private URL.

[Download](DOWNLOADS.md) · [Installation and file verification](INSTALL_V1.md)

## Make your first assessment

1. Choose **Level 1** (15 requirement identifiers) or **Level 2** (110 requirement identifiers). Name the system you're reviewing.
2. For **Level 2 objective-level** review, obtain the official NIST SP 800-171A supplemental CSV through an approved source, verify its SHA-256 fingerprint, and import it in the workbench. A new workspace does **not** include complete objectives or pass by default.
3. Choose a requirement, record the reviewer decision and explanation, and attach approved local evidence. Maximum file size: **10 MiB each**.
4. Review the status and export the **JSON report**. This is an internal readiness record, **not** official CMMC certification or an SPRS submission.

No coding or editing JSON is needed for these browser steps. Command-line scoring remains **draft-only**; it is not part of the browser workbench.

## If the browser rejects a form

An older tab may have an expired session token after the app restarts.

1. Close older OpenCMMC browser tabs.
2. Open the **new private URL** from the currently running command window.
3. If a **Refresh your assessment session** page appears, use its reload option and retry. The rejected form was not saved.

If it still fails, restart the launcher, confirm only one app is using port 8765, and report the browser/Windows version and error message. **Never share the private URL, cookies, or real assessment evidence. Do not disable CSRF protection.**

## Protect your information

- Evidence stays in a local private folder but is **not encrypted at rest by OpenCMMC**. Use approved device/storage encryption.
- The session URL is a secret. The workbench is **not** a multi-user authentication system or an approved CUI security boundary.
- A file hash checks integrity; it cannot establish that a security requirement has been met.
- Use synthetic evidence until your organization approves the software, storage, and handling of any controlled data.

[Air-gapped/CUI guidance](CMMC_AIRGAP.md) · [Technical threat model](THREAT_MODEL.md) · [Current limitations](RELEASE_GATES_2026-11-11.md)
