# FreeCompliance local browser workbench (engineering preview)

**Purpose:** Run a CMMC readiness workflow with clicks and forms instead of editing JSON by hand. The entire interface is served by the installed Go executable on **127.0.0.1 only**, using standard-library Go. It requires **no account, internet connection, cloud API, JavaScript, CDN, external font or telemetry**.

## Start on Windows

Extract the approved FreeCompliance ZIP into a local folder and double-click `START-CMMC-UI.cmd`. The command window starts a private local server. It automatically tries to open your default browser and also prints a private session URL beginning with `http://127.0.0.1:8765/?access=...`. Open that URL in a browser **on the same device** and keep the command window running. Close the window or press Ctrl+C to stop. The command-line version is:

```powershell
.\freecompliance.exe cmmc ui --data-dir .\private-assessment --port 8765
```

## Start on Linux

Extract the approved TAR archive, then:

```sh
./start-cmmc-ui.sh
```

Or:

```sh
./freecompliance cmmc ui --data-dir ./private-assessment --port 8765
```

Open the printed private URL on the **same Linux computer**.

**Current package note:** The previously published Preview 2 archives do not contain the browser UI. Use an independently verified release containing this source before following the GUI launcher instructions.

## Complete customer workflow available in this preview

1. Create a **Level 1** requirement-level workspace (15 safeguarding identifiers), or a **Level 2** objective-level workspace (110 requirement identifiers). A new Level 2 workspace starts with **zero objectives**, and therefore cannot be presented as passed or complete.
2. For Level 2, obtain the official June 2018 NIST SP 800-171A supplemental CSV in an approved staging environment. Independently establish a trusted SHA-256; import it using the browser file picker and digest field. This structural import does **not** independently reconcile the CSV to the normative NIST PDF.
3. Browse requirements without manual JSON editing. For each item, record a reviewer name, method (where applicable), assessment decision, and rationale. This remains a **reviewer assertion** until independently tested.
4. Attach approved local evidence files of up to 10 MiB each. The app stores files in its private `evidence/` directory, records SHA-256 and capture time, and does not read files outside the controlled assessment directory. Files are **copied locally** by the user's own browser to the local loopback process. No remote server receives uploads.
5. Inspect conservative computed statuses for each requirement. All unresolved or missing evidence remains unproven. The UI intentionally does not present a compliance percentage.
6. Download a JSON report of the evaluated findings for **approved internal review**. Export is initiated by the user. It is not an SPRS submission or an OSCAL conformance claim.

## Data protection and threat model

- Binds exclusively to IPv4 loopback `127.0.0.1`. Arbitrary bind addresses are not configurable.
- Requires a random one-session 256-bit access token, set as an HttpOnly SameSite=Strict browser cookie, plus a matching anti-CSRF form token.
- Rejects unexpected Host headers and cross-origin POSTs; sends `Cache-Control: no-store`, a restrictive CSP, `X-Frame-Options: DENY`, and `Referrer-Policy: no-referrer`.
- No JavaScript, outbound HTTP, remote image, font, or code dependency.
- Evidence is never served back over an evidence-download endpoint. Export is a user-initiated JSON report and can contain reviewer names, source paths and evidence hashes.
- Storage folders are mode `0700`, workspaces and evidence files are `0600` on compatible Unix systems; the app refuses symlink workspace files and symlink evidence directories. File access and copying are subject to the OS user's privileges.
- The private browser URL is a local session secret. Do not share it in logs, screenshots, support requests, or messages. Any malicious process running under the same local user may have access to sensitive files or session material; this is **not** a multi-user authentication or CUI-approved security boundary.
- Evidence files are copied into a local directory **without application-level encryption at rest**. Use an already approved encrypted device, filesystem, enclave, and backup mechanism where appropriate.
- Before connecting sensitive systems, obtain organization-specific CUI handling approval, IT authorization, system boundary review, and appropriately qualified assessment review.

## Deliberate limits

The GUI is **an offline readiness workbench**, not a complete independent assessment methodology. It does not independently verify all NIST PDF procedures, provide proven cross-system control effectiveness, official SPRS submission, C3PAO accreditation, OSCAL-certified export, government endorsement, or any government ATO.

The background CLI's draft scoring remains outside the graphical UI until independent verification and product security acceptance. FreeCompliance is always free and Apache-2.0 licensed.
