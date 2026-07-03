# Dockyard

Dockyard is a package manager, renderer, release tracker, dependency-aware installer, and security linter for Docker Compose applications.

Dockyard renders packages to plain Docker Compose YAML, records local release revisions, and delegates runtime operations to `docker compose`. Docker Compose remains the runtime source of truth.

## What Dockyard does

```text
Dockyard package
  -> values and schema validation
  -> optional env-file support
  -> Compose rendering
  -> policy and secrets checks
  -> lock/package/verify
  -> OCI push/pull through ORAS
  -> dependency planning and explicit dependency installation
  -> release revision state
  -> docker compose up/down
```

## Official package catalog

Dockyard has a companion catalog at `github.com/nandub/dockyard-packages`.

The default catalog registry is:

```text
ghcr.io/nandub/dockyard-packages
```

You can install known catalog packages with short names:

```bash
dockyard catalog list
dockyard catalog info redis
dockyard install redis
```

The shorthand resolves to the configured catalog package:

```text
dockyard install redis
# equivalent source: catalog://redis
# resolved OCI source: oci://ghcr.io/nandub/dockyard-packages/redis:0.1.0
```

For automation, use JSON dry-run output. Dockyard suppresses OCI progress output in this mode so stdout remains parseable JSON:

```bash
dockyard install --dry-run redis --json
```

Use a custom release name with a catalog package name:

```bash
dockyard install my-cache redis
```

Use an explicit catalog or OCI source when you want pinning or a non-default registry:

```bash
dockyard install redis catalog://redis:0.1.0
dockyard install redis oci://ghcr.io/nandub/dockyard-packages/redis:0.1.0
```

Override the default catalog registry for private catalogs:

```bash
export DOCKYARD_CATALOG=ghcr.io/my-org/my-dockyard-packages
```

The stable v1.x package manifest is `Dockyard.yaml` with:

```yaml
apiVersion: dockyard.dev/v1alpha1
```

## Current dependency workflow

Dockyard supports multi-service Compose packages and package-level dependencies.

```bash
dockyard package deps ./examples/team-dashboard
dockyard install-plan team-dashboard ./examples/team-dashboard
dockyard install --dry-run team-dashboard ./examples/team-dashboard
dockyard install --with-dependencies team-dashboard ./examples/team-dashboard
```

Dependency behavior is conservative:

- plain `dockyard install` installs only the root package;
- `--with-dependencies` is explicit opt-in;
- dependencies install before the root package;
- dependency release names are deterministic, such as `team-dashboard-db`;
- `dockyard list` and `dockyard status` show parent/child relationships;
- dependency releases are protected from accidental uninstall while active parents still depend on them;
- dependencies are not automatically removed when the root release is uninstalled.

## Documentation

- [Getting started](docs/getting-started.md) — build, local layout, and smoke tests.
- [Operator guide](docs/operator-guide.md) — values, env files, overlays, release state, pruning, and dependency operations.
- [Packaging and distribution](docs/packaging-and-distribution.md) — package quality checks, lockfiles, archives, dependencies, verification, and OCI.
- [Command reference](docs/command-reference.md) — CLI command groups and flags.
- [Security](docs/security.md) — policy checks, secrets, and hardening.
- [Docker Compose compatibility](docs/compose-compatibility.md) — what Dockyard handles versus what Compose handles.
- [Real-world example](docs/real-world-example.md) — Team Dashboard with PostgreSQL.
- [Release engineering](docs/release-engineering.md) — release artifacts, checksums, and SBOMs.
- [Upgrade policy](docs/upgrade-policy.md) — stable and experimental format expectations.
- [Support policy](docs/support-policy.md) — supported platforms and support boundaries.

## Install from release artifacts

GitHub releases include cross-platform binaries, `SHA256SUMS`, and an SPDX SBOM.

Windows PowerShell:

```powershell
Invoke-WebRequest `
  -Uri https://github.com/nandub/dockyard/releases/latest/download/dockyard-windows-amd64.exe `
  -OutFile dockyard.exe

Invoke-WebRequest `
  -Uri https://github.com/nandub/dockyard/releases/latest/download/SHA256SUMS `
  -OutFile SHA256SUMS

