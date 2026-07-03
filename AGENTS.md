# AGENTS.md

This file gives coding agents enough project context to make safe, idiomatic changes to Dockyard.

Dockyard is a Go CLI that adds a package, render, validation, release-state, policy, and distribution layer on top of Docker Compose. Docker Compose remains the runtime source of truth.

## Catalog source behavior

Dockyard v1.7 supports the official package catalog.

- `catalog://NAME[:VERSION]` resolves to `oci://$DOCKYARD_CATALOG/NAME:VERSION`.
- If `DOCKYARD_CATALOG` is unset, use `ghcr.io/nandub/dockyard-packages`.
- Bare install shorthand such as `dockyard install redis` is allowed only for known catalog packages.
- Existing local paths, archives, and explicit `oci://` references must keep precedence over catalog shorthand.
- JSON output must remain machine-readable. When planning requires OCI pulls, use quiet pull paths for `--json` modes.

## Repository basics

- Module: `github.com/nandub/dockyard`
- Language: Go
- Current Go version in `go.mod`: `1.23`
- CLI entry point: `cmd/dockyard/main.go`
- Main CLI framework: Cobra
- YAML package: `go.yaml.in/yaml/v4`
- JSON Schema package: `github.com/santhosh-tekuri/jsonschema/v6`
- OCI support: currently shells out to the `oras` CLI rather than linking a registry client library directly.

## Common commands

Run these from the repository root.

```sh
go mod tidy
go fmt ./...
go test ./...
go build -o bin/dockyard$(go env GOEXE) ./cmd/dockyard
```

The Makefile wraps the common commands:

```sh
make tidy
make fmt
make test
make build
make clean
```

On Windows, `go env GOEXE` returns `.exe`, so `make build` should produce:

```text
bin/dockyard.exe
```

On Linux and macOS, it should produce:

```text
bin/dockyard
```

## Example dependency packages

`examples/postgres` is the reusable PostgreSQL dependency package used by `examples/team-dashboard`. Keep it valid with `dockyard package lint ./examples/postgres --strict` and `dockyard package test ./examples/postgres --strict` before changing dependency-install behavior.

## Local smoke-test layout

Keep generated test packages, deployment values, rendered files, and archives outside the Dockyard repository to avoid accidental commits.

Recommended layout:

```text
dockyard/              # this source repository
../dockyard-work/      # generated Dockyard packages
../deploy-values/      # local operator values and env files
../dockyard-artifacts/ # generated package archives or rendered files
```

Example Windows smoke test:

```powershell
go mod tidy
go test ./...
go build -o bin/dockyard.exe ./cmd/dockyard

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

PowerShell aliases `curl` to `Invoke-WebRequest` in some environments. Use either:

```powershell
Invoke-WebRequest http://localhost:8080 -UseBasicParsing
```

or:

```powershell
curl.exe http://localhost:8080
```

## Architecture overview

Important packages:

```text
cmd/dockyard/       CLI entry point
internal/cli/       Cobra commands and command wiring
internal/dockpkg/   Dockyard package manifest loading and path safety
internal/values/    values.yaml loading, merge, and JSON Schema validation
internal/render/    placeholder rendering and render diagnostics
internal/policy/    Compose security policy checks
internal/state/     Dockyard home, release, revision, and metadata handling
internal/runner/    Docker and Docker Compose subprocess integration
internal/archive/   .dockyard.tgz package creation and verification helpers
internal/lock/      dockyard.lock creation and verification
internal/oci/       OCI push/pull integration through the oras CLI
internal/envfile/   .env template and env-file validation helpers
docs/               consolidated user documentation
```

Keep CLI code thin. Prefer putting testable logic in internal packages and calling that logic from Cobra commands.

## Product model

Dockyard should not replace Docker Compose.

Dockyard responsibilities:

- Load `Dockyard.yaml`.
- Load and merge `values.yaml` plus optional override files.
- Validate values against `values.schema.json`.
- Render standard Docker Compose YAML.
- Optionally validate rendered YAML with `docker compose config`.
- Run selected security policy checks.
- Store release revision state under `DOCKYARD_HOME`.
- Package, verify, lock, push, and pull Dockyard packages.
- Delegate runtime behavior to Docker Compose.

Docker Compose responsibilities:

- Interpret the final Compose YAML.
- Pull images.
- Create networks, volumes, and containers.
- Start, stop, and inspect runtime services.

Do not claim 100% internal Compose-spec support. Dockyard renders standard Compose YAML and validates selected fields, but it does not fully model every Compose feature internally.

## State model

Dockyard state is stored outside the package source directory.

Resolution order:

```text
1. --home flag
2. DOCKYARD_HOME environment variable
3. ~/.dockyard
```

Release layout:

```text
~/.dockyard/
  releases/
    <release>/
      current
      revisions/
        <revision>/
          Dockyard.yaml
          values.yaml
          compose.rendered.yaml
          release.json
          dockyard.lock   # when present/required
