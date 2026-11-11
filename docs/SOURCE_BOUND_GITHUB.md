# v0.4 experimental source-bound theorem evidence

OpenCMMC's `prove-github` command performs a **narrow model-relative assessment** of the GitHub *active ruleset declarations* that GitHub returns for a repository's default branch.

## Contract and provenance

1. Read repository identity and default branch with GET `/repos/{owner}/{repo}`.
2. Read active rules with GET `/repos/{owner}/{repo}/rules/branches/{branch}?per_page=100&page=N`.
3. Read no more than 5 pages and refuse proof when pagination completeness cannot be established.
4. Before admitting any evidence, require a matching repository identity and fresh, parseable HTTP Date/Age metadata for every source response. The maximum source age is 5 minutes. Reject redirects, non-200 statuses, malformed JSON, stale/absent dates, and incomplete pagination.
5. Store SHA-256 of exact raw response bytes, source endpoint, retrieval time, source time and source validity. Hashes are **not signatures**, and the public API response may change on the next query.
6. Read-only proofs are evaluated in an explicit eight-state model: review rule present/unproven, status check rule present/unproven, bypass possible/none. No observation asserts `bypass=none`.

### Interpretation

`MODEL_SUPPORTED` for `RULESET_REVIEW_AND_CHECK_DECLARATIONS` means the active ruleset API reported a `pull_request` rule with `required_approving_review_count > 0`, and a `required_status_checks` rule with at least one nonempty context. It does not establish whether these rules apply to every possible real deployment or whether they have been successfully enforced.

`CHANGE_CONTROL_WITH_NO_BYPASS` never receives a supported result from this source alone. Bypass actors can be omitted from ruleset details for read-only callers. Classic branch protection is a separate policy surface, not covered by the active ruleset endpoint. A modeled counterexample is a witness to uncertainty in the finite model, **not an observed bypass or failure**.

If required GitHub metadata cannot be obtained completely and freshly, the output is `UNKNOWN` with zero source-admitted observations and specific warnings. The CLI returns an error if repository identity metadata is unavailable or mismatched.

`--fail-on-unproven` returns a nonzero exit code when any claim is `UNKNOWN` or `MODEL_COUNTEREXAMPLE`, but still writes JSON and HTML reports first.

### Security and limitations

- Use HTTPS to `api.github.com` for production; BaseURL is injectable only in Go test code.
- Credential is read from `GITHUB_TOKEN`; token never appears in report or errors; redirects are disabled.
- Only GET methods are used. No writes, code execution, repository changes or attestation claims.
- HTTP Date/Age is API-supplied freshness, not independently signed provenance. Time of check is not time of enforcement.
- Treat evidence as sensitive and review reports before redistribution.
- The open-source proof is not an OSCAL-validated assessment results artifact; OSCAL interoperable mapping is future work and must be schema-tested before any compliance claim.

## Reproduce

```bash
go test ./...
go test -race ./...
go vet ./...
go run ./cmd/opencmmc prove-github --repo blocksifrdev/opencmmc --json github-proof.json --html github-proof.html
```

See adversarial mock tests in `internal/sourceproof/github_test.go` for permission denial, stale data, malformed responses, complete-empty rules, pagination uncertainty, prohibited redirects and `bypass=possible` preservation.