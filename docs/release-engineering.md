# Release engineering

Dockyard release builds are produced by `.github/workflows/release.yml` when a `v*` tag is pushed.

## Local build

For local development:

```bash
make dev-build
make test
dockyard version
```

For a non-mutating build:

```bash
go mod tidy
make build
dockyard version
```

Before committing or tagging:

```bash
make verify
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

On Windows, `make build` writes `bin/dockyard.exe`. On Linux and macOS, it writes `bin/dockyard`. The `build` target does not edit `go.mod` or `go.sum`.

## Version metadata

The release workflow injects version metadata through Go linker flags:

```bash
-X github.com/nandub/dockyard/internal/version.Version=v0.10.0
-X github.com/nandub/dockyard/internal/version.Commit=<commit>
-X github.com/nandub/dockyard/internal/version.Date=<timestamp>
```

The metadata is visible through:

```bash
dockyard version
dockyard version --json
```

Release revisions also record the Dockyard CLI version that created them in `release.json`.

## Release artifacts

The workflow builds:

```text
dockyard-windows-amd64.exe
dockyard-linux-amd64
dockyard-linux-arm64
dockyard-darwin-amd64
dockyard-darwin-arm64
SHA256SUMS
dockyard-source.spdx.json
```

Verify a downloaded binary with `SHA256SUMS` before installing it.

Windows:

```powershell
Get-FileHash .\dockyard-windows-amd64.exe -Algorithm SHA256
Get-Content .\SHA256SUMS
```

Linux/macOS:

```bash
sha256sum -c SHA256SUMS
```

## Release workflow verification

The release workflow runs:

```text
make verify
staticcheck
dockyard version
dockyard compat ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

The smoke test remains a local/manual gate because it requires a reachable Docker daemon.

## Tagging

```bash
git tag v1.0.0
git push origin v1.0.0
```


## Release candidate and final builds

To build a local release snapshot:

```bash
make release-snapshot VERSION=v1.0.0
```

To publish via GitHub Actions, create and push a tag:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The GitHub workflow uses the tag as the Dockyard version metadata.
