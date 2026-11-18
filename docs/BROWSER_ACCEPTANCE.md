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
