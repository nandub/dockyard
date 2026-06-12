# Dockyard

Dockyard is a package manager, renderer, release tracker, and security linter for Docker Compose applications.

Dockyard renders packages to plain Docker Compose YAML, records local release revisions, and delegates runtime operations to `docker compose`.

## What Dockyard does

```text
Dockyard package
  -> values validation
  -> optional env-file support
  -> Compose rendering
  -> policy checks
  -> lock/package/verify
  -> release revision state
  -> docker compose up/down
```

Docker Compose remains the runtime source of truth.

Dockyard v1.0 supports `Dockyard.yaml` manifests using `apiVersion: dockyard.dev/v1alpha1` as the stable package manifest contract for the v1.x line.

Dockyard v1.2 adds package dependency metadata. A package can declare dependencies in `Dockyard.yaml`, inspect them with `dockyard package deps`, and include them in `dockyard.lock`. Dockyard v1.3 adds `dockyard install-plan` and dependency-aware `install --dry-run`. Dockyard v1.4 adds explicit opt-in dependency installation with `dockyard install --with-dependencies`.

## Documentation

The docs are intentionally consolidated into a small set:

- [Getting started](docs/getting-started.md) — build, local layout, Windows/Docker Desktop smoke test.
- [Operator guide](docs/operator-guide.md) — values, env files, overlays, release state, and pruning.
- [Packaging and distribution](docs/packaging-and-distribution.md) — package quality checks, lockfiles, archives, verification, and OCI.
- [Security](docs/security.md) — policy checks, secrets, and advanced hardening.
- [Docker Compose compatibility](docs/compose-compatibility.md) — what Dockyard supports directly versus what Compose handles.
- [Real-world example](docs/real-world-example.md) — Team Dashboard with PostgreSQL.
- [v1.0 readiness](docs/v1-readiness.md) — format stability and compatibility checks.
- [Upgrade policy](docs/upgrade-policy.md) — stable and experimental format expectations.
- [Support policy](docs/support-policy.md) — supported platforms and support boundaries.
- [Release checklist](docs/release-candidate-checklist.md) — pre-tag validation steps for release candidates and final releases.

## Install and verify releases

GitHub releases include cross-platform binaries and `SHA256SUMS`.

Windows checksum example:

```powershell
Get-FileHash .\dockyard-windows-amd64.exe -Algorithm SHA256
Get-Content .\SHA256SUMS
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for local development, verification, and pull request expectations.


## OCI package distribution

Dockyard packages can be pushed to OCI registries through ORAS. Dockyard uses:

```text
artifact type: application/vnd.dockyard.package.v1+gzip
archive layer: application/vnd.dockyard.package.archive.v1+gzip
```

See [Packaging and distribution](docs/packaging-and-distribution.md) for GHCR examples.

## Compatibility and package quality checks

```bash
dockyard compat
dockyard compat ./examples/nginx
dockyard compat --release example
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
# Requires Docker daemon / Docker Desktop:
dockyard package test ./examples/nginx --smoke
```

## v1.0 release gate

For packages that should be ready to publish, use strict compatibility and package-quality gates:

```bash
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

See [v1.0 readiness](docs/v1-readiness.md) for the full release checklist.


## Examples

Runnable package examples live under `examples/`:

- `examples/nginx` — basic local smoke-test package.
- `examples/postgres` — reusable PostgreSQL dependency package.
- `examples/postgres-app` — app plus PostgreSQL package.
- `examples/team-dashboard` — dependency installation example.
- `examples/caddy-letsencrypt` — automatic HTTPS with Caddy.
- `examples/nginx-tls-mounted-certs` — TLS with operator-provided certificate/key files.
- `examples/traefik-letsencrypt` — Let's Encrypt with Traefik Docker labels.

The TLS examples show package design patterns. Dockyard does not manage certificates itself; the packages model TLS using standard Docker Compose.

## Local testing layout

Keep generated packages, deployment values, env files, rendered files, and package archives outside this repository.

```text
dockyard/              # Dockyard source repository
../dockyard-work/      # generated test packages
../deploy-values/      # operator values and env files
../dockyard-artifacts/ # generated package archives and rendered files
```

