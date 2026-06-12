# Packaging and Distribution

This guide covers lockfiles, package archives, verification, and OCI distribution.

## Package format

A package directory contains:

```text
Dockyard.yaml
values.yaml
values.schema.json
compose.yaml
README.md
SECURITY.md
```

`dockyard package` creates a `.dockyard.tgz` archive and includes:

```text
SHA256SUMS
package.provenance.json
dockyard.lock        # when generated before packaging
```

Package archives reject common secret-like files such as `.env`, `*.pem`, `*.key`, `id_rsa`, and `id_ed25519`.

## Lockfiles

Create `dockyard.lock` for a specific render:

```bash
dockyard lock ../dockyard-work/team-dashboard \
  -f ../deploy-values/dashboard-prod.yaml
```

The lockfile records:

```text
package identity
values digest
rendered Compose digest
package file digests
images found in rendered Compose
digest-pinned image references when already present
```

Use `--require-lock` during install or upgrade to ensure the lockfile still matches the package, values, and rendered Compose output:

```bash
dockyard install dashboard-prod ../dockyard-artifacts/team-dashboard-0.1.0.dockyard.tgz \
  -f ../deploy-values/dashboard-prod.yaml \
  --require-lock
```



## Dependency install plans

Packages can declare dependency metadata in `Dockyard.yaml`. Operators can preview the dependency-aware install order with:

```bash
dockyard install-plan team-dashboard ./examples/team-dashboard
```

Example output:

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

Dependency release names are deterministic: `RELEASE-ALIAS` when an alias is set, otherwise `RELEASE-NAME`. Use `--json` for automation or review tooling.

To install dependencies, use the explicit opt-in flag:

```bash
dockyard install --with-dependencies team-dashboard ./examples/team-dashboard
```

Dependency packages are installed before the root package. Existing deployed dependency releases are reused and left unchanged. Dependencies are not automatically uninstalled if root installation fails or when the root release is uninstalled.

The team-dashboard example depends on the reusable postgres package. Publish it first:

```bash
dockyard package ./examples/postgres \
  -o ../dockyard-artifacts/postgres-0.1.0.dockyard.tgz

dockyard push ../dockyard-artifacts/postgres-0.1.0.dockyard.tgz \
  oci://ghcr.io/nandub/dockyard/postgres:0.1.0
```



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

Smoke tests use a temporary Compose project name and do not write Dockyard release state. They run `docker compose up`, show container status with `docker compose ps --all`, and then run `docker compose down`. Smoke tests require Docker and a reachable Docker daemon. If a smoke test fails before Compose starts, run `dockyard doctor` and verify Docker Desktop or your Docker daemon is running.

Use `--strict` before publishing public examples:

```bash
dockyard package test ../dockyard-work/example-app   -f ../deploy-values/local.yaml   --strict
```

## Create a package archive

```bash
dockyard package ../dockyard-work/team-dashboard \
  --locked \
  -f ../deploy-values/dashboard-prod.yaml \
  -o ../dockyard-artifacts/team-dashboard-0.1.0.dockyard.tgz
```

## Verify a package archive

```bash
dockyard verify ../dockyard-artifacts/team-dashboard-0.1.0.dockyard.tgz \
  -f ../deploy-values/dashboard-prod.yaml \
  --require-lock
```

Verification checks package structure, unsafe paths, forbidden files, integrity hashes, `Dockyard.yaml`, values validation, Compose rendering, and policy checks.

## Install from archives

```bash
dockyard install dashboard-prod ../dockyard-artifacts/team-dashboard-0.1.0.dockyard.tgz \
  -f ../deploy-values/dashboard-prod.yaml \
  --require-lock
```

Archives are verified before extraction for install, upgrade, and diff.



## Package dependencies

Dockyard packages may declare dependency metadata in `Dockyard.yaml`.

```yaml
dependencies:
  - name: postgres
    alias: db
    version: 0.1.0
    source: oci://ghcr.io/nandub/dockyard/postgres:0.1.0
    values:
      database: dashboard
      username: dashboard
```

Dependency support in v1.2 is intentionally metadata-only. Dockyard validates dependency declarations, includes dependency references in `dockyard.lock`, and exposes them through:

```bash
dockyard package deps ./examples/team-dashboard
dockyard package deps oci://ghcr.io/nandub/dockyard/team-dashboard:0.2.0 --json
```

Dockyard does not automatically install, upgrade, or uninstall dependencies yet. Use this metadata to document package requirements and prepare for future dependency orchestration.

