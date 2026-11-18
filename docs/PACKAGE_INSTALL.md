OpenCMMC @OPENCMMC_VERSION@ | Offline Installation

This text applies to the exact bundled version, not any older RC archive.
No coding, subscription or account is needed to run the local workbench.

BEFORE INSTALLING
- Confirm your operating-system architecture matches this ZIP/TAR.GZ.
- Verify its SHA256 against the separately obtained release SHA256SUMS file.
- Also authenticate the separately published GitHub/Sigstore provenance
  and SBOM signatures against the canonical blocksifrdev/OpenCMMC repository
  using an independently trusted certificate root.
- SHA256 alone is not publisher authentication.
- Get your organization's software approval and approved encrypted storage.
  OpenCMMC does not itself encrypt assessment or backup files.

WINDOWS 11 (x64 OR ARM64)
1. In Downloads, right-click the matching OpenCMMC archive and Extract All.
2. Open the actual extracted folder (not the ZIP viewer).
3. Double-click START-CMMC-UI.cmd and keep its terminal window open.
4. The browser opens the private local workbench. If it does not, open the
   complete private http://127.0.0.1 URL printed only on this machine.
5. Never send, screenshot, share or bookmark the private access URL.
6. Confirm the running terminal or executable prints:
      OpenCMMC @OPENCMMC_VERSION@

LINUX (x64 OR ARM64, INCLUDING SUPPORTED CHROMEBOOK LINUX)
1. Extract the matching .tar.gz on approved local storage.
2. Open a terminal inside the extracted folder and run:
      ./opencmmc version
      ./start-cmmc-ui.sh
3. Confirm the version is OpenCMMC @OPENCMMC_VERSION@.
4. Open only the full private URL printed by your own local terminal.
5. If your Chromebook's Chrome browser rejects a valid form POST, use the
   approved Linux-native Firefox method before changing any security
   configuration. Never disable Origin validation, CSRF or host checks.

WORKBENCH
- Create a scoped Level 1 or Level 2 readiness assessment, record human
  reviewer decisions and locally approved evidence, and export JSON.
- Level 2 begins with 110 requirement identifiers. Import verified NIST
  assessment objectives before claiming objective-level completeness.
- Missing objectives, contradictory facts and unsupported evidence remain
  unresolved. A linked file or matching digest is not technical proof.

BACKUP AND RECOVERY (STOP THE WORKBENCH BEFORE MAKING A BACKUP)
- Use a separate approved encrypted destination, never the live workspace.
- The ZIP created by this command is plaintext and not encrypted; protect it.
      ./opencmmc cmmc backup --data-dir ./private-assessment --archive /approved/encrypted-backups/assessment.zip
      ./opencmmc cmmc verify-backup --archive /approved/encrypted-backups/assessment.zip
      ./opencmmc cmmc restore-backup --archive /approved/encrypted-backups/assessment.zip --out-dir ./restored-assessment
- On Windows, replace ./opencmmc with .\opencmmc.exe.
- A restore destination must not exist and must not overwrite an existing
  assessment. Backups require a trusted, separately approved encrypted drive.

SECURITY AND SCOPE
The browser listens on loopback and has local sessions and CSRF protection.
This app is not a managed multiuser compliance system, not a certified CMMC
assessor, not an SPRS interface, not a government ATO, and does not authorize
processing or storage of CUI. Organizations must independently approve
devices, controls, backup, permissions and security boundaries.

For updates, verified release assets and SECURITY.md:
https://github.com/blocksifrdev/OpenCMMC
