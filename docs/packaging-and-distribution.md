# Packaging and Distribution

This guide covers package format, lockfiles, quality gates, dependency metadata, package archives, verification, and OCI distribution.

## Official catalog sources

Dockyard v1.7 adds catalog-aware package resolution for the official package catalog.

```bash
dockyard catalog list
dockyard catalog info postgres
dockyard install postgres
dockyard install my-db postgres
dockyard install my-db catalog://postgres:0.1.0
```

The default catalog registry is:

```text
ghcr.io/nandub/dockyard-packages
```

Set `DOCKYARD_CATALOG` to point the same commands at a private catalog:

```bash
export DOCKYARD_CATALOG=ghcr.io/my-org/my-packages
dockyard install redis
```

Explicit local paths, archives, and `oci://` references keep their existing behavior and do not use catalog resolution.

Automation can inspect catalog installs with clean JSON output:

```bash
dockyard install --dry-run redis --json
```

When `--json` is used for dry-run/install-plan output, Dockyard keeps OCI pull progress out of stdout.

## Package format

A package directory contains:

```text
Dockyard.yaml
values.yaml
values.schema.json
compose.yaml
README.md
SECURITY.md
LICENSE
```

`Dockyard.yaml` uses the stable v1.x manifest API:

```yaml
apiVersion: dockyard.dev/v1alpha1
```

`dockyard package` creates a `.dockyard.tgz` archive and includes generated metadata such as:

```text
SHA256SUMS
package.provenance.json
dockyard.lock        # when generated before packaging
```

Package archives reject common secret-like files such as `.env`, `*.pem`, `*.key`, `id_rsa`, and `id_ed25519`.

Archive verification can be bypassed for `push` and `pull` with `--skip-verify`. Use that only for controlled troubleshooting; normal publish and consume paths should keep verification enabled.

## Lockfiles

Create `dockyard.lock` for a specific render:

```bash
dockyard lock ../dockyard-work/team-dashboard   -f ../deploy-values/dashboard-prod.yaml
```

The lockfile records:

```text
package identity
values digest
rendered Compose digest
package file digests
images found in rendered Compose
digest-pinned image references when already present
declared package dependency references
```

Use `--require-lock` during install or upgrade to ensure the lockfile still matches the package, values, and rendered Compose output:

```bash
dockyard install dashboard-prod ../dockyard-artifacts/team-dashboard-0.1.0.dockyard.tgz   -f ../deploy-values/dashboard-prod.yaml   --require-lock
```

## Package dependencies

Packages may declare dependency metadata in `Dockyard.yaml`.

```yaml
dependencies:
  - name: postgres
    alias: db
    version: 0.1.0
    source: oci://ghcr.io/nandub/dockyard/postgres:0.1.0
    values:
      database: dashboard
      username: dashboard
      password: "${DASHBOARD_DB_PASSWORD}"
```

Dependency rules:

- `name` is required and must use the same safe identifier format as package names;
- `source` is required;
- `alias` is optional and can describe the dependency role, such as `db`;
- OCI dependency sources must include an explicit tag or digest;
- duplicate dependency names or aliases are rejected;
- inline `values:` are passed to the dependency package when `--with-dependencies` is used.

Inspect dependencies:

```bash
dockyard package deps ./examples/team-dashboard
dockyard package deps ./examples/team-dashboard --json
```

Preview the dependency-aware install plan:

```bash
dockyard install-plan team-dashboard ./examples/team-dashboard
dockyard install --dry-run team-dashboard ./examples/team-dashboard
dockyard install --dry-run team-dashboard ./examples/team-dashboard --json
```

Example plan:

```text
Install plan for release team-dashboard

1. dependency: postgres as db@0.1.0
   source: oci://ghcr.io/nandub/dockyard/postgres:0.1.0
   planned release: team-dashboard-db
   action: install
   automatic install: use `dockyard install --with-dependencies`

2. root package: team-dashboard@0.2.0
   source: ./examples/team-dashboard
   planned release: team-dashboard
   action: install

Read-only: no releases were installed, upgraded, uninstalled, or modified.
```

Install dependencies with explicit opt-in:

```bash
dockyard install --with-dependencies team-dashboard ./examples/team-dashboard
```

Dependency behavior is conservative:

- plain `dockyard install` installs only the root package;
- dependency releases install before the root package;
- release names are deterministic: `RELEASE-ALIAS` when an alias exists, otherwise `RELEASE-NAME`;
- existing deployed dependency releases are reused;
- failed or pending dependency releases block automatic dependency installation;
- dependency releases are not automatically upgraded or uninstalled;
- uninstalling a root release does not automatically remove its dependencies.

When installed with `--with-dependencies`, Dockyard records parent/child relationship metadata in `release.json`. `dockyard list` shows `deps=N` for root releases and `child-of=RELEASE` for dependency releases. `dockyard status` shows the exact parent or dependency references.

Dockyard blocks direct uninstall of a dependency release while an active parent release still depends on it. Uninstall the parent first, then the dependency. Use `dockyard uninstall --force DEPENDENCY_RELEASE` only for deliberate recovery operations.

## Reusable dependency example

The `examples/team-dashboard` package depends on `examples/postgres` through GHCR. Publish the postgres package before running the end-to-end dependency install demo:

```bash
dockyard package lint ./examples/postgres --strict
dockyard package test ./examples/postgres --strict
dockyard package ./examples/postgres   -o ../dockyard-artifacts/postgres-0.1.0.dockyard.tgz

dockyard push ../dockyard-artifacts/postgres-0.1.0.dockyard.tgz   oci://ghcr.io/nandub/dockyard/postgres:0.1.0
```

Then install the root package with dependencies:

