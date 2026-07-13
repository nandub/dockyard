# Environment Model

This document records environment variables, subprocess environments, and environment-related precedence.

## Environment Variables Read by Dockyard

### DOCKYARD_HOME

Purpose:

- Sets Dockyard state home when `--home` is not provided.

Precedence:

1. `--home`
2. `DOCKYARD_HOME`
3. `~/.dockyard`

Implementation:

- `internal/state/state.go`

### DOCKYARD_CATALOG

Purpose:

- Sets the catalog metadata source for catalog commands and package shorthand resolution.

Default:

```text
oci://ghcr.io/nandub/dockyard-packages/catalog:latest
```

Accepted forms:

- `oci://...`
- `file://...`
- Local `.yaml` or `.yml` path.
- Registry prefix shorthand.

Registry prefix example:

```text
ghcr.io/my-org/my-packages
```

Normalizes to:

```text
oci://ghcr.io/my-org/my-packages/catalog:latest
```

Implementation:

- `internal/catalog/catalog.go`

## Environment Variables Passed Through to Subprocesses

Docker Compose subprocesses inherit the current process environment by default.

When an env file is supplied:

1. Dockyard parses the env file.
2. Dockyard converts entries to `KEY=VALUE`.
3. Dockyard appends those entries to `os.Environ()` for the subprocess.
4. Dockyard does not mutate the current process environment.

Implementation:

- `internal/envfile/envfile.go`
- `internal/runner/docker.go`
- `internal/cli/common.go`

## Env File Model

Supported syntax:

- Blank lines.
- Comments.
- `KEY=VALUE`.
- Optional `export KEY=VALUE`.
- Single-quoted values.
- Double-quoted values.

Validation:

- Invalid lines fail parsing.
- Invalid environment variable names fail parsing.
- Duplicate keys fail parsing.
- Unterminated quotes fail parsing.

Use cases:

- `--env-file` on lifecycle and Compose-validation commands.
- `dockyard env template`.
- `dockyard env check`.

Security behavior:

- Real production `.env` files should stay private and out of source.
- Env-file values are not stored in release state.
- Release metadata records only the env-file path.

## External Tool Environment

### Docker and Docker Compose

Dockyard runs:

- `docker --version`
- `docker compose version`
- `docker info`
- `docker compose config`
- `docker compose up -d`
- `docker compose down`
- `docker compose ps`

Subprocess environment:

- Current process environment.
- Optional parsed env-file entries.

Dockyard does not manage Docker daemon configuration. Docker CLI, Docker Compose, credentials, contexts, and daemon reachability are external environment concerns.

### ORAS

Dockyard runs:

- `oras pull`
- `oras push`

Subprocess environment:

- ORAS inherits the process environment.
- Registry authentication is handled by ORAS outside Dockyard.

Dockyard checks whether `oras` exists in `PATH` before OCI operations.

## PATH Requirements

The following executables are discovered through `exec.LookPath` or executed directly:

- `docker`
- `oras`

`docker compose` is checked by running:

```bash
docker compose version
```

## Runtime Home and Cache Paths

Dockyard home default:

```text
~/.dockyard
```

Release state:

```text
~/.dockyard/releases/<release>/
```

Catalog cache:

```text
~/.dockyard/cache/catalogs/
```

Important detail:

- The catalog cache path uses the OS user home directly, not the `--home` / `DOCKYARD_HOME` release-state home.

Implementation:

- Release state: `internal/state/state.go`
- Catalog cache: `internal/catalog/catalog.go`

## Temporary Directories

Temporary directories are created with `os.MkdirTemp`.

Observed prefixes:

- `dockyard-oci-*`
- `dockyard-src-*`
- `dockyard-dependency-values-*`
- `dockyard-catalog-*`
- `dockyard-verify-*`
- `dockyard-config-*`
- `dockyard-render-*`
- `dockyard-package-test-*`
- `dockyard-pull-*`
- `dockyard-name-*`

Temporary directories are generally cleaned with deferred `os.RemoveAll` cleanup.

## Output and Logging Environment

No environment variable for log level or output format was found.

Output format is controlled by command flags where supported:

- `--json`
- `--dry-run`
- `--explain`

Subprocess output behavior:

- Docker visible operations use stdout and stderr.
- Docker Compose config validation discards stdout.
- ORAS normal operations use stdout and stderr.
- ORAS quiet mode discards stdout and stderr.

## Environment Precedence Summary

Dockyard home:

```text
--home > DOCKYARD_HOME > ~/.dockyard
```

Catalog source:

```text
DOCKYARD_CATALOG > oci://ghcr.io/nandub/dockyard-packages/catalog:latest
```

Values:

```text
package values.yaml < --values override
```

Docker subprocess environment:

```text
os.Environ() + --env-file entries
```

Package source resolution:

```text
local path / archive / explicit oci:// > catalog:// or bare catalog shorthand
```

## Sensitive Environment Handling

Sensitive key detection appears in render diagnostics and env-file helpers.

Sensitive markers include words such as:

- `password`
- `passwd`
- `secret`
- `token`
- `api_key`
- `apikey`
- `private_key`
- `credential`
- `key`

Behavior:

- Render diagnostics mask sensitive values.
- Env template generation leaves sensitive values empty with comments.
- Env check flags populated secret-like values.
- Release metadata stores env-file path, not secret values.

## Plugin Environment

No plugin environment variables were found.

No plugin loading path was found.

## Evidence

- `internal/state/state.go`
- `internal/catalog/catalog.go`
- `internal/envfile/envfile.go`
- `internal/runner/docker.go`
- `internal/oci/oci.go`
- `internal/cli/common.go`
- `internal/cli/doctor.go`
- `internal/render/render.go`
