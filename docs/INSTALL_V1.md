# OpenCMMC v1.0 Release Candidate: Portable Offline Installation

These instructions apply to `opencmmc-v1.0.0-rc.1` archives, **not** the previously released `freecompliance-v0.5` packages.

## Verify before execution

Download the matching `opencmmc-v1.0.0-rc.1` ZIP or TAR.GZ from the project GitHub prerelease, plus `SHA256SUMS`. Validate the cryptographic checksum and review the published commit and CI workflow. SHA-256 integrity is not a publisher signature. Windows binaries are not claimed to be Authenticode-signed. Follow your organization's software-control and removable-media requirements.

## Windows ARM64 or x64

Extract `opencmmc-v1.0.0-rc.1-windows-arm64.zip` on Windows ARM64 or `opencmmc-v1.0.0-rc.1-windows-amd64.zip` on Windows x64. Double-click `START-CMMC-UI.cmd` to open the private local browser UI. Or use PowerShell in the extracted folder:

```powershell
.\opencmmc.exe version
.\opencmmc.exe cmmc ui --data-dir .\private-assessment --port 8765
```

The browser must run on the **same machine** and use the private URL printed by the program. Do not share the private URL or bypass browser security controls.

## Linux ARM64 or x64

Extract the matching TAR.GZ and run:

```sh
./opencmmc version
./start-cmmc-ui.sh
```

The local workbench needs no cloud access, subscription, external script or outbound telemetry. The legacy `freecompliance` executable remains included as a compatible command alias.

## Assessment scope and limitations

This is an offline evidence-readiness workbench, not a complete official assessment or certified CMMC platform. The supplied Level 2 requirement workspace uses 110 identifiers and requires a separately verified official objective source for objective-level review. Reports and reviewer assertions do **not** establish control effectiveness, independently validated scores or an official SPRS submission. Advanced organizational identity, SSP/POA&M generation, authorized C3PAO assessment and official submission are not claimed. Use synthetic test evidence unless data-handling and CUI use are explicitly approved by your organization.

If a stale browser tab causes HTTP 409 after an application restart, reopen the new private local URL. Do not disable CSRF protection.

OpenCMMC is free and Apache-2.0 licensed. See README.md, CMMC_AIRGAP.md, LOCAL_UI.md and SECURITY.md for detailed scope.
