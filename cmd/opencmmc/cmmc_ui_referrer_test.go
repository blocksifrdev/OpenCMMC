package main

import (
    "net/http"
    "net/url"
    "testing"
)

// Referrer-Policy: no-referrer makes Chromium serialize the Origin of
// ordinary HTML form POSTs as "null". Using same-origin keeps strict
// CSRF/Origin enforcement viable without disclosing referrers to external sites.
func TestUILocalBrowserReferrerPolicySupportsStrictOrigin(t *testing.T) {
    ui := testUI(t)
    initial := uiRequest(t, ui, http.MethodGet, "/", nil, true, "")
    if initial.Code != http.StatusOK {
        t.Fatalf("expected workbench, got %d", initial.Code)
    }
    if got := initial.Header().Get("Referrer-Policy"); got != "same-origin" {
        t.Fatalf("browser-safe referrer policy required, got %q", got)
    }
    posted := uiRequest(t, ui, http.MethodPost, "/create", url.Values{
        "scope": {"Synthetic Chromium browser origin regression"},
        "level": {"1"},
        "csrf": {ui.secret},
    }, true, "http://127.0.0.1:8765")
    if posted.Code != http.StatusSeeOther {
        t.Fatalf("valid same-origin authenticated POST failed: %d", posted.Code)
    }
}