```

Do not store secret values in `release.json`. Recording an env-file path is acceptable; copying secret env files into release state is not.

## Security rules

Follow these rules for all changes:

- Never hardcode credentials, tokens, API keys, or real secrets.
- Do not print secret values in normal output, diagnostics, logs, or errors.
- Mask values with keys containing words such as `password`, `secret`, `token`, `key`, or `credential`.
- Treat package archives as untrusted input.
- Reject archive entries and package paths that escape the intended directory.
- Be careful with Windows path handling; check both `/` and `\` separators.
- Avoid `panic` in library code. Return explicit errors with useful context.
- Do not execute arbitrary hooks or scripts from packages.
- Be conservative with host-path mounts, Docker socket mounts, privileged containers, and host networking.
- Keep OCI credentials outside Dockyard. The current design relies on `oras` authentication.

## Windows compatibility notes

This repository is actively tested on Windows with Docker Desktop.

Important details:

- Build output must use `go env GOEXE`.
- Path containment checks must handle Windows paths such as `..\secret`.
- Do not hardcode `/tmp`; use `os.MkdirTemp`.
- Do not assume Unix path separators in tests.
- Use `filepath` for filesystem paths.
- Use `path` only for archive-internal slash-separated paths.
- PowerShell line continuations use backticks, not backslashes.

## JSON Schema validation note

`github.com/santhosh-tekuri/jsonschema/v6` expects `Compiler.AddResource` to receive a decoded JSON value, not an `*os.File`.

Correct pattern:

```go
data, err := os.ReadFile(filepath.Clean(schemaPath))
// handle err

var schemaDoc any
if err := json.Unmarshal(data, &schemaDoc); err != nil {
    return fmt.Errorf("parse schema JSON: %w", err)
}

compiler := jsonschema.NewCompiler()
if err := compiler.AddResource("schema.json", schemaDoc); err != nil {
    return fmt.Errorf("load schema: %w", err)
}
```

Do not reintroduce file handles into `AddResource`.

## Generated starter package

The default `dockyard init` package should run successfully on a clean Docker Desktop installation.

The starter app uses `nginx`. Do not over-harden the generated default Compose file with settings that make stock nginx fail, such as:

```yaml
read_only: true
user: "101:101"
cap_drop:
  - ALL
