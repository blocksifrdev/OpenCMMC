# FreeCompliance v0.2.0

Free and open-source repository evidence readiness, with an optional read-only GitHub API assessor. No subscription, no remote account required for local scans, no source upload, and no artificial compliance score.

## Included

- Single-binary Go CLI with no third-party module dependencies.
- 14 local repository evidence assertions with explicit `VERIFIED`, `PARTIAL`, `GAP`, `MISSING` and `ATTESTATION_NEEDED` statuses.
- Local GitHub Actions supply-chain lint checks (pinning and explicit write-all), file evidence hashes, HTML and JSON reporting.
- Read-only GitHub repository assessor with 5 additional metadata assertions: API visibility, active status, default branch protected flag, required reviews unknown, required checks unknown.
- Drift comparison, unit tests, race tests, CI workflow, composite GitHub Action and security documentation.

## Limitations

This does not certify compliance with any framework. Local artifact existence is not control effectiveness. GitHub `protected=true` does not prove any particular approval or check is enforced. The GitHub API assessor requires network access and optionally uses a narrowly scoped read-only token.
