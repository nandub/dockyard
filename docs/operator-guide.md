# Operator Guide

This guide covers values files, environment files, overlays, release state, and pruning.

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
dockyard list          # active releases by default
dockyard list --all    # include uninstalled history
dockyard uninstall myapp --dry-run
```

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
