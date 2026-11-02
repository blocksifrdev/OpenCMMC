# Threat model and evidence limitations

This is not a security or compliance certification. Threats include untrusted repository content, symlink escapes, oversized files, inaccurate configuration parsers, arbitrary network URLs, token exposure, report leaks, stale API metadata, API rate limits, and falsely inferred control effectiveness.

## Local scanner

- Operates without network access or credentials. Never executes scanned code.
- Skips `.git`, vendor files and common generated directories.
- Does not follow file symlinks. Only hashes regular readable artifacts with a 1 MiB size cap.
- Does not render raw evidence content into HTML. User-controlled paths render through Go's `html/template` escaping.
- Creates output files with restrictive permissions (0600 on POSIX). Output paths should stay outside scanned source directories.
- Records SHA-256 digests, not digital signatures or trusted timestamps. A digest cannot prove that policy controls ran.
- Uses conservative pattern checks for known GitHub Actions risks. The checks are not a complete YAML parser and can miss alternate syntax or runtime behavior. Confirm consequential results independently.

## GitHub assessor

- Only uses GitHub REST `GET` requests on fixed official API endpoints. `--repo` only accepts `owner/name`.
- Optionally reads a token from `GITHUB_TOKEN`, never from command-line parameters.
- Does not emit token values or GitHub response bodies into reports. Error messages exclude remote response bodies.
- Public GitHub API calls require no token. Private repositories require a least-privilege read-only token. Avoid admin-scoped and classic tokens where possible.
- Branch `protected=true` is **not** equivalent to required reviews, enforced checks, secure bypass settings, or passed security controls.
- GitHub repository metadata and default branch can change between requests. Neither request is a transactional snapshot.
- An unreachable or denied repository metadata request fails the assessment. A failed branch query produces `ATTESTATION_NEEDED`, never `VERIFIED`.

## Explicit non-goals of v0.2.0

No source-code upload, multi-cloud IAM assessment, SOC 2/CMMC/ISO 27001 certification, actual policy execution, remote rule modifications, signed evidence receipts, or auditor opinion. Never claim those capabilities from scanner output.
