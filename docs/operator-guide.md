# Operator Guide

This guide covers values files, environment files, overlays, release state, pruning, and dependency operations.

## Installing from the official catalog

For common services, operators can use the configured Dockyard catalog instead of copying long OCI references.

```bash
dockyard catalog list
dockyard catalog info redis
dockyard install redis
dockyard install my-cache redis
```

By default, catalog packages resolve under:

```text
ghcr.io/nandub/dockyard-packages
```

Use `DOCKYARD_CATALOG` to switch to an internal catalog while preserving the same commands:

```bash
export DOCKYARD_CATALOG=ghcr.io/my-org/internal-dockyard-packages
dockyard install redis
```

## Values files

Dockyard separates reusable package defaults from operator-owned deployment values.

```text
Package author owns:
  Dockyard.yaml
  compose.yaml
  values.yaml
  values.schema.json

Operator owns:
  ../deploy-values/prod.yaml
  ../deploy-values/staging.yaml
  ../deploy-values/local.yaml
```

Generate an operator-friendly values file:

```bash
dockyard values template ../dockyard-work/team-dashboard \
  -o ../deploy-values/dashboard-prod.yaml
```

Overwrite during local testing:

```bash
dockyard values template ../dockyard-work/team-dashboard \
  -o ../deploy-values/dashboard-prod.yaml \
  --force
```

Validate values:

```bash
dockyard values validate ../dockyard-work/team-dashboard \
  -f ../deploy-values/dashboard-prod.yaml
```

Print the schema:

```bash
dockyard values schema ../dockyard-work/team-dashboard
```

Schema descriptions become comments in generated values templates. Sensitive fields can be marked with `x-dockyard-sensitive: true`.

## Environment files

Dockyard can pass a private dotenv file to Docker Compose without mutating your current shell environment.

Use this when a values file intentionally contains environment references:

```yaml
database:
  password: "${DASHBOARD_DB_PASSWORD}"
```

Keep the real value in a private file:

```dotenv
DASHBOARD_DB_PASSWORD=use-a-long-random-password
```

Generate templates:

```bash
dockyard env template ../dockyard-work/team-dashboard \
  -o ../deploy-values/team-dashboard.env.example

dockyard env template ../dockyard-work/team-dashboard \
  --sensitive-only \
  -o ../deploy-values/team-dashboard.secrets.env.example
```

Check env files:

```bash
dockyard env check ../deploy-values/team-dashboard.env.example
```

Use the env file during Compose-facing commands:

```bash
dockyard config ../dockyard-work/team-dashboard \
  -f ../deploy-values/dashboard-prod.yaml \
  --env-file ../deploy-values/dashboard-prod.env

dockyard install dashboard-prod ../dockyard-work/team-dashboard \
  -f ../deploy-values/dashboard-prod.yaml \
  --env-file ../deploy-values/dashboard-prod.env
```

Dockyard records the env-file path in release metadata, but does not store secret values in `release.json`.

## Values versus overlays

Use values files for environment-specific settings. Use overlays only when the Compose structure itself needs to change.

Use values for:

```text
image tags
ports
hostnames
database names
feature flags
resource values
secret references
```

Use overlays for:

```text
adding or removing services
different networks
reverse proxy labels
production logging
volume layout changes
structural security changes
```

Beginner path without overlays:

```bash
dockyard install dashboard-prod ../dockyard-work/team-dashboard \
  -f ../deploy-values/dashboard-prod.yaml \
  --require-lock
```

Advanced path with an overlay:

```bash
dockyard install dashboard-prod ../dockyard-work/team-dashboard \
  -f ../deploy-values/dashboard-prod.yaml \
  --overlay prod \
  --require-lock
```

## Release state

Dockyard stores release state in this precedence order:

1. `--home`
2. `DOCKYARD_HOME`
3. `~/.dockyard`

Example:

```bash
export DOCKYARD_HOME=/var/lib/dockyard
dockyard install myapp ../dockyard-work/example-app
```

Release layout:

```text
~/.dockyard/
  releases/
    myapp/
      current
      revisions/
        1/
          Dockyard.yaml
          values.yaml
          dockyard.lock
          compose.rendered.yaml
          release.json
```

Useful commands:

```bash
dockyard status myapp
dockyard status myapp --compose-ps
dockyard status myapp --compose-ps --all
dockyard inspect myapp
dockyard list                         # active releases by default
dockyard list --all                   # include uninstalled history
dockyard list --status uninstalled    # filter by release state
dockyard uninstall myapp --dry-run
```

`dockyard list` includes a `RELATION` column. Standalone releases show `-`, root releases with dependencies show `deps=N`, and dependency releases show `child-of=RELEASE`.

## Release pruning

Dockyard keeps release revision history under `DOCKYARD_HOME`.

Dry-run all releases:

```bash
dockyard prune --dry-run
```

Keep the newest five revisions per release:

```bash
dockyard prune --keep 5
```

Prune only one release:

```bash
dockyard prune --release myapp --keep 3
```


## Dependency operations

Dependency installs are explicit. Preview the plan first:

```bash
dockyard install-plan team-dashboard ./examples/team-dashboard
dockyard install --dry-run team-dashboard ./examples/team-dashboard
```

Install dependencies before the root package only when you opt in:

```bash
dockyard install --with-dependencies team-dashboard ./examples/team-dashboard
```

Dockyard records relationship metadata for releases installed with `--with-dependencies`. Operators can use `dockyard list` to identify roots and dependency releases, and `dockyard status RELEASE` to inspect exact parent/dependency links.

Example list output:

```text
NAME               STATUS    REVISION  PACKAGE               RELATION
team-dashboard     deployed  1         team-dashboard@0.2.0  deps=1
team-dashboard-db  deployed  1         postgres@0.1.0        child-of=team-dashboard
```

Dependency releases are protected from accidental direct removal. If an active release still references a dependency release, `dockyard uninstall DEPENDENCY_RELEASE` fails with a message listing the active dependent release.

Correct uninstall order:

```powershell
dockyard uninstall team-dashboard
dockyard uninstall team-dashboard-db
```

Use `--force` only for manual recovery when you intentionally want to remove a dependency release while a parent still references it.

```powershell
dockyard uninstall team-dashboard-db --force
```

Dockyard does not automatically remove dependency releases when a root release is uninstalled.
