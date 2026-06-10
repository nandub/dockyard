# Command reference

Dockyard commands are grouped around packaging, validation, release management, and operator ergonomics.

## Global flags

```text
--home string   Dockyard home directory; overrides DOCKYARD_HOME
-h, --help      command help
```

## Core package commands

```bash
dockyard init DIRECTORY
dockyard lint PACKAGE_DIR [-f values.yaml] [--overlay name]
dockyard render PACKAGE_DIR [-f values.yaml] [--validate-compose] [--env-file file]
dockyard config PACKAGE_SOURCE [-f values.yaml] [--env-file file]
```

`render` prints Dockyard's rendered Compose YAML. `config` prints Docker Compose's normalized configuration.

## Values, env, and secrets

```bash
dockyard values template PACKAGE_DIR -o values.yaml [--force]
dockyard values validate PACKAGE_DIR -f values.yaml
dockyard values schema PACKAGE_DIR

dockyard env template PACKAGE_DIR -o app.env.example [--sensitive-only] [--prefix PREFIX_] [--force]
dockyard env check ENV_FILE

dockyard secrets scan PACKAGE_DIR [-f values.yaml] [--strict] [--json]
```

Use values files for deployment settings. Use `--env-file` for environment-backed secrets that Compose must resolve.

## Packaging and distribution

```bash
dockyard lock PACKAGE_DIR [-f values.yaml]
dockyard package lint PACKAGE_DIR [--strict] [--allow-advisory] [--json]
dockyard package test PACKAGE_SOURCE [--strict] [--allow-advisory] [--smoke] [--env-file file]
dockyard package PACKAGE_DIR --locked [-f values.yaml] -o app-0.1.0.dockyard.tgz
dockyard verify PACKAGE_ARCHIVE [-f values.yaml] [--require-lock]
dockyard push PACKAGE_ARCHIVE oci://registry/repository/name:tag
dockyard pull oci://registry/repository/name:tag
```

Run `dockyard package lint --strict` before publishing packages. It checks package documentation, forbidden local artifacts, schema quality, sensitive markers, default rendering, and policy findings.

Run `dockyard package test` for a fuller package-author pipeline. It prepares local directories, archives, or OCI sources, runs quality checks, renders with selected values, runs Dockyard policy checks, and validates the result with `docker compose config`. Add `--smoke` for safe examples that can be started and stopped with a temporary Compose project name. Smoke tests require a reachable Docker daemon; run `dockyard doctor` first when troubleshooting Docker Desktop or daemon connectivity.

OCI push/pull uses the `oras` CLI and relies on external registry authentication.

## Release lifecycle

```bash
dockyard install RELEASE PACKAGE_SOURCE [-f values.yaml] [--env-file file]
dockyard diff RELEASE PACKAGE_SOURCE [-f values.yaml]
dockyard upgrade RELEASE PACKAGE_SOURCE [-f values.yaml] [--env-file file]
dockyard rollback RELEASE REVISION
dockyard status RELEASE [--compose-ps] [--all] [--json]
dockyard inspect RELEASE [--revision N] [--json]
dockyard list
dockyard uninstall RELEASE [--volumes] [--purge] [--dry-run]
dockyard prune [--release RELEASE] [--keep N] [--dry-run]
```

`PACKAGE_SOURCE` may be a local package directory, a `.dockyard.tgz` archive, or an `oci://` reference.

## Policy and diagnostics

```bash
dockyard doctor
dockyard policy list [--json]
dockyard policy check PACKAGE_SOURCE [-f values.yaml]
dockyard version [--json]
```

`doctor` checks local prerequisites such as Docker, Docker Compose, Dockyard home, and optional `oras`.


## `dockyard compat`

Show supported Dockyard format versions or check a package/release for compatibility issues.

```bash
dockyard compat
dockyard compat ./examples/nginx
dockyard compat ./examples/nginx --strict
dockyard compat nginx-0.1.0.dockyard.tgz
dockyard compat --release example
dockyard compat --json
```

The command is intended for v1.0 readiness work and for diagnosing package, archive, lockfile, and release-state format issues.


## Release-candidate checks

For v1.0 release-candidate preparation, run:

```bash
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

Use `--strict` to treat warnings as failures for release-candidate gates. `package lint` and `package test` also support `--allow-advisory` for advisory warnings such as a missing package-local `LICENSE` in private/internal packages.


### Strict package gates

`--strict` has the same meaning across compatibility and package-quality commands: warnings become command failures.

```bash
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

For private/internal packages that intentionally rely on repository-level licensing instead of a package-local `LICENSE`, use:

```bash
dockyard package lint ./internal-package --strict --allow-advisory
dockyard package test ./internal-package --strict --allow-advisory
```

Public examples in this repository include package-local `LICENSE` files so the strict gate can pass without `--allow-advisory`.