```

Those settings are useful for advanced hardening but require image-specific writable paths or configuration. Keep the starter example runnable first, and document hardening separately.

## CLI behavior guidelines

Good command behavior:

- Refuse to overwrite existing user files unless a `--force` flag exists and is explicitly passed.
- Support `--dry-run` for operations that mutate Docker or release state.
- Give actionable error messages.
- Keep Docker/Compose subprocess errors concise and avoid leaking local environment values.
- Prefer `install` for new or previously uninstalled releases.
- Prefer `upgrade` when a release is already deployed.
- `render --validate-compose` should validate quietly and print only Dockyard's rendered YAML.
- `config` should intentionally print Docker Compose's normalized config output.
- `status --compose-ps` should show Compose status.
- `status --compose-ps --all` should include stopped containers.

## Docs policy

Docs are intentionally consolidated. Avoid creating one new doc file per feature unless the topic is large enough to stand alone.

Current docs structure:

```text
docs/getting-started.md
docs/operator-guide.md
docs/packaging-and-distribution.md
docs/security.md
docs/compose-compatibility.md
docs/real-world-example.md
```

Prefer updating one of these files instead of adding a new file.

When changing commands, update:

- `README.md`
- relevant docs under `docs/`
- `CHANGELOG.md`

Use examples that keep generated artifacts outside the repo:

```powershell
..\dockyard-work\
..\deploy-values\
..\dockyard-artifacts\
```

## Testing expectations

For logic changes, add or update unit tests in the relevant internal package.

Before submitting a change, run:

```sh
go mod tidy
go fmt ./...
go test ./...
go build -o bin/dockyard$(go env GOEXE) ./cmd/dockyard
```

For CLI/runtime changes, also run a smoke test with Docker Desktop or Docker Engine when possible:

```sh
dockyard doctor
dockyard init ../dockyard-work/example-app
dockyard values template ../dockyard-work/example-app -o ../deploy-values/local.yaml
dockyard lint ../dockyard-work/example-app -f ../deploy-values/local.yaml
dockyard render ../dockyard-work/example-app -f ../deploy-values/local.yaml --validate-compose
dockyard install example ../dockyard-work/example-app -f ../deploy-values/local.yaml
dockyard status example --compose-ps
dockyard uninstall example
```

## Dependency policy

Keep dependencies minimal.

Before adding a dependency, consider whether the standard library is sufficient. For new dependencies:

- Prefer mature, maintained packages.
- Do not pin arbitrary versions without checking compatibility.
- Run `go mod tidy`.
- Ensure licenses are acceptable for an open-source CLI.

## Release checklist

For a release-like change:

```sh
go mod tidy
go fmt ./...
go test ./...
make build
```

Then check:

```sh
dockyard --help
dockyard doctor
```

Update:

```text
README.md
CHANGELOG.md
docs/*
```

Do not commit generated local test directories, deployment values, `.env` files, package archives, or `bin/` outputs.

### Windows Makefile notes

The Makefile uses `go env GOEXE` so `make build` emits `bin/dockyard.exe` on Windows and `bin/dockyard` on Unix-like systems. Build metadata must avoid Unix-only shell assumptions; use `$(OS)` conditionals and `NUL` vs `/dev/null` when adding Make targets.


## Makefile targets

Use these targets for local development and CI-style checks:

```text
make build      # build the platform-native binary; non-mutating
make dev-build  # run go mod tidy, gofmt, then build
make verify     # run tidy-check, fmt-check, and tests
make test       # run go test ./...
make clean      # remove bin/
```

`make build` should not update `go.mod` or `go.sum`. Run `go mod tidy` or `make dev-build` explicitly when dependencies change.


## Makefile notes

- `make build` must remain non-mutating. It should not run `go mod tidy` or rewrite source files.
- `make dev-build` is the local convenience path and may run `go mod tidy`, formatting, and build.
- `make fmt-check` uses `go run ./tools/fmtcheck` instead of shell-specific `gofmt -l` loops so it works consistently on Windows, Linux, and macOS.


## v1.0 readiness notes

Prefer compatibility-preserving changes. Keep format API versions centralized in `internal/format`. New release state records should include `apiVersion`; readers should remain tolerant of legacy v0.x records where practical. Update `docs/v1-readiness.md` when changing package, lockfile, provenance, or release-state formats.


## Package quality checks

For package/example changes, run:

```bash
dockyard package lint ./examples/nginx --strict
```

Use this before changing `examples/`, package templates, archive behavior, values/schema generation, or packaging docs. The command checks required package docs, forbidden local artifacts, schema descriptions, sensitive markers, default rendering, and policy findings.


## Package validation pipeline

Use `dockyard package test PACKAGE_SOURCE` when adding or changing example packages. The default mode is non-destructive and runs quality checks, rendering, policy checks, and `docker compose config`. Use `--smoke` only for examples that are safe to start locally and cleanly stop with `docker compose down`.

## Package smoke tests

`dockyard package test --smoke` is allowed only for examples that are safe to start and stop locally. It must use a temporary Compose project name and must not write Dockyard release state. Always preflight Docker availability before smoke operations and return actionable errors that point users to `dockyard doctor` when Docker Desktop or the Docker daemon is not reachable.



## v1.0 / release work

Avoid expanding the CLI surface unless the change directly improves stability, compatibility, release quality, or documentation accuracy.

Before changing file formats, update:

- `internal/format`
- `docs/v1-readiness.md`
- `docs/compose-compatibility.md` when behavior affects Compose support
- `CHANGELOG.md`

For release checks, prefer:

```sh
make verify
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

Use `dockyard package test --smoke` only when Docker is available.


## Strict mode and advisory warnings

`--strict` must mean warnings fail across compatibility and package-quality commands. For package quality checks, use `--allow-advisory` only for explicitly advisory warnings such as private packages that rely on repository-level licensing instead of a package-local `LICENSE`.


## v1.x agent guidance

Prefer changes that improve:

- compatibility checks
- documentation accuracy
- release metadata
- test coverage
- examples
- migration/support notes

Treat `Dockyard.yaml` with `apiVersion: dockyard.dev/v1alpha1` as the stable package manifest contract. Dependency metadata in `Dockyard.yaml` is now operational: it is validated, shown in dependency commands, included in install plans, and installed only when operators explicitly pass `dockyard install --with-dependencies`. Treat `dockyard.lock`, `package.provenance.json`, and `release.json` as experimental generated formats unless instructed otherwise.


## v1.1 release focus

Post-1.0 changes should improve reliability, packaging, release engineering, and adoption without breaking `Dockyard.yaml` `dockyard.dev/v1alpha1`.

For OCI:
- use `application/vnd.dockyard.package.v1+gzip` as the artifact type,
- use `application/vnd.dockyard.package.archive.v1+gzip` as the archive layer media type,
- do not pass absolute archive paths to ORAS,
- keep ORAS authentication outside Dockyard.

Before changing release workflows, ensure:
- `make verify` runs,
- Staticcheck runs,
- release binaries can run `version`,
- the Linux AMD64 release binary can run strict package checks.

## Dependency planning

`dockyard install-plan RELEASE PACKAGE_SOURCE` is intentionally read-only. Keep dependency planning separate from dependency installation until lifecycle behavior is explicitly implemented. The command should validate planned release names, show existing-release status, and avoid writing release state.

## Dependency planner safety

Keep `dockyard install-plan` and `dockyard install --dry-run` aligned. Both commands should use the same planner and remain read-only.

## Dependency installation rules

`dockyard install --with-dependencies` must remain explicit opt-in. Plain `install` installs only the root package. Dependency install behavior should stay conservative: reuse existing deployed dependencies, reinstall only dependencies marked `uninstalled`, do not auto-upgrade dependencies, and do not auto-remove dependencies on root uninstall or partial failure. Keep `install-plan`, `install --dry-run`, and actual dependency installation behavior aligned through tests.


## Dependency relationship metadata

When changing install behavior, preserve relationship metadata in `state.Release.Parent` and `state.Release.Dependencies`. Do not automatically uninstall dependencies unless an explicit future design adds reference counting and strong safety checks.