Dependency rules:

- `name` is required and must use the same safe identifier format as package names.
- `source` is required.
- `alias` is optional and can describe the release or service role, such as `db`.
- OCI dependency sources must include an explicit tag or digest.
- Duplicate dependency names or aliases are rejected.


## OCI registry support

Dockyard can push and pull `.dockyard.tgz` package archives using OCI registries through the `oras` CLI.

Dockyard publishes packages with the OCI artifact type:

```text
application/vnd.dockyard.package.v1+gzip
```

and the package archive layer media type:

```text
application/vnd.dockyard.package.archive.v1+gzip
```

This lets registries and tooling distinguish Dockyard packages from generic ORAS blobs.

Install and authenticate with `oras` before using these commands:

```bash
oras login ghcr.io
```

Dockyard delegates authentication to `oras`. Do not pass credentials directly to Dockyard commands.

Push a verified archive:

```bash
dockyard push ../dockyard-artifacts/team-dashboard-0.1.0.dockyard.tgz \
  oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.0
```

Pull an archive:

```bash
dockyard pull oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.0 \
  -o ../dockyard-artifacts/team-dashboard-0.1.0.dockyard.tgz
```

Install directly from an OCI reference:

```bash
dockyard install dashboard-prod \
  oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.0 \
  -f ../deploy-values/dashboard-prod.yaml \
  --require-lock
```

Upgrade or diff directly from OCI:

```bash
dockyard diff dashboard-prod \
  oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.1 \
  -f ../deploy-values/dashboard-prod.yaml \
  --require-lock

dockyard upgrade dashboard-prod \
  oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.1 \
  -f ../deploy-values/dashboard-prod.yaml \
  --require-lock
```

`dockyard doctor` reports whether `oras` is available. OCI references must include an explicit tag or digest.

### Publish official examples

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
dockyard push ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz \
  oci://ghcr.io/nandub/dockyard/nginx:0.1.0
```

Test the published artifact:

```bash
dockyard package test oci://ghcr.io/nandub/dockyard/nginx:0.1.0 --smoke
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

- `Dockyard.yaml` loads and uses a supported API version.
- Recommended metadata files exist: `README.md`, `SECURITY.md`, and `LICENSE`.
- Forbidden local artifacts are absent, including `.dockyard/`, `.git/`, `deploy-values/`, `.env`, private keys, and certificate key files.
- `values.yaml` loads successfully.
- `values.schema.json` validates the default values.
- Public schema leaf values include `description`.
- Secret-like schema fields use `x-dockyard-sensitive: true`.
- The default Compose render succeeds.
- The default render passes the configured Dockyard policy.

Use `--strict` when preparing packages for examples, internal catalogs, or OCI publishing. In strict mode, missing README/SECURITY files and schema quality warnings become failures. In strict mode, warnings fail by default. Use `--allow-advisory` when private/internal packages intentionally rely on repository-level licensing instead of a package-local `LICENSE`.


## Dependency install planning and dry runs

Dependency support remains non-mutating. `dockyard install-plan RELEASE PACKAGE_SOURCE` and `dockyard install --dry-run RELEASE PACKAGE_SOURCE` share the same planner and must produce equivalent plans. This keeps package dependency behavior observable before Dockyard adds automatic dependency installation.

Use JSON output for automation:

```bash
dockyard install-plan team-dashboard ./examples/team-dashboard --json
dockyard install --dry-run team-dashboard ./examples/team-dashboard --json
```

## Installing packages with dependencies

Packages may declare dependencies in `Dockyard.yaml`. Dockyard does not install them by default; operators must opt in after reviewing the plan.

```sh
dockyard install-plan team-dashboard ./examples/team-dashboard
dockyard install --with-dependencies team-dashboard ./examples/team-dashboard
```

Dependency release names are deterministic. A dependency with `alias: db` under root release `team-dashboard` is planned as `team-dashboard-db`; without an alias, Dockyard uses the dependency name.

Dependency inline values are passed only to the dependency package. Root package `--values` and `--overlay` flags are intentionally not reused for dependencies because those files are package-specific.

Existing deployed dependency releases are reused. Dockyard does not upgrade or uninstall dependencies automatically.

Failed or pending dependency releases block automatic dependency installation; resolve them before re-running `dockyard install --with-dependencies`.


### Dependency release metadata

When a package is installed with `--with-dependencies`, Dockyard records each installed dependency release in the root release metadata and records the parent release on dependency releases. This metadata is operational only; package manifests remain the source of declared dependencies.
