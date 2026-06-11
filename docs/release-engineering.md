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

## Tagging

```bash
git tag v0.10.0
git push origin v0.10.0
```


## Release candidate builds

To build a local release-candidate snapshot:

```bash
make release-snapshot VERSION=v1.0.0-rc.1
```

To publish via GitHub Actions, create and push a tag:

```bash
git tag v1.0.0-rc.1
git push origin v1.0.0-rc.1
```

The GitHub workflow uses the tag as the Dockyard version metadata.
