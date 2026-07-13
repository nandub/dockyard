# Dockyard Plugin Model

This document records the extension and plugin model discovered in the repository.

## Summary

No first-class plugin system was found in the Dockyard source tree.

There are no discovered plugin manifests, plugin loader packages, dynamic module interfaces, embedded scripting hooks, or runtime execution hooks for package-provided code.

Dockyard's extension points are data/configuration based:

- Dockyard packages.
- Compose files.
- Values files.
- Values schemas.
- Compose overlays.
- Package dependency metadata.
- OCI package/catalog references.
- External Docker Compose behavior.

## Explicit Non-Plugin Boundaries

Repository guidance states:

- Do not execute arbitrary hooks or scripts from packages.
- Docker Compose remains the runtime source of truth.
- OCI credentials stay outside Dockyard.
- ORAS authentication is handled by external `oras` state.

Evidence:

- `AGENTS.md`
- `internal/runner/docker.go`
- `internal/oci/oci.go`
- `internal/catalog/catalog.go`

## Package Model

Dockyard packages are directory or archive inputs containing:

- `Dockyard.yaml`
- `compose.yaml` or configured Compose file
- `values.yaml`
- optional `values.schema.json`
- optional `dockyard.lock`
- package documentation files

The stable package manifest API version is:

- `dockyard.dev/v1alpha1`

Evidence:

- `internal/dockpkg/package.go`
- `internal/format/format.go`
- `examples/*/Dockyard.yaml`

## Dependency Model

Package dependencies are declared in `Dockyard.yaml`.

Example discovered:

```yaml
dependencies:
  - name: postgres
    alias: db
    source: oci://ghcr.io/nandub/dockyard/postgres:0.1.0
```

Dependency behavior discovered in docs and code:

- `dockyard install-plan` is read-only.
- Plain `dockyard install` installs only the root package.
- Dependency installation requires `dockyard install --with-dependencies`.
- Dependency release names are deterministic.
- Existing deployed dependency releases are reused.
- Uninstalled dependency releases can be reinstalled.
- Failed or pending dependency releases block automatic dependency installation.
- Dependencies are not automatically removed when the root release is uninstalled.

Evidence:

- `examples/team-dashboard/Dockyard.yaml`
- `internal/cli/install_plan.go`
- `internal/cli/install.go`
- `docs/command-reference.md`
- `AGENTS.md`

## Catalog Model

Catalog support is an extension mechanism for discovering package references, not a code plugin system.

Catalog source:

- Environment variable: `DOCKYARD_CATALOG`
- Default: `oci://ghcr.io/nandub/dockyard-packages/catalog:latest`

Supported catalog source forms:

- `oci://...`
- `file://...`
- local `.yaml` or `.yml` path
- registry prefix shorthand, normalized to `oci://.../catalog:latest`

Supported package source shorthand:

- `catalog://NAME[:VERSION]`
- bare package name, when found in the configured catalog

Catalog API version:

- `dockyard.dev/catalog/v1alpha1`

Evidence:

- `internal/catalog/catalog.go`
- `internal/cli/catalog.go`
- `README.md`
- `docs/command-reference.md`

## OCI Model

OCI support is implemented through the external `oras` CLI, not through a linked registry client library.

Package artifact metadata:

- Artifact type: `application/vnd.dockyard.package.v1+gzip`
- Archive layer media type: `application/vnd.dockyard.package.archive.v1+gzip`

Catalog metadata mentioned in docs:

- Artifact type: `application/vnd.dockyard.catalog.v1+yaml`
- Layer media type: `application/vnd.dockyard.catalog.index.v1+yaml`

Evidence:

- `internal/oci/oci.go`
- `internal/catalog/catalog.go`
- `docs/packaging-and-distribution.md`
- `README.md`

## Compose Extension Model

Dockyard packages use Docker Compose as the runtime extension surface.

Supported package-level Compose features discovered:

- Base Compose file through `compose.base`.
- Overlay map through `compose.overlays`.
- Values placeholders rendered into Compose files.
- `docker compose config` validation.

Evidence:

- `internal/dockpkg/package.go`
- `internal/render/render.go`
- `internal/runner/docker.go`
- `examples/*/Dockyard.yaml`
- `examples/*/compose.yaml`

## Values and Schema Model

Values are the primary package configuration extension point.

Discovered value files:

- `examples/*/values.yaml`

Discovered schema files:

- `examples/*/values.schema.json`

Schema validation is implemented with:

- `github.com/santhosh-tekuri/jsonschema/v6`

Evidence:

- `go.mod`
- `internal/values/values.go`
- `examples/*/values.schema.json`

## Policy and Quality Extension Model

Policy and quality behavior is built into Dockyard, not externally pluggable in the discovered code.

Relevant components:

- `internal/policy`
- `internal/quality`
- `internal/cli/secrets.go`
- `internal/cli/compat.go`

Commands:

- `dockyard policy list`
- `dockyard policy check`
- `dockyard secrets scan`
- `dockyard compat`
- `dockyard package lint`
- `dockyard package test`

Evidence:

- `internal/cli/policy.go`
- `internal/cli/secrets.go`
- `internal/cli/compat.go`
- `internal/cli/package.go`

## Absent Plugin Artifacts

No files found matching plugin-related implementation markers during repository inventory:

- No `plugins/` directory.
- No plugin manifest files discovered.
- No script files discovered for package hooks.
- No Dockerfile-based plugin builds discovered.
- No vendored plugin SDK or dynamic loading package discovered.

Absence evidence:

- Full tracked file inventory from `git ls-files`.
- `rg --files` scans for scripts, Dockerfiles, and Terraform files.
- Directory scan for `vendor`, `node_modules`, `.cache`, `dist`, `tmp`, and similar generated/cache directories.

## Open Verification Items

The following would require runtime checks or deeper design confirmation:

- Whether a future plugin model is planned outside this repository.
- Whether catalog packages are intended to be treated as the public extension boundary.
- Whether policy checks are expected to become externally configurable.
- Whether package dependency behavior is considered a plugin-like mechanism by project maintainers.
