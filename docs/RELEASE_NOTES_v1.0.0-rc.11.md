# OpenCMMC v1.0.0-rc.11 | Chromium Browser Form Reliability

**Status:** Public free/open-source release candidate for offline CMMC readiness.
Not certified CMMC assessment software, SPRS submission, security accreditation,
CUI authorization or `v1.0.0` general availability.

## P0 correction: browser form submission

A real Chromium end-to-end test reproduced `ORIGIN_MISMATCH` on a normal
authenticated local HTML form. `Referrer-Policy: no-referrer` made the browser
send `Origin: null`. The app correctly rejected an opaque Origin but therefore
blocked legitimate user work.

- Authenticated assessment and local handoff pages now use
  `Referrer-Policy: same-origin`.
- The private access-code bootstrap redirect, rejected/unauthenticated
  responses and report downloads retain `no-referrer` to protect sensitive
  URLs.
- External service sites receive no Referer; all opt-in links retain explicit
  opt-in and no automatic transfer.
- Loopback-only binding, valid port-scoped cookies, exact CSRF, cross-origin
  blocking, and input validation remain unchanged.
- No unauthenticated or opaque-origin write is accepted by this change.

## Additional verified release gates

The actual Chromium test now runs in general CI **and** before a release can be
published. It exercises Level 1 assessment creation, a reviewer gap, local
synthetic file evidence, JSON export, hostile cross-site rejection, a forwarded
IPv6-to-IPv4 browser Origin, and Level 2 creation/export.

Baseline validation:
- Go tests and Chromium:
  https://github.com/blocksifrdev/OpenCMMC/actions/runs/44378443450
- Windows x64/ARM64, Linux x64/ARM64, archive integrity, Chromium:
  https://github.com/blocksifrdev/OpenCMMC/actions/runs/44378443601

The previous rc.10 recovery fix is retained: valid workspaces orphaned by an
interrupted first write are recognized without being overwritten. Ambiguous or
corrupt state remains blocked.

## Install

Obtain the matching portable archive, verify the release's SHA256SUMS and
publisher provenance under your organization's software policy, and extract
into an approved location. Preserve your existing `private-assessment`
directory and verify `opencmmc version` reports `v1.0.0-rc.11` before
launching. Do not disclose the printed private access URL or any CUI.

## Still required for a production-authorization claim

- Verify a successful local POST on the customer's real Chromebook; [issue
  #49](https://github.com/blocksifrdev/OpenCMMC/issues/49) remains open.
- Independent security review of local storage, browser, origin, session,
  filesystem and update-chain boundaries.
- Qualified independent CMMC baseline/objective and scoring reconciliation.
- Organization-specific approval before working with CUI, including approved
  device encryption, access control, backup/retention and authorized scope.

This candidate is not an official assessment and is not approved for CUI by
virtue of being offline. Follow current official DoD requirements and each
actual contract; CMMC Phase II was suspended in July 2026 but contractor
cybersecurity obligations remain.
