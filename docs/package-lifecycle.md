# Package Lifecycle

This document summarizes the observed Dockyard package lifecycle.

## Lifecycle

1. Package source: a Dockyard package directory with `Dockyard.yaml` and related files.
2. Validate: lint, values/schema validation, policy checks, and Compose validation where requested.
3. Build package archive: create `.dockyard.tgz`.
4. Verify package archive: extract safely and validate checksums/provenance/manifest.
5. Push OCI package: use ORAS to push package artifacts.
6. Catalog: publish or load catalog metadata for package discovery and shorthand resolution.
7. Pull or resolve: use explicit OCI references, catalog references, or shorthand according to precedence.
8. Install: render Compose, validate/policy-check, write release state, and delegate to Docker Compose.
9. Upgrade: render and apply updated package state.
10. Rollback: restore a previous revision.
11. Uninstall/delete: delegate container lifecycle to Docker Compose and update release state.

## Boundaries

Dockyard manages package metadata, validation, rendering, state, archives, locks, and OCI orchestration. Docker Compose manages runtime containers, images, volumes, and networks.

## Evidence

- `internal/cli/package.go`
- `internal/archive/`
- `internal/lock/`
- `internal/oci/`
- `internal/catalog/`
- `internal/state/`
- `internal/runner/`

Historical detail: `.ai/onboarding/reports/plugin-model.md`, `.ai/onboarding/reports/runtime.md`, and `.ai/onboarding/reports/security.md`.
