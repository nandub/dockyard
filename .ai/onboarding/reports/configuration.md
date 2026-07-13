# Configuration Model

This document describes Dockyard configuration loading, precedence, and persistence.

## Configuration Sources

Dockyard configuration comes from:

- CLI flags.
- Environment variables.
- Package files.
- Values override files.
- Env files passed to Docker Compose subprocesses.
- Catalog metadata.
- Release state.
- External Docker, Docker Compose, and ORAS configuration.

There is no discovered central configuration file for Dockyard itself.

## Global Configuration

### Dockyard Home

Dockyard home controls where release state is stored.

Precedence:

1. `--home`
2. `DOCKYARD_HOME`
3. `~/.dockyard`

Implementation:

- `internal/state/state.go`

State layout:

```text
<home>/
  releases/
    <release>/
      current
      revisions/
        <revision>/
          Dockyard.yaml
          values.yaml
          compose.rendered.yaml
          release.json
          dockyard.lock
```

`dockyard.lock` appears in release state only when present in the package source.

### Catalog Source

Catalog source controls package shorthand resolution.

Precedence:

1. `DOCKYARD_CATALOG`
2. `oci://ghcr.io/nandub/dockyard-packages/catalog:latest`

Accepted `DOCKYARD_CATALOG` forms:

- `oci://...`
- `file://...`
- Local `.yaml` or `.yml` path.
- Registry prefix shorthand, normalized to `oci://<prefix>/catalog:latest`.

Implementation:

- `internal/catalog/catalog.go`

## Package Configuration

Each Dockyard package is configured by `Dockyard.yaml`.

Fields loaded by the manifest model:

- `apiVersion`
- `name`
- `description`
- `version`
- `appVersion`
- `type`
- `compose`
- `security`
- `dependencies`

Required fields enforced by code:

- `apiVersion`
- `name`
- `version`
- `compose.base`

Supported package `type` values:

- empty
- `application`
- `library`

Stable manifest API:

- `dockyard.dev/v1alpha1`

Implementation:

- `internal/dockpkg/package.go`
- `internal/format/format.go`

## Values Configuration

Default values:

- `values.yaml`

Optional override:

- `--values` / `-f`

Precedence:

1. Package `values.yaml`
2. CLI-supplied values override file

Merge behavior:

- Maps are merged recursively.
- Override scalar values replace base values.

Schema validation:

- If `values.schema.json` exists, Dockyard validates merged values against it.
- Missing schema is allowed.

Implementation:

- `internal/values/values.go`

## Compose Configuration

Compose configuration is declared in `Dockyard.yaml`:

```yaml
compose:
  base: compose.yaml
  overlays:
    prod: compose.prod.yaml
```

Precedence:

1. Base Compose file.
2. Optional overlay selected by `--overlay`.

Overlay behavior:

- The overlay name must exist in `compose.overlays`.
- Base and overlay are rendered with values.
- Overlay YAML is merged over base YAML.

Rendering:

- Placeholders use `${path.to.value}` syntax.
- Placeholder names must match the allowed reference pattern.
- Missing values fail rendering.
- Sensitive placeholder diagnostics are masked.

Implementation:

- `internal/render/render.go`

## Env File Configuration

Commands can accept `--env-file` for Docker Compose subprocesses.

Observed commands using env-file behavior include:

- `install`
- `upgrade`
- `config`
- `render --validate-compose`
- `package test`

Env-file behavior:

- Parsed as dotenv-style `KEY=VALUE`.
- Supports blank lines, comments, optional `export ` prefix, and quoted values.
- Rejects duplicate keys.
- Rejects invalid environment variable names.
- Does not mutate the Dockyard process environment.
- Converts entries into `KEY=VALUE` strings appended to subprocess environment.

Release metadata:

- The env-file path is recorded in `release.json`.
- Env-file values are not copied into release state.

Implementation:

- `internal/envfile/envfile.go`
- `internal/cli/common.go`

