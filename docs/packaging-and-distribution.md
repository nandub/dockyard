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

## OCI registry support

Dockyard can push and pull `.dockyard.tgz` package archives using OCI registries through the `oras` CLI.

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