```bash
dockyard install --with-dependencies team-dashboard ./examples/team-dashboard
```

## Package quality checks

Use `dockyard package lint` before creating or publishing a package.

```bash
dockyard package lint ./examples/nginx
dockyard package lint ./examples/nginx --strict
dockyard package lint ./examples/nginx --json
```

The quality checker is stricter than `dockyard compat`. It is intended for package authors and example maintainers.

It checks:

- `Dockyard.yaml` loads and uses a supported API version;
- dependency metadata is valid;
- recommended metadata files exist: `README.md`, `SECURITY.md`, and `LICENSE`;
- forbidden local artifacts are absent, including `.dockyard/`, `.git/`, `deploy-values/`, `.env`, private keys, and certificate key files;
- `values.yaml` loads successfully;
- `values.schema.json` validates the default values;
- public schema leaf values include `description`;
- secret-like schema fields use `x-dockyard-sensitive: true`;
- the default Compose render succeeds;
- the default render passes the configured Dockyard policy.

Use `--strict` when preparing packages for examples, internal catalogs, or OCI publishing. In strict mode, warnings fail. Use `--allow-advisory` when private/internal packages intentionally rely on repository-level licensing instead of a package-local `LICENSE`.

## Test a package

Use `dockyard package test` before publishing or sharing a package:

```bash
dockyard package test ../dockyard-work/team-dashboard   -f ../deploy-values/dashboard-prod.yaml
```

The default test pipeline is non-destructive. It:

```text
prepares the source directory, archive, or OCI package
runs package quality checks
validates values and schema
renders Compose with the selected values
runs Dockyard policy checks
runs docker compose config
```

For examples that are safe to run locally, add `--smoke`:

```bash
dockyard package test ../dockyard-work/example-app   -f ../deploy-values/local.yaml   --smoke
```

Smoke tests use a temporary Compose project name and do not write Dockyard release state. They run `docker compose up`, show container status with `docker compose ps --all`, and then run `docker compose down`.

Use `--strict` before publishing public examples:

```bash
dockyard package test ../dockyard-work/example-app   -f ../deploy-values/local.yaml   --strict
```

## Create and verify a package archive

Create an archive:

```bash
dockyard package ../dockyard-work/team-dashboard   --locked   -f ../deploy-values/dashboard-prod.yaml   -o ../dockyard-artifacts/team-dashboard-0.1.0.dockyard.tgz
```

Verify an archive:

```bash
dockyard verify ../dockyard-artifacts/team-dashboard-0.1.0.dockyard.tgz   -f ../deploy-values/dashboard-prod.yaml   --require-lock
```

Verification checks package structure, unsafe paths, forbidden files, integrity hashes, `Dockyard.yaml`, values validation, Compose rendering, and policy checks.

Install from an archive:

```bash
dockyard install dashboard-prod ../dockyard-artifacts/team-dashboard-0.1.0.dockyard.tgz   -f ../deploy-values/dashboard-prod.yaml   --require-lock
```

Archives are verified before extraction for install, upgrade, and diff.

## OCI registry support

Dockyard can push and pull `.dockyard.tgz` package archives using OCI registries through the `oras` CLI.

Dockyard publishes packages with:

```text
artifact type: application/vnd.dockyard.package.v1+gzip
archive layer: application/vnd.dockyard.package.archive.v1+gzip
```

This lets registries and tooling distinguish Dockyard packages from generic ORAS blobs.

Install and authenticate with `oras` before using these commands:

```bash
oras login ghcr.io
```

Dockyard delegates authentication to `oras`. Do not pass credentials directly to Dockyard commands.

Dockyard does not implement OCI package or catalog signature verification. Prefer immutable digest references or tightly controlled tags for high-trust environments.

Push a verified archive:

```bash
dockyard push ../dockyard-artifacts/team-dashboard-0.1.0.dockyard.tgz   oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.0
```

Pull an archive:

```bash
dockyard pull oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.0   -o ../dockyard-artifacts/team-dashboard-0.1.0.dockyard.tgz
```

Install directly from an OCI reference:

```bash
dockyard install dashboard-prod   oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.0   -f ../deploy-values/dashboard-prod.yaml   --require-lock
```

Upgrade or diff directly from OCI:

```bash
dockyard diff dashboard-prod   oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.1   -f ../deploy-values/dashboard-prod.yaml   --require-lock

dockyard upgrade dashboard-prod   oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.1   -f ../deploy-values/dashboard-prod.yaml   --require-lock
```

`dockyard doctor` reports whether `oras` is available. OCI references must include an explicit tag or digest.

## Publish official examples

Before publishing an example package:

```bash
dockyard lock ./examples/nginx
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
dockyard package ./examples/nginx --locked -o ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz
dockyard verify ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz --require-lock
```

Push to GHCR:

```bash
oras login ghcr.io
dockyard push ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz   oci://ghcr.io/nandub/dockyard/nginx:0.1.0
```

Test the published artifact:

```bash
dockyard package test oci://ghcr.io/nandub/dockyard/nginx:0.1.0 --smoke
```


## Publishing a catalog index

Dockyard catalog metadata is distributed as an OCI artifact. Publish a `catalog.yaml` file to a catalog reference such as:

```bash
oras push --artifact-type application/vnd.dockyard.catalog.v1+yaml ghcr.io/nandub/dockyard-packages/catalog:latest catalog.yaml:application/vnd.dockyard.catalog.index.v1+yaml
```

Operators can then run:

```bash
DOCKYARD_CATALOG=oci://ghcr.io/nandub/dockyard-packages/catalog:latest dockyard catalog list
```

Use the same pattern for private catalogs. Package artifacts remain separate OCI objects, for example `oci://ghcr.io/nandub/dockyard-packages/redis:0.1.0`.
