# OCI Model

Dockyard supports OCI package and catalog operations through the external `oras` CLI.

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

Dockyard shells out to `oras` for OCI push and pull. OCI authentication, credential storage, TLS behavior, and registry-specific auth are delegated to ORAS and the user environment.

## Package Artifacts

Package archives are `.dockyard.tgz` files. OCI package work intersects with archive verification, lockfiles, provenance metadata, catalog metadata, and release/package commands.

## Security Notes

No OCI signature verification was observed in historical onboarding. Do not claim signature validation unless source proves it.

Historical detail: `.ai/onboarding/reports/security.md` and `.ai/onboarding/reports/dependency-graph.md`.
