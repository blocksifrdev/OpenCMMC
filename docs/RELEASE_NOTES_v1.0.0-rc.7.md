# OpenCMMC v1.0.0-rc.7 | ChromeOS Local Origin Compatibility

**Release class:** Free, public release candidate. Not official CMMC assessment, certification, or validated SPRS score.

## Targeted ORIGIN_MISMATCH repair

OpenCMMC rc.6 correctly reported the remaining browser issue as ORIGIN_MISMATCH. ChromeOS automatically forwards local Linux server connections into host Chrome and documents `localhost` and `penguin.linux.test` browser addresses. Requests that arrive at OpenCMMC's 127.0.0.1 listener can still carry one of those browser-visible origins.

- Accept only exact local **HTTP** Origin hostnames on the listener's port: `127.0.0.1`, `localhost`, `penguin.linux.test`, and IPv6 loopback `::1`.
- Keep the incoming Host restricted to `127.0.0.1` on that same port; do not accept wildcard domains, remote hosts, HTTPS origins, wrong ports, origin paths, userinfo or opaque origins.
- Require the existing port-specific local session cookie and an exact CSRF form token for every state-changing request.
- Regression tests exercise all three browser forwarding cases against the real POST handler, reject hostile and malformed origins, and verify free local assessment exports still work.
- Retain visible build version and distinct diagnostics for rejected browser requests.

## Chromebook update

Stop the old local OpenCMMC process. Verify the downloaded rc.7 Linux archive and extract it into the existing `~/OpenCMMC` directory **without deleting `private-assessment`**. Verify `./opencmmc version` is rc.7. Start `./start-cmmc-ui.sh` and open the complete new private local URL printed by the terminal.

Google's ChromeOS documentation: https://developers.google.com/chromeos/app-development/develop/web-environment.

**The fix is tested in CI, not confirmed on this customer's Chromebook until a fresh browser form succeeds.** Local files are not encrypted by the application; obtain proper authorization before storing any CUI.
