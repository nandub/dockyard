# Dockyard Dependency Graph

This document records dependency relationships discovered by static inspection.

## Module Dependencies

Go module:

- `github.com/nandub/dockyard`

Go version:

- `1.23`

Direct dependencies from `go.mod`:

- `github.com/spf13/cobra v1.10.2`
- `github.com/santhosh-tekuri/jsonschema/v6 v6.0.2`
- `go.yaml.in/yaml/v4 v4.0.0-rc.5`

Indirect dependencies from `go.mod`:

- `github.com/inconshreveable/mousetrap v1.1.0`
- `github.com/spf13/pflag v1.0.9`
- `golang.org/x/text v0.14.0`

Evidence:

- `go.mod`
- `go.sum`

## Build and Tool Dependencies

Local build and verification commands are defined in `Makefile`.

Make targets:

- `fmt`: `go fmt ./...`
- `fmt-check`: `go run ./tools/fmtcheck ./cmd ./internal ./tools`
- `tidy`: `go mod tidy`
- `tidy-check`: `go mod tidy -diff`
- `test`: `go test ./...`
- `build`: `go build ... ./cmd/dockyard`
- `dev-build`: `tidy fmt build`
- `verify`: `tidy-check fmt-check test`
- `release-snapshot`: cross-compiles release binaries into `bin/`
- `clean`: removes `bin/`

External developer tools referenced by docs and workflows:

- Staticcheck: `go run honnef.co/go/tools/cmd/staticcheck@latest ./...`
- govulncheck: `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`
- Semgrep: `semgrep scan --config auto`
- Syft: `syft dir:. -o spdx-json=dist/dockyard-source.spdx.json`

## Runtime Dependencies

Docker runtime path:

```text
dockyard CLI
  -> internal/runner
    -> docker CLI
      -> docker compose
      -> Docker daemon
```

Evidence:

- `internal/runner/docker.go`
- `internal/cli/doctor.go`

OCI runtime path:

```text
dockyard CLI
  -> internal/oci
    -> oras CLI
      -> OCI registry
```

Evidence:

- `internal/oci/oci.go`
- `internal/cli/push.go`
- `internal/cli/pull.go`

Catalog runtime path:

```text
dockyard CLI
  -> internal/catalog
    -> DOCKYARD_CATALOG or default catalog reference
    -> file path, file:// path, or oci:// reference
    -> oras CLI for OCI catalog pulls
    -> ~/.dockyard/cache/catalogs
```

Evidence:

- `internal/catalog/catalog.go`

Default catalog:

- `oci://ghcr.io/nandub/dockyard-packages/catalog:latest`

## Internal Package Dependency Shape

High-level dependency direction:

```text
cmd/dockyard
  -> internal/cli

internal/cli
  -> internal/archive
  -> internal/catalog
  -> internal/dockpkg
  -> internal/envfile
  -> internal/format
  -> internal/lock
  -> internal/oci
  -> internal/policy
  -> internal/quality
  -> internal/render
  -> internal/runner
  -> internal/state
  -> internal/values
  -> internal/version

internal/archive
  -> internal/dockpkg
  -> internal/format
  -> internal/lock

internal/dockpkg
  -> internal/format

internal/lock
  -> internal/dockpkg
  -> internal/format
  -> internal/render

internal/render
  -> internal/dockpkg

internal/state
  -> internal/format

internal/oci
  -> external oras CLI

internal/runner
  -> external docker CLI

internal/catalog
  -> external oras CLI for OCI references
```

This graph is inferred from import and source inspection, not from `go list`, which was not successfully verified during discovery.

## Package Source Dependency Graph

Dockyard package examples:

```text
examples/team-dashboard
  -> dependency: postgres
  -> alias: db
  -> source: oci://ghcr.io/nandub/dockyard/postgres:0.1.0
```

Other examples have no dependency metadata found in `Dockyard.yaml`:

- `examples/caddy-letsencrypt`
- `examples/nginx`
- `examples/nginx-tls-mounted-certs`
- `examples/postgres`
- `examples/postgres-app`
- `examples/traefik-letsencrypt`

Evidence:

- `examples/*/Dockyard.yaml`

## Container Image Dependencies

From repository Compose examples and values:

- `nginx:1.27`
- `postgres:16.4`
- `caddy:2.8`
- `traefik:3.1`
- `traefik/whoami:v1.10`

Parameterized Compose references:

- `${image.repository}:${image.tag}`
- `${app.image}:${app.tag}`
- `${database.image}:${database.tag}`
- `${caddy.image}:${caddy.tag}`
- `${traefik.image}:${traefik.tag}`
- `${whoami.image}:${whoami.tag}`

Evidence:

- `examples/*/compose.yaml`
- `examples/*/values.yaml`

CI container image:

- `semgrep/semgrep:latest`

Evidence:

- `.github/workflows/security.yml`

## Artifact Dependencies

Archive generation adds:

- `package.provenance.json`
- `SHA256SUMS`

Optional package lock:

- `dockyard.lock`

Release workflow generates:

- `dockyard-windows-amd64.exe`
- `dockyard-linux-amd64`
- `dockyard-linux-arm64`
- `dockyard-darwin-amd64`
- `dockyard-darwin-arm64`
- `dockyard-source.spdx.json`
- `SHA256SUMS`

Evidence:

- `internal/archive/archive.go`
- `Makefile`
- `.github/workflows/release.yml`
- `docs/release-engineering.md`

## Absence Findings

No evidence found for:

- Vendored Go code in `vendor/`.
- Node dependencies in `node_modules/`.
- Dockerfiles.
- Terraform files.
- Shell, PowerShell, Python, JavaScript, or TypeScript scripts.

Evidence:

- `rg --files` with script, Dockerfile, Kustomize, and Terraform globs.
- Directory scan for `vendor`, `node_modules`, `.cache`, `.dockyard`, `dist`, `coverage`, `tmp`, and `temp`.
