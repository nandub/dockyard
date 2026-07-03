# Command Reference

Dockyard commands are grouped around packaging, validation, release management, dependency operations, and operator diagnostics.

## Global flags

```text
--home string   Dockyard home directory; overrides DOCKYARD_HOME
-h, --help      command help
```

## Catalog commands

```bash
dockyard catalog list [--json]
dockyard catalog info PACKAGE [--json]
```

The default catalog registry is `ghcr.io/nandub/dockyard-packages`. Override it with:

```bash
export DOCKYARD_CATALOG=ghcr.io/my-org/my-dockyard-packages
```

Catalog source forms:

```bash
dockyard install redis
dockyard install my-cache redis
dockyard install redis catalog://redis:0.1.0
```

These resolve to OCI references under the configured catalog registry.

## Core package commands

```bash
dockyard init DIRECTORY
dockyard lint PACKAGE_DIR [-f values.yaml] [--overlay name]
dockyard render PACKAGE_DIR [-f values.yaml] [--validate-compose] [--env-file file]
dockyard config PACKAGE_SOURCE [-f values.yaml] [--env-file file]
```

`render` prints Dockyard's rendered Compose YAML. `config` prints Docker Compose's normalized configuration after validation by `docker compose config`.

## Values, env, and secrets

```bash
dockyard values template PACKAGE_DIR -o values.yaml [--force]
dockyard values validate PACKAGE_DIR -f values.yaml
dockyard values schema PACKAGE_DIR

dockyard env template PACKAGE_DIR -o app.env.example [--sensitive-only] [--prefix PREFIX_] [--force]
dockyard env check ENV_FILE

dockyard secrets scan PACKAGE_DIR [-f values.yaml] [--strict] [--json]
```

Use values files for deployment settings. Use `--env-file` for environment-backed secrets that Compose must resolve. Dockyard records the env-file path in release metadata, not the secret values.

## Packaging and distribution

```bash
dockyard lock PACKAGE_DIR [-f values.yaml]
dockyard package lint PACKAGE_DIR [--strict] [--allow-advisory] [--json]
dockyard package deps PACKAGE_SOURCE [--json]
dockyard package test PACKAGE_SOURCE [--strict] [--allow-advisory] [--smoke] [--env-file file]
dockyard package PACKAGE_DIR [--locked] [-f values.yaml] -o app-0.1.0.dockyard.tgz
dockyard verify PACKAGE_ARCHIVE [-f values.yaml] [--require-lock]
dockyard push PACKAGE_ARCHIVE oci://registry/repository/name:tag
dockyard pull oci://registry/repository/name:tag [-o PACKAGE_ARCHIVE]
```

Run `dockyard package lint --strict` before publishing packages. It checks package documentation, forbidden local artifacts, dependency metadata, schema quality, sensitive markers, default rendering, and policy findings.

Run `dockyard package deps` to inspect declared dependencies in `Dockyard.yaml`.

Run `dockyard package test` for a fuller package-author pipeline. It prepares local directories, archives, or OCI sources, runs quality checks, renders with selected values, runs policy checks, and validates the result with `docker compose config`. Add `--smoke` for examples that can be started and stopped with a temporary Compose project name.

OCI push/pull uses the `oras` CLI and relies on external registry authentication. Dockyard publishes packages with:

```text
artifact type: application/vnd.dockyard.package.v1+gzip
archive layer: application/vnd.dockyard.package.archive.v1+gzip
```

## Dependency planning and installation

```bash
dockyard install-plan RELEASE PACKAGE_SOURCE [--json]
dockyard install RELEASE PACKAGE_SOURCE [--dry-run] [--json]
dockyard install PACKAGE
# or
 dockyard install RELEASE PACKAGE_SOURCE [--with-dependencies]
# PACKAGE_SOURCE may be a path, archive, oci:// reference, catalog:// reference, or catalog package name
```

`dockyard install-plan` and `dockyard install --dry-run` use the same read-only dependency-aware planner. Use either command to preview dependency release names, existing-release actions, and root package installation order. Add `--json` for automation.

`dockyard install --with-dependencies` installs declared dependencies before the root package. Plain `dockyard install` installs only the root package.

Dependency behavior:

