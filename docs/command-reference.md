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
dockyard package PACKAGE_DIR --locked [-f values.yaml] -o app-0.1.0.dockyard.tgz
dockyard verify PACKAGE_ARCHIVE [-f values.yaml] [--require-lock]
dockyard push PACKAGE_ARCHIVE oci://registry/repository/name:tag
dockyard pull oci://registry/repository/name:tag
```

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
