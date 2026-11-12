# OpenCMMC v1.0.0-rc.8 | ChromeOS Forwarded Same-Origin Forms

**Public release candidate, free, offline local CMMC readiness only.** Not a certification, an official CMMC assessment, or SPRS submission.

## Browser form rejection repair

After rc.7, a Chromebook user continued to report `ORIGIN_MISMATCH`. A fixed list of browser Origin hostnames was inadequate for the actual ChromeOS forwarding path.

- The server now also accepts a browser-attested HTML form navigation with `Sec-Fetch-Site: same-origin`, `Sec-Fetch-Mode: navigate`, and `Sec-Fetch-Dest: document` when textual Origin differs from the container-received Host.
- Browsers generate those Fetch Metadata headers. Web scripts cannot modify them.
- This path **still requires** a valid port-scoped private session cookie, matching hidden CSRF token, safe form encoding and server Host bound to `127.0.0.1` with the designated port.
- `cross-site`, `same-site`, `none`, HTTPS, opaque, invalid-origin and forged/incomplete metadata form requests are blocked.
- Automated regression tests cover legitimate forwarded POSTs and hostile form attempts without private data disclosure.

## Upgrade and verify

Stop previously running OpenCMMC processes. Extract the Linux archive over the application directory without deleting `private-assessment`. Check `./opencmmc version` returns v1.0.0-rc.8. Start `./start-cmmc-ui.sh`, open the freshly printed private loopback URL in the browser and submit a form.

If the error persists, share only the rejected POST's `Origin`, `Sec-Fetch-Site`, `Sec-Fetch-Mode`, and `Sec-Fetch-Dest` values. Do not disclose cookies, tokens, assessment evidence, private URLs or file paths.

**Release CI does not establish success on every physical Chromebook.** Local evidence is not encrypted by the application. Obtain organizational authorization before using CUI.
