# OpenCMMC release operations

Canonical repository: https://github.com/blocksifrdev/OpenCMMC

The source Go module `github.com/blocksifrdev/freecompliance` and `cmd/freecompliance` package path intentionally remain compatible. New release binaries are named `opencmmc`, and include a legacy executable alias. The historical v0.5 preview artifacts are never renamed.

## Current status

The **latest published prerelease** is [`v1.0.0-rc.1`](https://github.com/blocksifrdev/OpenCMMC/releases/tag/v1.0.0-rc.1), published November 11, 2026, following successful native smoke tests across four architectures. Historical `v0.5.0-preview.4` remains available. The `opencmmc-release-candidate.yml` workflow publishes a prerelease only from the dedicated `release/v1.0.0-rc.1` branch after all test jobs pass. Neither candidate publication nor green CI establishes feature-complete `v1.0.0` GA.

## Validate fresh source

```bash
git clone https://github.com/blocksifrdev/OpenCMMC.git
cd OpenCMMC
go test ./...
go test -race ./internal/cmmc ./cmd/freecompliance
go vet ./...
python3 -m unittest discover -s scripts -p 'test_*.py' -v
go build -trimpath -o opencmmc ./cmd/freecompliance
./opencmmc version
python3 scripts/package_release.py --version v1.0.0-rc.1 --out dist
(cd dist && sha256sum -c SHA256SUMS)
```

Native Windows ARM64, Windows x64, Linux ARM64 and Linux x64 smoke tests must pass before the package can be described as verified on those architectures.

## Promotion

Never automatically promote to `v1.0.0`. Complete the independent CMMC, operational, security, upgrade, publishing and legal release gates documented in [the gate register](RELEASE_GATES_2026-11-11.md). Neither a readiness report nor a software release confers CMMC certification, official SPRS submission, C3PAO accreditation or government endorsement.
