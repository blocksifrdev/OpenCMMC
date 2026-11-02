# FreeCompliance publication and release operations

Canonical public repository: https://github.com/blocksifrdev/freecompliance

## First publication

The initial source code is intentionally published under the maintainer account. It is a separate open-source project and does not include private BlockSiFr Control code. Release tags, binaries and GitHub Releases are distinct from repository publication and should never be implied to exist before verification.

## Verify from a fresh checkout

```bash
git clone https://github.com/blocksifrdev/freecompliance.git
cd freecompliance
go test ./...
go test -race ./...
go vet ./...
go build -trimpath -o freecompliance ./cmd/freecompliance
./freecompliance version
./freecompliance scan --path . --json readiness.json --html readiness.html --fail-on-gap
sha256sum freecompliance
```

Verify the public CI run after the initial commit. Do not merge or tag a subsequent release until CI has passed.

## Tag and publish

When ready, create immutable or signed tag `v0.2.0` pointing to the reviewed source commit. Publish a GitHub Release with `docs/RELEASE_NOTES_v0.2.0.md`, verified binary checksums and provenance if distributing binaries. Prefer pinned commit SHA in any GitHub Action consumers. A checksum by itself is not publisher authentication.

## Important scope

This version provides repository-local evidence discovery and optional read-only GitHub metadata assessment. It does not provide OSCAL exports, cloud IAM assessment, regulatory certification or auditor opinions. The repository contains no customer reports or private production evidence.
