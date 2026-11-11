# OpenCMMC v1.0.0-rc.4 | Private Service Requests

**Release class:** Free, open-source release candidate. Not feature-complete general availability or an official CMMC assessment.

## What's new

- Users with a local CMMC assessment may now choose **Prepare a private request** for CortexTrace, BlockSiFr Control or BLAKE.
- The service-specific screen explains the intended next step, retains the local session boundary and requires the user to confirm before generating an optional file.
- The downloaded JSON intent contains an allowlisted service choice, predefined purpose, random local request ID and timestamp. It contains **no assessment findings, CUI, scope, evidence, repository data, file paths or credentials**.
- Preparing the file does not create an account, call a remote API, activate runtime controls, start a coding session or launch a scan.
- Original CMMC assessment functions and JSON readiness export remain free, independent and offline.
- Tests cover expired form tokens, origin restrictions, missing consent, unknown services, extra form fields and absence of assessment-data leakage.

## Install

Download the correct Windows ARM64/x64 or Linux ARM64/x64 archive, check its SHA-256 against the supplied manifest, extract, then run START-CMMC-UI.cmd on Windows or start-cmmc-ui.sh on Linux. See [installation](INSTALL_V1.md).

## Next integration stage

Actual customer-authorized commercial service activation remains a separate milestone. It requires tenant authentication, scope and ownership proof, explicit data inventory and consent, policy enforcement, and valid service-side receipts. See [Issue #42](https://github.com/blocksifrdev/OpenCMMC/issues/42).

This candidate provides no government certification, independent CMMC assessment, validated SPRS score or official submission. Windows binaries are portable and not claimed to carry a publisher Authenticode signature.
