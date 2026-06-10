# Recommended Project Layout

Dockyard separates reusable packages from operator-owned deployment values.

A simple repository layout:

```text
my-deployment/
  packages/
    team-dashboard/
      Dockyard.yaml
      values.yaml
      values.schema.json
      compose.yaml
      README.md
      SECURITY.md

  deploy-values/
    local.yaml
    staging.yaml
    prod.yaml

  artifacts/
    .gitkeep
```

## Package author owns

```text
packages/team-dashboard/
  Dockyard.yaml
  values.yaml
  values.schema.json
  compose.yaml
```

These files describe the reusable application package.

## Operator owns

```text
deploy-values/prod.yaml
deploy-values/staging.yaml
deploy-values/local.yaml
```

These files describe one environment. They may contain hostnames, exposed ports, image tags, and secret references.

Do not commit production secrets unless your repository and secret-management policy explicitly allow it.

## Example

Generate an operator values file:

```bash
dockyard values template ./packages/team-dashboard \
  -o ./deploy-values/prod.yaml
```

Validate it:

```bash
dockyard values validate ./packages/team-dashboard \
  -f ./deploy-values/prod.yaml
```

Install with it:

```bash
dockyard install dashboard-prod ./packages/team-dashboard \
  -f ./deploy-values/prod.yaml \
  --require-lock
```


## Environment files

For applications that need secrets, prefer committing only examples:

```text
deploy-values/
  prod.yaml                 # private if it contains secrets
  prod.example.yaml         # safe example values
  prod.env.example          # safe example environment variables
```

Generate examples with:

```bash
dockyard values template ./team-dashboard -o deploy-values/prod.example.yaml
dockyard env template ./team-dashboard --sensitive-only -o deploy-values/prod.env.example
```
