# OpenCMMC v1.0.0-rc.6 | Local Session Isolation and Form Diagnostics

**Release class:** Public release candidate, not general availability or any form of CMMC certification.

## Security and ChromeOS usability hotfix

- **Port-scoped session cookies.** Browsers share cookies across ports on the same host. Two local instances previously used the identical `oc_session` cookie name. Different listeners now use distinct names such as `oc_session_p8765`, preventing cross-port session collisions.
- **Actionable error causes.** Instead of one `invalid form token or request origin` message, the workbench clearly identifies `FORM_TOKEN_EXPIRED`, `ORIGIN_MISMATCH`, or `UNSUPPORTED_FORM_TYPE`, with a safe recovery path for each.
- **Visible application version.** Both the terminal startup banner and the browser workbench show the running version to detect old or conflicting processes.
- **Security checks retained.** Session authorization, exact Origin equality, anti-CSRF token checks, loopback-only binding, no-referrer response and offline data handling remain enforced. Rejected POSTs cannot update assessment data or reveal tokens.
- Regression tests cover simultaneous instances on distinct ports, private session creation, valid submissions, origin mismatch, missing/expired token, unsupported form type and no user-data reflection.

## Chromebook update

Stop old OpenCMMC instances, unpack the rc.6 Linux x64 archive into your existing app directory without deleting `private-assessment`, and run `./start-cmmc-ui.sh`. Open the newly printed private URL from the same computer. Verify `./opencmmc version` reports `v1.0.0-rc.6`.

**No claim of hardware validation on every Chromebook.** If a form still fails, the new diagnostic class will identify the remaining condition without sharing sensitive data.

OpenCMMC remains free for offline CMMC readiness review. Optional local planning requests for CortexTrace, BlockSiFr Control and BLAKE do not activate paid services or transfer assessment contents. Local evidence is not encrypted by OpenCMMC.
