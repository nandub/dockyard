# OCI Model

Dockyard supports OCI package and catalog operations through the embedded ORAS Go client.

Evidence:

- `internal/oci/`
- `internal/catalog/`
- `internal/archive/`
- `internal/lock/`
- `docs/packaging-and-distribution.md`

## References

Observed source forms include local paths, archives, explicit `oci://` references, catalog references, and shorthand resolution through catalog metadata.

Existing local paths, archives, and explicit `oci://` package references should keep precedence over catalog shorthand.

## ORAS Boundary

Dockyard links `oras.land/oras-go/v2` for OCI package push, package pull, and catalog metadata pull. Package archives are pushed as named OCI layers so pulls restore the original archive filename.

Dockyard does not store registry credentials. Registry authentication uses Docker-compatible credential configuration when available, including Docker config files and configured credential helpers. Anonymous registry access is used when no matching credentials are configured.

TLS, proxy, retry, and registry protocol behavior are handled by Go's HTTP stack and the ORAS Go client rather than by an external `oras` executable.

## Package Artifacts

Package archives are `.dockyard.tgz` files. OCI package work intersects with archive verification, lockfiles, provenance metadata, catalog metadata, and release/package commands.

## Security Notes

No OCI signature verification is implemented in the current source. Do not claim signature validation unless source proves it.

`dockyard push` verifies local archives before publishing unless `--skip-verify` is used. `dockyard pull` verifies pulled archives unless `--skip-verify` is used.

Catalog metadata can be loaded from a configured OCI reference, a local YAML path, a `file://` path, or a short package name resolved through the configured catalog. The catalog cache is stored below the operating-system user home. Treat catalog configuration as a trust decision.

Historical detail: `.ai/onboarding/reports/security.md` and `.ai/onboarding/reports/dependency-graph.md`.