## Dependency Configuration

Dependencies are declared in `Dockyard.yaml`.

Fields:

- `name`
- `source`
- `alias`
- `version`
- `values`

Validation:

- `name` is required and must match package-name rules.
- `source` is required.
- Duplicate dependency names are rejected.
- Duplicate aliases are rejected.
- OCI dependency sources must include a tag or digest.

Install behavior:

- Plain `install` installs only the root package.
- `install --with-dependencies` installs dependencies first.
- Dependency release names are deterministic.
- Root `--values` and `--overlay` are not reused for dependencies.
- Dependency inline `values:` are written to a temporary values file.
- Dependency inline values cannot be combined with an existing values file.

Implementation:

- `internal/dockpkg/package.go`
- `internal/cli/install_plan.go`
- `internal/cli/install.go`

## Source Resolution Configuration

Package source resolution supports:

- Local directory.
- Local archive: `.dockyard.tgz`, `.tgz`, `.tar.gz`.
- Explicit `oci://...`.
- Explicit `catalog://NAME[:VERSION]`.
- Bare catalog package shorthand.

Precedence:

1. Existing local path.
2. Archive path.
3. Explicit OCI reference.
4. Catalog resolution.

This keeps local paths, archives, and explicit OCI references ahead of catalog shorthand.

Implementation:

- `internal/cli/common.go`
- `internal/catalog/catalog.go`
- `internal/oci/oci.go`

## Release State Configuration

Release state records:

- API version.
- Dockyard version.
- Release name.
- Package identity.
- App version.
- Revision.
- Status.
- Timestamps.
- Compose project name.
- Source.
- Env-file path.
- Parent/dependency relationship metadata.

Statuses observed in code:

- `pending`
- `deployed`
- `failed`
- `uninstalled`

Current revision:

- Stored in `<home>/releases/<release>/current`.

Revision contents:

- `Dockyard.yaml`
- `values.yaml`
- optional `dockyard.lock`
- `compose.rendered.yaml`
- `release.json`

Implementation:

- `internal/state/state.go`
- `internal/cli/common.go`
- `internal/cli/install.go`
- `internal/cli/upgrade.go`
- `internal/cli/uninstall.go`

## Command Flags Affecting Configuration

Common package build flags:

- `--values`, `-f`
- `--env-file`
- `--overlay`
- `--allow-risk`
- `--skip-policy`
- `--skip-compose-config`
- `--require-lock`

Read-only or planning flags:

- `--dry-run`
- `--json`

Lifecycle flags:

- `--with-dependencies`
- `--volumes`
- `--purge`
- `--force`

## External Tool Configuration

Docker and Compose:

- Dockyard shells out to `docker`.
- Docker and Docker Compose configuration are owned by the user's Docker environment.
- `docker compose` inherits `os.Environ()` plus optional parsed env-file entries.

ORAS:

- Dockyard shells out to `oras`.
- Registry authentication is handled outside Dockyard by ORAS.
- Dockyard does not accept registry credentials directly in the discovered code.

## Logging Configuration

No runtime logging configuration was found.

No log levels, log config file, or structured logger setup were discovered.

Output behavior is command-specific:

- Human output goes to stdout.
- JSON output goes to stdout.
- Subprocess stderr generally goes to stderr.
- Some validation stdout is discarded.
- Quiet OCI pull discards stdout and stderr.

## Plugin Configuration

No plugin configuration was found.

No plugin directory, plugin manifest, or plugin loader was discovered.

## Evidence

- `internal/state/state.go`
- `internal/catalog/catalog.go`
- `internal/dockpkg/package.go`
- `internal/values/values.go`
- `internal/render/render.go`
- `internal/envfile/envfile.go`
- `internal/cli/common.go`
- `internal/cli/install.go`
- `internal/cli/install_plan.go`
- `internal/runner/docker.go`
- `internal/oci/oci.go`
