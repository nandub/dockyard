# Runtime Model

This document describes how Dockyard executes at runtime, based on static inspection of the codebase.

## Startup Sequence

Startup is intentionally small:

1. `cmd/dockyard/main.go` calls `cli.NewRootCommand().Execute()`.
2. `internal/cli/root.go` constructs a Cobra root command.
3. The root command registers all subcommands.
4. Cobra parses arguments and flags.
5. Cobra dispatches to the matching command's `RunE`.
6. If command execution returns an error, `main` exits with status code `1`.

Evidence:

- `cmd/dockyard/main.go`
- `internal/cli/root.go`

There is no application container, long-running daemon, background worker, or server process in the discovered runtime path.

## Command Dispatch

Command dispatch is handled by Cobra.

The root command is:

```text
dockyard
```

Registered commands include:

- `init`
- `lint`
- `render`
- `config`
- `catalog`
- `install`
- `install-plan`
- `uninstall`
- `list`
- `status`
- `inspect`
- `diff`
- `upgrade`
- `rollback`
- `doctor`
- `lock`
- `values`
- `package`
- `verify`
- `push`
- `pull`
- `policy`
- `secrets`
- `env`
- `prune`
- `version`
- `compat`

Nested command groups include:

- `catalog list`
- `catalog info`
- `values template`
- `values validate`
- `values schema`
- `package deps`
- `package lint`
- `package test`
- `policy list`
- `policy check`
- `secrets scan`
- `env template`
- `env check`

## Dependency Injection

Dockyard does not use a dependency injection framework.

Dependencies are constructed directly in command handlers and helper functions:

- `state.Home(global.home)` resolves Dockyard home.
- `preparePackageSource(...)` resolves local, archive, OCI, and catalog package sources.
- `buildPackage(...)` loads manifest, values, schema, render output, lockfile, and policy checks.
- `runner.DockerComposeRunner{...}` is constructed where Docker Compose is needed.
- `context.WithTimeout(...)` is used for bounded Docker/ORAS operations.

The only global option structure found is:

```go
type globalOptions struct {
    home string
}
```

It is passed into commands that need state-home resolution.

## Configuration Loading

Dockyard loads configuration from several layers:

1. CLI flags.
2. Environment variables for selected global behavior.
3. Package files.
4. Release state files.
5. External tool configuration owned by Docker, Docker Compose, and ORAS.

Package configuration:

- `Dockyard.yaml` is loaded by `dockpkg.LoadManifest`.
- `values.yaml` is loaded by `values.LoadValues`.
- An optional `--values` file is loaded and merged over package defaults.
- `values.schema.json` is loaded when present and used for validation.
- Compose base and overlay files are selected from `Dockyard.yaml`.

Release configuration:

- Release metadata is stored in `release.json`.
- Current revision is stored in a `current` file.
- Rendered Compose is stored as `compose.rendered.yaml`.

## Configuration Precedence

Observed precedence:

- Dockyard home: `--home` flag, then `DOCKYARD_HOME`, then `~/.dockyard`.
- Catalog source: `DOCKYARD_CATALOG`, then `oci://ghcr.io/nandub/dockyard-packages/catalog:latest`.
- Values: package `values.yaml`, then optional CLI `--values` override.
- Dependency inline values: only used for dependency package installs; they are written to a temporary values file and cannot be combined with an existing dependency `--values` setting.
- Dependency installs do not reuse the root package `--values` or `--overlay`.
- Compose overlay: no overlay unless `--overlay` is passed.
- Package source: existing local paths, archives, and explicit `oci://` references take precedence over catalog shorthand.

## Package Loading

Package source preparation is centralized in `preparePackageSourceWithOptions`.

Package source forms:

- Local directory.
- Local archive: `.dockyard.tgz`, `.tgz`, or `.tar.gz`.
- Explicit OCI reference: `oci://...`.
- Explicit catalog reference: `catalog://NAME[:VERSION]`.
- Bare catalog shorthand.

Resolution flow:

1. Resolve catalog shorthand when applicable.
2. If the source is OCI, pull it with ORAS into a temporary directory.
3. If source verification is requested, verify the archive.
4. Extract archives into a temporary source directory.
5. Use local directories directly.
6. Return a prepared source with cleanup behavior.

Package build flow:

1. Load `Dockyard.yaml`.
2. Load and merge values.
3. Validate values against `values.schema.json`, when present.
4. Render Compose.
5. Verify `dockyard.lock`, when `--require-lock` is set.
6. Run policy checks unless skipped.
7. Fail high-severity policy findings unless `--allow-risk` is set.

## Registry Communication

Dockyard does not link a registry client library directly.

Registry communication for package artifacts is delegated to the external `oras` CLI.

Catalog metadata can be loaded from:

- `file://...`
- local `.yaml` or `.yml`
- `oci://...`

For OCI catalog metadata:

1. Dockyard checks that `oras` exists in `PATH`.
2. Dockyard runs `oras pull`.
3. It expects `catalog.yaml` or `catalog.yml` in the pulled artifact.
4. It parses and validates the catalog index.
5. It caches valid catalog metadata under `~/.dockyard/cache/catalogs` for up to five minutes.