This avoids accidentally committing operator-owned values or secrets.



## Development make targets

`make build` is intentionally non-mutating. If `go.mod` or `go.sum` need updates, run `go mod tidy` explicitly first.

Common targets:

```text
make build      # build bin/dockyard or bin/dockyard.exe
make dev-build  # go mod tidy + gofmt + build for local development
make test       # run go test ./...
make verify     # check go.mod/go.sum, gofmt, and tests without intentionally changing files
make clean      # remove bin/
```

## Quick start

PowerShell:

```powershell
go mod tidy
go test ./...
make build
.\bin\dockyard.exe version

.\bin\dockyard.exe doctor

.\bin\dockyard.exe init ..\dockyard-work\example-app

.\bin\dockyard.exe values template ..\dockyard-work\example-app `
  -o ..\deploy-values\local.yaml

.\bin\dockyard.exe lint ..\dockyard-work\example-app `
  -f ..\deploy-values\local.yaml

.\bin\dockyard.exe render ..\dockyard-work\example-app `
  -f ..\deploy-values\local.yaml `
  --validate-compose

.\bin\dockyard.exe install example ..\dockyard-work\example-app `
  -f ..\deploy-values\local.yaml

.\bin\dockyard.exe status example --compose-ps

Invoke-WebRequest http://localhost:8080 -UseBasicParsing

.\bin\dockyard.exe uninstall example
```

Linux/macOS:

```bash
go mod tidy
go test ./...
go build -o bin/dockyard ./cmd/dockyard

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

## Common workflow

```bash
dockyard values template ../dockyard-work/team-dashboard \
  -o ../deploy-values/dashboard-prod.yaml

dockyard env template ../dockyard-work/team-dashboard \
  -o ../deploy-values/dashboard-prod.env.example

dockyard env check ../deploy-values/dashboard-prod.env.example

dockyard lint ../dockyard-work/team-dashboard \
  -f ../deploy-values/dashboard-prod.yaml

dockyard config ../dockyard-work/team-dashboard \
  -f ../deploy-values/dashboard-prod.yaml \
  --env-file ../deploy-values/dashboard-prod.env

dockyard lock ../dockyard-work/team-dashboard \
  -f ../deploy-values/dashboard-prod.yaml

dockyard package ../dockyard-work/team-dashboard \
  --locked \
  -f ../deploy-values/dashboard-prod.yaml \
  -o ../dockyard-artifacts/team-dashboard-0.1.0.dockyard.tgz

dockyard verify ../dockyard-artifacts/team-dashboard-0.1.0.dockyard.tgz \
  -f ../deploy-values/dashboard-prod.yaml \
  --require-lock

dockyard install dashboard-prod ../dockyard-artifacts/team-dashboard-0.1.0.dockyard.tgz \
  -f ../deploy-values/dashboard-prod.yaml \
  --env-file ../deploy-values/dashboard-prod.env \
  --require-lock
```

## Useful commands

```bash
dockyard doctor
dockyard init ../dockyard-work/example-app
dockyard values template ../dockyard-work/example-app -o ../deploy-values/prod.yaml
dockyard env template ../dockyard-work/example-app -o ../deploy-values/example.env.example
dockyard lint ../dockyard-work/example-app -f ../deploy-values/prod.yaml
dockyard render ../dockyard-work/example-app -f ../deploy-values/prod.yaml --explain
dockyard config ../dockyard-work/example-app -f ../deploy-values/prod.yaml
dockyard policy list
dockyard policy check ../dockyard-work/example-app -f ../deploy-values/prod.yaml
dockyard secrets scan ../dockyard-work/example-app --strict
dockyard lock ../dockyard-work/example-app -f ../deploy-values/prod.yaml
dockyard package ../dockyard-work/example-app --locked -f ../deploy-values/prod.yaml -o ../dockyard-artifacts/example-app-0.1.0.dockyard.tgz
dockyard verify ../dockyard-artifacts/example-app-0.1.0.dockyard.tgz -f ../deploy-values/prod.yaml --require-lock
dockyard install myapp ../dockyard-artifacts/example-app-0.1.0.dockyard.tgz -f ../deploy-values/prod.yaml --require-lock
dockyard status myapp --compose-ps
dockyard inspect myapp
dockyard list               # active releases by default
dockyard list --all         # include uninstalled release history
dockyard uninstall myapp --dry-run
dockyard prune --dry-run
```

