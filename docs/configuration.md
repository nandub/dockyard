# Configuration

Observed configuration sources include CLI flags, environment variables, package files, values files, schemas, lockfiles, catalog metadata, Docker configuration, and Docker-compatible registry credential configuration.

## Dockyard Home

Observed precedence:

1. `--home` flag.
2. `DOCKYARD_HOME`.
3. `~/.dockyard`.

## Catalog

`DOCKYARD_CATALOG` points to a catalog metadata source. Historical onboarding observed a default OCI catalog reference and registry-prefix compatibility; verify current source before changing catalog behavior.

## Package Inputs

Package behavior is driven by `Dockyard.yaml`, values files, optional schemas, Compose templates, overlays, lockfiles, archives, and OCI/catalog references.

## Environment

Env files can be passed to Docker Compose subprocesses. Release metadata should not store secret values.

Historical detail: `.ai/onboarding/reports/configuration.md` and `.ai/onboarding/reports/environment.md`.
