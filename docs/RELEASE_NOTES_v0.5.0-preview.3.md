# FreeCompliance v0.5.0 Preview 3: The Local Browser Workbench

**Publicly accessible free software. Official Veterans Day service launch: November 11, 2026.**

Preview 3 is the first self-contained, offline-first **browser-based CMMC readiness workflow**. It builds on Preview 2 and lets small defense suppliers review CMMC requirements and evidence without manually editing JSON or running individual assessment commands.

## Start with the browser

1. Download the ZIP or TAR.GZ for your operating system and architecture.
2. Verify SHA-256 and obtain your organization's permission to install on the target device.
3. Extract the archive.
4. On Windows, double-click **START-CMMC-UI.cmd**. On Linux, run **./start-cmmc-ui.sh**. These launchers attempt to open the default browser automatically.
5. If the browser fails to open automatically, copy the private URL shown in the terminal to a browser on **the same computer**.
6. Choose CMMC Level 1 or 2, name your system boundary, and work through the requirements.

## What works offline today

- Create a Level 1 (15 items) or Level 2 (110 requirement identifiers) local evidence workspace through a browser.
- Import the official NIST SP 800-171A supplemental CSV for Level 2 with an independently verified expected SHA-256. This is **not** independent PDF reconciliation.
- View requirement/objective text from the imported source, record reviewer decisions, select EXAMINE, INTERVIEW or TEST, and add rationale.
- Attach local evidence files up to 10 MiB; record hashes, timestamps and file-integrity results without cloud upload.
- See gaps, unresolved objectives, evidence review-ready findings, and controls with no official objective lists.
- Download a JSON readiness report for internal review.
- Keep all existing CLI functionality, including draft-only scoring worksheets and the experimental bounded theorem evaluator.

## Trust and deployment boundaries

The HTTP server listens only on 127.0.0.1, requires a random private session token, sets HttpOnly SameSite=Strict cookies and CSRF tokens, and provides CSP and Host/Origin protections. It uses no external scripts, images, font services, or telemetry.

**This is a portable executable with a local web interface, not an MSI installer or a government-approved CUI handling system.** There is no independent publisher signature or organization-level ATO. Integrity checksums are not signatures. The workspace stores sensitive evidence unencrypted at the application layer; use approved device encryption and enclave restrictions.

## Not claimed

The UI is **not** CMMC certification, authoritative cross-system effectiveness testing, accredited third-party assessment, official SPRS scoring/submission, independent PDF-reconciled objective completeness, schema-validated OSCAL assessment export, or government endorsement. All incomplete objective lists and unsupported effectiveness assertions remain unproven.

## Downloads

- `freecompliance-v0.5.0-preview.3-windows-amd64.zip`
- `freecompliance-v0.5.0-preview.3-windows-arm64.zip`
- `freecompliance-v0.5.0-preview.3-linux-amd64.tar.gz`
- `freecompliance-v0.5.0-preview.3-linux-arm64.tar.gz`
- `SHA256SUMS`, `SHA256SUMS.sha256`

See [local browser UI documentation](https://github.com/blocksifrdev/freecompliance/blob/main/docs/LOCAL_UI.md) and [installation guide](https://github.com/blocksifrdev/freecompliance/blob/main/docs/DOWNLOADS.md).

**Built in gratitude. Given in service.**