## Build with Make

```bash
make tidy
make test
make build
```

`make build` uses `go env GOEXE`, so it writes `bin/dockyard.exe` on Windows and `bin/dockyard` on Linux/macOS.

## Notes

- OCI push/pull uses the `oras` CLI. Run `dockyard doctor` to check whether `oras` is available.
- Package archives reject common secret-like files such as `.env`, `*.pem`, `*.key`, `id_rsa`, and `id_ed25519`.
- The starter package generated by `dockyard init` is intentionally runnable with stock `nginx`; stricter hardening is documented in [Security](docs/security.md).


## Install from release artifacts

Windows PowerShell:

```powershell
Invoke-WebRequest `
  -Uri https://github.com/nandub/dockyard/releases/latest/download/dockyard-windows-amd64.exe `
  -OutFile dockyard.exe

.\dockyard.exe doctor
.\dockyard.exe version
```

Linux amd64:

```bash
curl -L -o dockyard https://github.com/nandub/dockyard/releases/latest/download/dockyard-linux-amd64
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

## OCI package distribution

Dockyard packages can be pushed to OCI registries through ORAS. Dockyard uses:

```text
artifact type: application/vnd.dockyard.package.v1+gzip
archive layer: application/vnd.dockyard.package.archive.v1+gzip
```

See [Packaging and distribution](docs/packaging-and-distribution.md) for GHCR examples.

## Compatibility and package quality checks

```bash
dockyard compat
dockyard compat ./examples/nginx
dockyard compat --release example
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
# Requires Docker daemon / Docker Desktop:
dockyard package test ./examples/nginx --smoke
```

## v1.0 release gate

For packages that should be ready to publish, use strict compatibility and package-quality gates:

```bash
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

See [v1.0 readiness](docs/v1-readiness.md) for the full release checklist.


## Examples

Runnable example packages live in `examples/`:

```text
examples/nginx
examples/postgres-app
```

Keep local smoke-test artifacts outside this repository, for example:

```text
../dockyard-work/
../deploy-values/
../dockyard-artifacts/
```

## Documentation

- [Getting started](docs/getting-started.md)
- [Operator guide](docs/operator-guide.md)
- [Packaging and distribution](docs/packaging-and-distribution.md)
- [Security](docs/security.md)
- [Docker Compose compatibility](docs/compose-compatibility.md)
- [Command reference](docs/command-reference.md)
- [Release engineering](docs/release-engineering.md)
- [Real-world example](docs/real-world-example.md)


## Strict package gates

Strict mode treats warnings as failures for release-candidate checks:

```bash
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

For private/internal packages that intentionally rely on repository-level licensing instead of a package-local `LICENSE`, package quality commands support `--allow-advisory`.

### Dependency dry runs

`dockyard install-plan RELEASE PACKAGE_SOURCE` and `dockyard install --dry-run RELEASE PACKAGE_SOURCE` use the same read-only planner. Use either command to preview dependency release names, existing release actions, and root package installation order before Dockyard supports automatic dependency installation.

```bash
dockyard install --dry-run team-dashboard ./examples/team-dashboard
dockyard install --dry-run team-dashboard ./examples/team-dashboard --json
```

### Dependency installation

Dockyard supports explicit dependency installation for packages that declare `dependencies:` in `Dockyard.yaml`.
Preview first:

```sh
dockyard install-plan team-dashboard ./examples/team-dashboard
# or
dockyard install --dry-run team-dashboard ./examples/team-dashboard
```

Install dependencies before the root package with explicit opt-in:

```sh
dockyard install --with-dependencies team-dashboard ./examples/team-dashboard
```

Dependency installs are conservative: existing deployed dependency releases are reused, uninstalled dependency releases are reinstalled, dependency inline values are applied to dependency packages, and dependencies are not automatically removed on root uninstall.


Failed or pending dependency releases block automatic dependency installation; resolve them before re-running `dockyard install --with-dependencies`.
