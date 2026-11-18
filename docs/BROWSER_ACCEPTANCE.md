# Real-browser acceptance gate

OpenCMMC's core Go runtime remains free, open source and offline-first.
Chromium and Playwright are **test-only CI dependencies**, not packaged with
the application, not downloaded by customers, and not given any assessment data.

## Browser acceptance

The `tests/browser_e2e.mjs` script builds a fresh **synthetic** workspace and
uses an actual headless Chromium browser to validate:

1. Local private-token bootstrap followed by removal of the token from the
   address bar.
2. Rejection of a cross-site form before a workspace exists.
3. Level 1 workspace creation, reviewer gap decision, local synthetic evidence
   upload, and downloaded JSON readiness report.
4. A second browser session via an IPv6 loopback reverse proxy which preserves
   the browser's `[::1]` Origin but forwards to the application's IPv4
   loopback listener with a rewritten Host. Saving another review must succeed
   without disabling the session cookie or CSRF check.
5. Level 2 workspace creation and JSON export with no imported objectives.

Both `ci.yml` and the native-archive candidate workflow run this test.
Public release candidate publication **must depend** on passing native Linux
and Windows package smoke checks **and** the browser test.

The Chromium test covers one defined forwarded-origin condition. It is not a
substitute for browser validation on the customer's actual Chromebook,
independent penetration testing, official CMMC method review, storage
encryption, authorized CUI handling, or a government ATO.

## Local developer invocation

With Go and Node.js installed on a development host:

```sh
go build -trimpath -o opencmmc-e2e ./cmd/opencmmc
npm install --package-lock=false --ignore-scripts --no-audit --no-fund
npx playwright install --with-deps chromium
OPENCMMC_BIN="$PWD/opencmmc-e2e" npm run test:browser
```

This CI-only installation requires internet access; the released OpenCMMC
binary and local assessment workflow do not. Never test with customer CUI,
production credentials, or a real private workbench URL.


## Confirmed Chromium form-origin root cause

The initial real Chromium run exposed `ORIGIN_MISMATCH` with a fixed-category
diagnostic of `origin=opaque; fetch-site=same-origin; fetch-mode=navigate;
fetch-dest=document`. The server had set `Referrer-Policy: no-referrer`.
Under this policy, browsers may use `Origin: null` for ordinary HTML form
POSTs. The correction is `Referrer-Policy: same-origin`, not relaxing origin
checks, accepting opaque Origins, turning off CSRF, or publishing more
ChromeOS-specific exceptions. Same-origin discloses no Referer to external
sites; optional external service links also have `rel=noreferrer`.

References:
- https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Referrer-Policy
- https://issuetracker.google.com/issues/336222284
- https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html

A successful CI browser run verifies this behavior on GitHub-hosted Chromium,
not on every end user's browser or the user's physical Chromebook.
