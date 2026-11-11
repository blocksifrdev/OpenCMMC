# OpenCMMC v1.0.0-rc.5 | Chromebook Session Repair

**Release class:** Public, free OpenCMMC release candidate. Not general availability, CMMC certification, SPRS scoring or an official government submission.

## Session repair

- Adjust the loopback browser-session cookie to **SameSite=Lax**, addressing a Chromium/ChromeOS cross-context launch scenario in which a new Strict cookie may be omitted from the first redirect.
- Keep the private access-code handshake, loopback-only Host checking, HttpOnly cookie, exact-Origin check on POST, and antiforgery token. This change does not allow unauthenticated access or bypass security checks.
- When a browser has no valid session, show a clear **Reconnect OpenCMMC** page with instructions for opening the **newly issued** private URL from the local terminal. The server never displays the access code in an error response.
- Add regression coverage for the bootstrap cookie, redirect, recovered workbench session, hostile Origin and missing CSRF.
- Update the README and installation troubleshooting specifically for Chromebook users.

## Existing capabilities

Offline local CMMC Level 1/2 readiness review and evidence capture remain free, with optional user-approved local planning requests for CortexTrace, BlockSiFr Control and BLAKE. No assessment content is submitted by these requests; they do not activate paid services.

## Security and data

Your browser must be on the same computer as the local OpenCMMC process. Do not share or bookmark the private URL or disable security protections. Local evidence is not encrypted at rest by the application. Before using CUI, obtain your organization's authorization and use an approved encrypted device.

For verified archives, SHA256SUMS, SPDX metadata, supported platform downloads and extraction instructions, see [Installation](INSTALL_V1.md).
