# Build Workflow

This document records Dockyard build, format, restore, lint, test, package, and publish commands.

## Command Authority

Local build authority:

- `Makefile`

Hosted build authority:

- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`

Documentation sources:

- `README.md`
- `CONTRIBUTING.md`
- `docs/getting-started.md`
- `docs/packaging-and-distribution.md`
- `docs/release-engineering.md`
- `docs/release-candidate-checklist.md`

## Bootstrap

Documented first local setup:

```bash
go mod tidy
make dev-build
```

Alternative first-time local development path:

```bash
make dev-build
make test
```

`make dev-build` is a convenience target. It may rewrite module and source formatting state because it runs `tidy` and `fmt`.

## Restore

Mutation:

```bash
go mod tidy
make tidy
```

Check:

```bash
make tidy-check
```

Underlying commands:

```bash
go mod tidy
go mod tidy -diff
```

## Build

Non-mutating local build:

```bash
make build
```

Underlying command:

```bash
go build -ldflags "$(LDFLAGS)" -o bin/dockyard$(go env GOEXE) ./cmd/dockyard
```

The Makefile injects:

```text
github.com/nandub/dockyard/internal/version.Version
github.com/nandub/dockyard/internal/version.Commit
github.com/nandub/dockyard/internal/version.Date
```

Default values:

- `VERSION ?= dev`
- `COMMIT` from `git rev-parse --short HEAD`, or `unknown`
- `DATE` from `git show -s --format=%cI HEAD`, or `unknown`

Manual documented builds:

```powershell
go mod tidy
go test ./...
go build -o bin/dockyard.exe ./cmd/dockyard
```

```bash
go mod tidy
go test ./...
go build -o bin/dockyard ./cmd/dockyard
```

CI build:

```bash
go build -o bin/dockyard ./cmd/dockyard
```

Release build:

```bash
go build -trimpath -ldflags "...version metadata..." -o "dist/${name}" ./cmd/dockyard
```

## Format

Mutation:

```bash
make fmt
```

Underlying command:

```bash
go fmt ./...
```

Check:

```bash
make fmt-check
```

Underlying command:

```bash
go run ./tools/fmtcheck ./cmd ./internal ./tools
```

CI mutation-and-diff pattern:

```bash
gofmt -w ./cmd ./internal
git diff --exit-code
```

## Lint

Staticcheck:

```bash
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

Security lint/checks:

```bash
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
semgrep scan --config auto
```

Package lint:

```bash
dockyard package lint ./examples/nginx --strict
```

Compatibility lint:

```bash
dockyard compat ./examples/nginx --strict
```

## Test

Unit tests:

```bash
make test
```

Underlying command:

```bash
go test ./...
```

Verification:

```bash
make verify
```

Underlying sequence:

```text
make tidy-check
make fmt-check
make test
```

Package test:

```bash
dockyard package test ./examples/nginx --strict
```

Smoke test when Docker is available:

```bash
dockyard package test ./examples/nginx --smoke
```

## Package

Create a locked package archive:

```bash
dockyard lock ./examples/nginx
dockyard package ./examples/nginx --locked -o ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz
```

Verify package archive:

```bash
dockyard verify ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz --require-lock
```

Full documented package gate:

```bash
dockyard lock ./examples/nginx
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
dockyard package ./examples/nginx --locked -o ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz
dockyard verify ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz --require-lock
```

## Publish

Package publish:

```bash
oras login ghcr.io
dockyard push ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz oci://ghcr.io/nandub/dockyard/nginx:0.1.0
```

Catalog publish:

```bash
oras push --artifact-type application/vnd.dockyard.catalog.v1+yaml ghcr.io/nandub/dockyard-packages/catalog:latest catalog.yaml:application/vnd.dockyard.catalog.index.v1+yaml
```

Official release publish:

- Push a `v*` tag.
- `.github/workflows/release.yml` builds binaries, generates an SBOM, generates checksums, and uploads release assets.

## Outputs

Local outputs:

- `bin/dockyard.exe`
- `bin/dockyard`
- `bin/dockyard-windows-amd64.exe`
- `bin/dockyard-linux-amd64`
- `bin/dockyard-linux-arm64`
- `bin/dockyard-darwin-amd64`
- `bin/dockyard-darwin-arm64`

Release workflow outputs:

- `dist/dockyard-windows-amd64.exe`
- `dist/dockyard-linux-amd64`
- `dist/dockyard-linux-arm64`
- `dist/dockyard-darwin-amd64`
- `dist/dockyard-darwin-arm64`
- `dist/dockyard-source.spdx.json`
- `dist/SHA256SUMS`

Package outputs:

- `*.dockyard.tgz`
- `SHA256SUMS` inside package archives
- `package.provenance.json` inside package archives
- optional `dockyard.lock`

## Contradictions and Notes

- `make build` injects version metadata. CI's normal build command does not.
- `make build` is documented as non-mutating. `make dev-build` is explicitly mutating because it runs tidy and formatting.
- CI uses mutating formatting and tidy commands followed by `git diff --exit-code`; local verification uses check targets.
- `make fmt` covers `./...`; `make fmt-check` covers `./cmd ./internal ./tools`; CI `gofmt` covers only `./cmd ./internal`.