- release names are deterministic, such as `RELEASE-ALIAS`;
- existing deployed dependency releases are reused;
- uninstalled dependency releases are reinstalled as a new revision;
- dependency inline `values:` are applied only to dependency package installs;
- root `--values` and `--overlay` are not reused for dependencies;
- failed or pending dependency releases block automatic dependency installation;
- dependencies are not automatically removed if a later step fails or when the root release is uninstalled.

Example:

```bash
dockyard install-plan team-dashboard ./examples/team-dashboard
dockyard install --with-dependencies team-dashboard ./examples/team-dashboard
```

## Release lifecycle

```bash
dockyard install PACKAGE [-f values.yaml] [--env-file file]
dockyard install RELEASE PACKAGE_SOURCE [-f values.yaml] [--env-file file] [--require-lock]
dockyard diff RELEASE PACKAGE_SOURCE [-f values.yaml]
dockyard upgrade RELEASE PACKAGE_SOURCE [-f values.yaml] [--env-file file]
dockyard rollback RELEASE REVISION
dockyard status RELEASE [--compose-ps] [--all] [--json]
dockyard inspect RELEASE [--revision N] [--json]
dockyard list [--all] [--status STATUS]
dockyard uninstall RELEASE [--volumes] [--purge] [--dry-run] [--force]
dockyard prune [--release RELEASE] [--keep N] [--dry-run]
```

`PACKAGE_SOURCE` may be a local package directory, a `.dockyard.tgz` archive, an `oci://` reference, a `catalog://PACKAGE[:VERSION]` reference, or a known catalog package name. With one argument, `dockyard install redis` installs the configured catalog package `redis` as release `redis`.

### `dockyard list`

`dockyard list` shows active releases by default. Use `--all` to include uninstalled release history and `--status STATUS` to filter by a state such as `deployed`, `uninstalled`, `failed`, or `pending`.

```bash
dockyard list
dockyard list --all
dockyard list --status uninstalled
```

The `RELATION` column shows dependency relationships:

```text
-                      standalone release
deps=N                 root release with N dependencies
child-of=RELEASE       dependency release
```

### `dockyard status`

`dockyard status RELEASE` prints the release status, revision, package, app version, Compose project, and Dockyard version. For relationship-aware releases, status also prints dependency references for root releases and parent information for dependency releases.

```bash
dockyard status team-dashboard
dockyard status team-dashboard-db --compose-ps
```

### Dependency uninstall safety

When a release is recorded as a dependency of an active parent release, Dockyard blocks direct uninstall by default. Uninstall the parent release first, then uninstall the dependency release. Use `--force` only when intentionally breaking that relationship, such as during manual recovery.

```powershell
dockyard uninstall team-dashboard
dockyard uninstall team-dashboard-db

dockyard uninstall team-dashboard-db --force
```

## Policy and diagnostics

```bash
dockyard doctor
dockyard policy list [--json]
dockyard policy check PACKAGE_SOURCE [-f values.yaml]
dockyard compat [PACKAGE_SOURCE] [--release RELEASE] [--strict] [--json]
dockyard version [--json]
```

`doctor` checks local prerequisites such as Docker, Docker Compose, Dockyard home, and optional `oras`.

`compat` shows supported Dockyard format versions or checks a package/release for compatibility issues. Use `--strict` for release gates.

## Package quality gates

Use these commands before publishing a package or cutting a Dockyard release:

```bash
dockyard compat PACKAGE_DIR --strict
dockyard package lint PACKAGE_DIR --strict
dockyard package test PACKAGE_DIR --strict
dockyard package test PACKAGE_DIR --smoke
```

`--strict` treats warnings as failures. `package lint` and `package test` also support `--allow-advisory` for private/internal packages that intentionally accept advisory warnings such as a missing package-local `LICENSE`.

Public examples in this repository include package-local `LICENSE` files so strict gates can pass without `--allow-advisory`.

## Release artifact checks

GitHub releases publish:

```text
dockyard-windows-amd64.exe
dockyard-linux-amd64
dockyard-linux-arm64
dockyard-darwin-amd64
dockyard-darwin-arm64
dockyard-source.spdx.json
SHA256SUMS
```

`SHA256SUMS` includes the five binaries and the SBOM.
