# postgres example

Reusable PostgreSQL dependency package used by `examples/team-dashboard`.

This example is intentionally small and suitable for local dependency-installation tests. It uses a named Docker volume for database storage and a PostgreSQL healthcheck.

## Validate

```powershell
dockyard package lint ./examples/postgres --strict
dockyard package test ./examples/postgres --strict
```

## Package and publish to GHCR

`dockyard package` takes an archive path with `-o` / `--output`.

```powershell
dockyard package ./examples/postgres `
  -o ../dockyard-artifacts/postgres-0.1.0.dockyard.tgz

dockyard push ../dockyard-artifacts/postgres-0.1.0.dockyard.tgz `
  oci://ghcr.io/nandub/dockyard/postgres:0.1.0
```

## Use as a dependency

`examples/team-dashboard` declares this package as:

```yaml
dependencies:
  - name: postgres
    alias: db
    version: 0.1.0
    source: oci://ghcr.io/nandub/dockyard/postgres:0.1.0
```

Then run:

```powershell
dockyard install --with-dependencies team-dashboard ./examples/team-dashboard
```

## Security notes

The default password value uses a Compose environment default so local tests can run without a committed secret. For real deployments, pass a secret through an environment file or secret manager-backed environment variable and do not commit production values.
