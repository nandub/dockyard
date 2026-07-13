# Release Workflow

This document records Dockyard release, package, publish, and release-verification workflows.

## Authoritative Sources

Official release automation:

- `.github/workflows/release.yml`

Human release process:

- `docs/release-engineering.md`
- `docs/release-candidate-checklist.md`

Supporting package and publish process:

- `docs/packaging-and-distribution.md`
- `README.md`
- `docs/command-reference.md`

## Release Prerequisites

Before committing or tagging:

```bash
make verify
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

Release checklist local verification:

```bash
go mod tidy
make verify
make dev-build
dockyard version
dockyard compat
```

Example package verification:

```bash
dockyard lock ./examples/nginx
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

When Docker is available:

```bash
dockyard package test ./examples/nginx --smoke
```

The checklist says to repeat the non-smoke gate for each public example package.

## Local Release Snapshot

Command:

```bash
make release-snapshot VERSION=v1.0.0
```

Outputs:

- `bin/dockyard-windows-amd64.exe`
- `bin/dockyard-linux-amd64`
- `bin/dockyard-linux-arm64`
- `bin/dockyard-darwin-amd64`
- `bin/dockyard-darwin-arm64`

The Makefile has separate Windows and non-Windows implementations for `release-snapshot`.

## Official GitHub Release

Trigger:

```bash
git tag v1.0.0
git push origin v1.0.0
```

Workflow:

- `.github/workflows/release.yml`

Trigger pattern:

```yaml
tags:
  - "v*"
```

Build matrix:

- Windows AMD64
- Linux AMD64
- Linux ARM64
- Darwin AMD64
- Darwin ARM64

Build flags:

```text
-trimpath
-ldflags "-s -w
  -X github.com/nandub/dockyard/internal/version.Version=${VERSION}
  -X github.com/nandub/dockyard/internal/version.Commit=${COMMIT}
  -X github.com/nandub/dockyard/internal/version.Date=${DATE}"
```

Version metadata source:

- `VERSION`: Git tag name.
- `COMMIT`: GitHub SHA.
- `DATE`: GitHub event head commit timestamp.

## Release Verification in GitHub Actions

Every matrix build runs:

```bash
make verify
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
go build ...
```

Linux AMD64 additionally runs:

```bash
./dist/dockyard-linux-amd64 version
./dist/dockyard-linux-amd64 compat ./examples/nginx --strict
./dist/dockyard-linux-amd64 package test ./examples/nginx --strict
```

The release workflow does not run Docker smoke tests. `docs/release-engineering.md` says smoke tests remain a local/manual gate because they require a reachable Docker daemon.

## Release Artifacts

GitHub release assets:

- `dockyard-windows-amd64.exe`
- `dockyard-linux-amd64`
- `dockyard-linux-arm64`
- `dockyard-darwin-amd64`
- `dockyard-darwin-arm64`
- `SHA256SUMS`
- `dockyard-source.spdx.json`

SBOM generation:

```bash
syft dir:. -o spdx-json=dist/dockyard-source.spdx.json
```

Checksum generation:

```bash
cd dist
sha256sum \
  dockyard-windows-amd64.exe \
  dockyard-linux-amd64 \
  dockyard-linux-arm64 \
  dockyard-darwin-amd64 \
  dockyard-darwin-arm64 \
  dockyard-source.spdx.json > SHA256SUMS
```

Release upload:

- `softprops/action-gh-release@v2`

## Package Publishing

Example package gate:

```bash
dockyard lock ./examples/nginx
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
dockyard package ./examples/nginx --locked -o ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz
dockyard verify ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz --require-lock
```

Publish to OCI:

```bash
oras login ghcr.io
dockyard push ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz oci://ghcr.io/nandub/dockyard/nginx:0.1.0
```

Test published package:

```bash
dockyard package test oci://ghcr.io/nandub/dockyard/nginx:0.1.0 --smoke
```

## Catalog Publishing

Catalog index publish command documented in `docs/packaging-and-distribution.md`:

```bash
oras push --artifact-type application/vnd.dockyard.catalog.v1+yaml ghcr.io/nandub/dockyard-packages/catalog:latest catalog.yaml:application/vnd.dockyard.catalog.index.v1+yaml
```

Catalog consumption:

```bash
DOCKYARD_CATALOG=oci://ghcr.io/nandub/dockyard-packages/catalog:latest dockyard catalog list
```

## Documentation Review Before Tagging

The release checklist names these files for review:

- `README.md`
- `docs/getting-started.md`
- `docs/operator-guide.md`
- `docs/packaging-and-distribution.md`
- `docs/security.md`
- `docs/compose-compatibility.md`
- `docs/v1-readiness.md`
- `docs/upgrade-policy.md`
- `docs/support-policy.md`
- `docs/release-candidate-checklist.md`
- `AGENTS.md`
- `CHANGELOG.md`

## Format Stability

For `v1.0.0`, the release checklist classifies:

Stable:

- `Dockyard.yaml` API version `dockyard.dev/v1alpha1`

Experimental:

- `dockyard.lock`
- `package.provenance.json`
- `release.json`

## Contradictions and Drift

- `docs/release-engineering.md` describes release workflow verification as `make verify`, Staticcheck, `dockyard version`, `dockyard compat ./examples/nginx --strict`, and `dockyard package test ./examples/nginx --strict`. In `.github/workflows/release.yml`, the Dockyard binary checks only run for Linux AMD64.
- Local release snapshots write to `bin/`; official release workflow writes to `dist/`.
- Local `make release-snapshot` does not create `SHA256SUMS` or SBOM; the official workflow does.
- `docs/release-candidate-checklist.md` says to repeat non-smoke gates for each public example package; the release workflow only checks `examples/nginx`.
- Package publishing examples target `oci://ghcr.io/nandub/dockyard/...`, while catalog examples target `ghcr.io/nandub/dockyard-packages/catalog:latest`.

## Unverified

This document records configured and documented release workflow. It does not verify:

- Current GitHub release status.
- Current tag state.
- Whether package publishing credentials exist.
- Whether ORAS is installed locally.
- Whether Docker is available for smoke tests.
