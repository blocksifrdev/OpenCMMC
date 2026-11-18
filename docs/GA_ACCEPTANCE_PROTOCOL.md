# OpenCMMC v1.0 general-availability acceptance protocol

**Status: NO-GO.** This is a test protocol, not an issued approval, accredited
assessment, ATO, certified CMMC methodology or permission to handle CUI.

## Supported v1 scope

OpenCMMC v1 is a free, local/offline **readiness and evidence-organization
workbench**: Level 1's 15 FAR safeguarding items, Level 2's 110 NIST
SP 800-171 Revision 2 requirement identifiers, user-verified assessment
objectives, scoped local evidence, reviewer observations, clear unresolved
states, local JSON export and validated backup/restore. It does not determine
actual control effectiveness from evidence hashes, certify CMMC compliance,
submit an SPRS score, provide an accredited C3PAO assessment, enforce an
organization's security policy, or substitute for a CUI system authorization.

If authoritative objective completeness or scoring cannot be independently
reconciled, they **must remain separately labeled experimental/draft or
excluded from GA marketing**. Link-only commercial service discovery is not
active policy enforcement or security scanning.

## Evidence-bearing gate matrix

| Gate | Required evidence | Owner | Current result |
| --- | --- | --- | --- |
| Source and runtime | Exact release commit, supported compiler version, Go tests/race/vet, no unresolved known reachable advisories | Release engineering | Integrated CI PASS, GA-tag verification OPEN |
| Real browser and origin | Successful local session bootstrap, Level 1 and Level 2 POST, evidence review and exports, spoofed-origin and CSRF rejection | Engineering and external reviewer | Headless Chromium PASS, actual Chromebook OPEN |
| Native distribution | Signed verified Windows x64/ARM64 and Linux x64/ARM64 archives; first-run, reports, backup, verify, restore and overwrite rejection | Engineering | Integrated CI PASS, signed GA asset OPEN |
| Supply chain | Public tag, exact archive digests, signed SLSA provenance, signed SBOM, independent trust anchor, offline verification | Security/release owner | PR-attested CI PASS, public GA asset OPEN |
| Assessment methodology | Normative Rev 2 baseline, objective IDs and methods, omissions, scope, exceptions and reviewed draft scoring; written independent reconciliation | Independent qualified CMMC assessor | OPEN |
| Independent security | Authorized threat model, external penetration test against loopback/browser/filesystem/update chain, signed report, zero unresolved critical/high, retest | Independent security tester | OPEN |
| Physical air gap and recovery | Real isolated Windows/Linux, offline startup, adverse evidence and restore scenarios; exact source/build SHA | Authorized pilot operator | OPEN |
| Chromebook hardware | Successful host-Chrome or supported native Linux browser Level 1/2 forms and export on the actual device; no secret URLs or customer records disclosed | Device operator | OPEN (#49) |
| Accessibility | Keyboard-only and assisted-technology review, relevant WCAG/Section 508 findings triaged | Independent accessibility reviewer | OPEN |
| Software support | Private vulnerability reporting, support/release ownership, rollback and recovery instructions, branch protection and approved signing identity | Release owner | PARTIAL |

## Current internal, reproducible evidence

- Integrated source merge:
  `a11468862ebfa26508b9c58fbca80e786494fc30`
- Go and real Chromium CI:
  https://github.com/blocksifrdev/OpenCMMC/actions/runs/44393349219
- CodeQL Go, Python, JavaScript:
  https://github.com/blocksifrdev/OpenCMMC/actions/runs/44393349466
- Patched Go 1.27.4, `govulncheck`, both GitHub/Sigstore signatures,
  ZIP/TAR checksums, four OS/architecture smoke runs, browser acceptance:
  https://github.com/blocksifrdev/OpenCMMC/actions/runs/44393349748

These runs covered the PR integration source revision. The later squash-merge
commit and any eventual GA tag need their own exact artifact provenance.
A CI workflow's green status does not imply absence of published alerts.
No release candidate should be silently relabeled as v1.0.0.

## Chromebook acceptance, synthetic data only

Download the precise approved build and independently verify its signature and
SHA256. Do not delete or reuse an existing customer assessment directory.
From Chromebook Linux, run:

```sh
./opencmmc version
./opencmmc cmmc ui --data-dir "$HOME/opencmmc-ga-synthetic-L1" --port 8876
```

Open **only locally** the private URL printed by that process, then:
1. Create Level 1 synthetic assessment; mark one requirement GAP; attach a
   nonsensitive test text file; export and open local JSON.
2. Close the server and start another with a fresh `...-L2` directory on
   port `8877`; create Level 2; export JSON; confirm 110 baseline requirement
   identifiers and unresolved objectives rather than fabricated completeness.
3. In the same installation, exercise `cmmc backup`, `verify-backup` and
   `restore-backup` against an approved encrypted test location.
4. Report **version, success/failure, safe error class, tested browser type
   and date only**. Never provide the URL, access code, cookie, CSRF token,
   evidence, HAR data or CUI.

## Independent review signoff record

Each reviewer supplies an identifiable dated, signed report with: explicit
scope, exact source and archive SHA-256, independent role/organization, test
methods and denominator, discovered findings and severity, resolved and
unresolved exclusions, remediation/retest evidence, limitations, and an
explicit recommendation. Keep private or sensitive report contents in
approved storage, with a public SHA256 and a redacted disposition if needed.

**GA decision:** owner may authorize `v1.0.0` only after the advertised
functions and all applicable critical gates have documented supporting
evidence. No automatic script, self-attestation, or AI evaluation substitutes
for the human external signoff.
