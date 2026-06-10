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
dockyard package lint PACKAGE_DIR [--strict] [--json]
dockyard package test PACKAGE_SOURCE [--strict] [--smoke] [--env-file file]
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
dockyard compat nginx-0.1.0.dockyard.tgz
dockyard compat --release example
dockyard compat --json
```

The command is intended for v1.0 readiness work and for diagnosing package, archive, lockfile, and release-state format issues.