Get-FileHash .\dockyard.exe -Algorithm SHA256
Get-Content .\SHA256SUMS

.\dockyard.exe doctor
.\dockyard.exe version
```

Linux amd64:

```bash
curl -L -o dockyard https://github.com/nandub/dockyard/releases/latest/download/dockyard-linux-amd64
curl -L -o SHA256SUMS https://github.com/nandub/dockyard/releases/latest/download/SHA256SUMS
sha256sum -c SHA256SUMS --ignore-missing
chmod +x dockyard
./dockyard doctor
./dockyard version
```

macOS Apple Silicon:

```bash
curl -L -o dockyard https://github.com/nandub/dockyard/releases/latest/download/dockyard-darwin-arm64
chmod +x dockyard
./dockyard doctor
./dockyard version
```

## Local development

Keep generated packages, deployment values, env files, rendered files, and package archives outside this repository.

```text
dockyard/              # Dockyard source repository
../dockyard-work/      # generated test packages
../deploy-values/      # operator values and env files
../dockyard-artifacts/ # generated package archives and rendered files
```

Common development commands:

```bash
go mod tidy
make verify
make dev-build
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

`make build` is intentionally non-mutating. If `go.mod` or `go.sum` need updates, run `go mod tidy` explicitly first.

## Quick start

PowerShell:

```powershell
New-Item -ItemType Directory -Force ..\dockyard-work, ..\deploy-values, ..\dockyard-artifacts | Out-Null

.\bin\dockyard.exe doctor
.\bin\dockyard.exe init ..\dockyard-work\example-app
.\bin\dockyard.exe values template ..\dockyard-work\example-app -o ..\deploy-values\local.yaml
.\bin\dockyard.exe lint ..\dockyard-work\example-app -f ..\deploy-values\local.yaml
.\bin\dockyard.exe render ..\dockyard-work\example-app -f ..\deploy-values\local.yaml --validate-compose
.\bin\dockyard.exe install example ..\dockyard-work\example-app -f ..\deploy-values\local.yaml
.\bin\dockyard.exe status example --compose-ps
Invoke-WebRequest http://localhost:8080 -UseBasicParsing
.\bin\dockyard.exe uninstall example
```

Linux/macOS:

```bash
mkdir -p ../dockyard-work ../deploy-values ../dockyard-artifacts

./bin/dockyard doctor
./bin/dockyard init ../dockyard-work/example-app
./bin/dockyard values template ../dockyard-work/example-app -o ../deploy-values/local.yaml
./bin/dockyard lint ../dockyard-work/example-app -f ../deploy-values/local.yaml
./bin/dockyard render ../dockyard-work/example-app -f ../deploy-values/local.yaml --validate-compose
./bin/dockyard install example ../dockyard-work/example-app -f ../deploy-values/local.yaml
./bin/dockyard status example --compose-ps
curl http://localhost:8080
./bin/dockyard uninstall example
```

## Package publishing example

```bash
dockyard lock ./examples/nginx
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
dockyard package ./examples/nginx --locked -o ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz
dockyard verify ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz --require-lock
dockyard push ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz oci://ghcr.io/nandub/dockyard/nginx:0.1.0
```

Dockyard publishes packages with:

```text
artifact type: application/vnd.dockyard.package.v1+gzip
archive layer: application/vnd.dockyard.package.archive.v1+gzip
```

OCI push/pull uses the `oras` CLI. Run `dockyard doctor` to check whether `oras` is available.

## Examples

Runnable package examples live under `examples/`:

- `examples/nginx` — basic local smoke-test package.
- `examples/postgres` — reusable PostgreSQL dependency package.
- `examples/postgres-app` — app plus PostgreSQL in one package.
- `examples/team-dashboard` — dependency planning and `--with-dependencies` example.
- `examples/caddy-letsencrypt` — automatic HTTPS with Caddy.
- `examples/nginx-tls-mounted-certs` — TLS with operator-provided certificate/key files.
- `examples/traefik-letsencrypt` — Let's Encrypt with Traefik Docker labels.

The TLS examples show package design patterns. Dockyard does not manage certificates itself; packages model TLS using standard Docker Compose.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for local development, verification, and pull request expectations.
