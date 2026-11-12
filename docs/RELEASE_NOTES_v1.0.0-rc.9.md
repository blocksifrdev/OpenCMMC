# OpenCMMC v1.0.0-rc.9 | Browser Origin Metadata Correction

**Release status:** Free open-source prerelease, not CMMC certification, an official CMMC assessment or an SPRS submission.

## Fix
- Correct a specific contradictory check: previously an allowlisted local HTTP Origin could be rejected only because `Sec-Fetch-Site` was `same-site` or `none`. These categories no longer veto a known local Origin with a valid port-specific session cookie and exact form CSRF token.
- `Sec-Fetch-Site: cross-site`, other unrecognized metadata, hostile/unknown Origins, unauthorized sessions, stale CSRF tokens, and unsupported form encodings remain blocked.
- On `ORIGIN_MISMATCH`, show a **Safe request classification** with enum labels for Origin category and Fetch Metadata, never raw untrusted headers, tokens, URLs, cookies, evidence or assessment content.
- Add regression tests for form submissions with local Origin and `same-origin`, `same-site`, `none`, and missing Fetch Metadata, with negative tests for explicit cross-site and hostile requests.

## Upgrade Chromebook
Stop the old OpenCMMC process (Ctrl+C in its terminal). Extract the rc.9 Linux x64 archive over `~/OpenCMMC` without deleting `private-assessment`; verify `./opencmmc version` prints rc.9. Open the freshly printed private URL on the same computer and submit a form.

If it still fails, share **only** the new Safe request classification line and build version. Never share the private URL (including access code), form tokens, cookies, HAR captures, file paths or assessment evidence.

**CI passing does not confirm real-device ChromeOS success.** Issue #49 remains open until a successful browser form submission is observed by the user. Local files are not encrypted at rest by OpenCMMC; use approved storage and avoid introducing CUI without authorization.