## OCI Communication

Package OCI operations are implemented in `internal/oci`.

OCI pull:

1. Validate that the input starts with `oci://`.
2. Require an explicit tag or digest.
3. Require `oras` in `PATH`.
4. Create the output directory.
5. Run `oras pull <ref> -o <dir>`.
6. Find exactly one archive in the pull output.

OCI push:

1. Validate the `oci://` reference.
2. Require an explicit tag or digest.
3. Require `oras` in `PATH`.
4. Resolve the archive path.
5. Run `oras push --artifact-type application/vnd.dockyard.package.v1+gzip ...`.

Quiet pull mode:

- `--json` install-plan/dry-run flows can use quiet OCI pulls so stdout remains machine-readable JSON.

## Plugin Loading

No plugin loading system was found.

There is no discovered runtime plugin registry, plugin directory scan, dynamic loading, package hook execution, or script execution model.

Extension behavior is data-based:

- Dockyard packages.
- Values.
- Compose overlays.
- Package dependencies.
- OCI/catalog references.

## Logging and Output

No structured logging framework was found.

Observed output model:

- Human command output uses `fmt.Print`, `fmt.Println`, and `fmt.Printf`.
- JSON output uses `encoding/json` with indentation.
- Docker subprocess stdout is usually passed through for visible commands.
- Docker Compose config validation sends stdout to `io.Discard`.
- Docker subprocess stderr is passed to `os.Stderr`.
- ORAS push/pull normally passes stdout/stderr through.
- Quiet ORAS pull mode discards stdout/stderr.

Sensitive output controls:

- Render diagnostics mask sensitive keys.
- Env-file helpers warn against real production secret files.
- Release state records the env-file path, not env-file values.

## Lifecycle

### Install

Install lifecycle:

1. Parse release/source arguments.
2. Resolve catalog shorthand if applicable.
3. Validate release name.
4. If `--dry-run`, build and print an install plan.
5. If `--with-dependencies`, plan and install dependencies first.
6. Prepare package source.
7. Resolve Dockyard home.
8. Determine revision number.
9. Load optional env-file entries for subprocess use.
10. Build package.
11. Write pending release revision.
12. Optionally validate with `docker compose config`.
13. Run `docker compose up -d`.
14. Mark release `deployed`.
15. Write current revision pointer.

If Compose validation or `up` fails, Dockyard marks the revision `failed`.

### Upgrade

Upgrade lifecycle:

1. Validate release name.
2. Prepare package source.
3. Resolve Dockyard home.
4. Require an existing current revision.
5. Load optional env-file entries.
6. Build package.
7. Allocate next revision.
8. If `--dry-run`, print the planned upgrade.
9. Write pending release revision.
10. Optionally validate with `docker compose config`.
11. Run `docker compose up -d`.
12. Mark release `deployed`.
13. Update current revision pointer.

### Uninstall

Uninstall lifecycle:

1. Resolve Dockyard home.
2. Read current release and rendered Compose path.
3. Check active dependent releases.
4. Block uninstall of active dependencies unless `--force`.
5. If `--dry-run`, print the Compose down command.
6. Run `docker compose down`, optionally with `--volumes`.
7. Mark release `uninstalled`.
8. If `--purge`, remove release metadata directory.

### Package Test Smoke

Package smoke tests use a temporary Compose project and run Compose up/down without writing Dockyard release state.

## Shutdown

Dockyard is a short-lived CLI process.

Shutdown behavior:

- Command returns `nil`: process exits successfully.
- Command returns an error: Cobra returns the error and `main` exits with status `1`.
- Temporary directories are cleaned with deferred cleanup functions where used.
- Context cancellations are deferred after timeout-bound Docker/ORAS operations.
- No signal handling or graceful daemon shutdown path was found.

## Error Handling

Error handling is explicit Go error returns.

Patterns:

- Functions return contextual errors with `fmt.Errorf`.
- Cobra command handlers return errors from `RunE`.
- External Docker command failures are summarized as `docker command failed`.
- ORAS push/pull failures are summarized as `oras push failed` or `oras pull failed`.
- Some state failures during failure handling are intentionally ignored after marking revisions failed.
- Archive and package path errors are fail-closed.

No `panic` use was found in normal library runtime paths during this phase.

## Runtime Timeouts

Timeouts found:

- General Docker/ORAS command helper context: 10 minutes.
- Catalog OCI pull context: 2 minutes.
- Doctor Docker checks: 15 seconds.

## Evidence

- `cmd/dockyard/main.go`
- `internal/cli/root.go`
- `internal/cli/common.go`
- `internal/cli/install.go`
- `internal/cli/upgrade.go`
- `internal/cli/uninstall.go`
- `internal/state/state.go`
- `internal/dockpkg/package.go`
- `internal/values/values.go`
- `internal/render/render.go`
- `internal/runner/docker.go`
- `internal/catalog/catalog.go`
- `internal/oci/oci.go`
- `internal/archive/archive.go`
