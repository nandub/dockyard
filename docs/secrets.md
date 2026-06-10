# Secret Handling

Dockyard does not try to be a secret manager.

The recommended model is:

```text
Package author owns:
  Dockyard.yaml
  compose.yaml
  values.yaml
  values.schema.json

Operator owns:
  deploy-values/prod.yaml
  deploy-values/staging.yaml
  deploy-values/local.yaml
```

Keep real production secrets out of reusable packages and package archives.

## Generate operator values

Generate an environment-specific values file:

```bash
dockyard values template ./team-dashboard -o ./deploy-values/dashboard-prod.yaml
```

Sensitive fields are generated as empty values when the schema marks them as sensitive or their names look secret-like:

```yaml
database:
  # Password for the PostgreSQL application user.
  # Sensitive value. Keep this file private and do not commit production secrets.
  password: ""
```

## Scan package defaults

Scan package defaults for populated secret-like values:

```bash
dockyard secrets scan ./team-dashboard
```

Fail CI if populated secret-like defaults are found:

```bash
dockyard secrets scan ./team-dashboard --strict
```

Scan an operator values file before committing by mistake:

```bash
dockyard secrets scan ./team-dashboard -f ./deploy-values/dashboard-prod.yaml --strict
```

A finding does not prove that a value is a real secret. It means the key looks sensitive and has a populated value.

## Production recommendation

For production deployments, prefer one of these patterns:

- generate `deploy-values/prod.yaml` at deploy time from a secret manager
- keep private values outside the package repository
- use platform-native Docker Compose secret support where appropriate
- never include `.env`, private keys, certificates, or production passwords in Dockyard package archives
