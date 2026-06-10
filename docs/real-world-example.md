# Real-World Dockyard Example: Team Dashboard with PostgreSQL

This guide shows a realistic Dockyard workflow for packaging, validating, locking, installing, upgrading, rolling back, verifying, distributing, and uninstalling a Docker Compose application.

The example application is an internal dashboard backed by PostgreSQL.

The main path intentionally does **not** use a Compose overlay. It uses a production values file only. See [Compose Overlays](overlays.md) for the advanced case where the Compose structure itself needs to change.

## Example names

| Item | Name |
|---|---|
| Dockyard package | `team-dashboard` |
| Installed release | `dashboard-prod` |
| Operator values file | `deploy-values/dashboard-prod.yaml` |
| Package archive | `team-dashboard-0.1.0.dockyard.tgz` |

## 1. Build Dockyard

From the repository root:

```bash
go mod tidy
go test ./...
go build -o bin/dockyard ./cmd/dockyard
```

Check the binary:

```bash
./bin/dockyard --help
```

Optional: choose a Dockyard state directory.

```bash
export DOCKYARD_HOME="$HOME/.dockyard"
```

Run the environment check:

```bash
./bin/dockyard doctor
```

## 2. Create a package

```bash
./bin/dockyard init ./team-dashboard
```

Create `Dockyard.yaml`:

```bash
cat > ./team-dashboard/Dockyard.yaml <<'EOF'
apiVersion: dockyard.dev/v1alpha1
name: team-dashboard
description: Internal dashboard with PostgreSQL
version: 0.1.0
appVersion: "1.0.0"
type: application

compose:
  base: compose.yaml
  overlays: {}

security:
  requireNonRoot: false
  requireHealthchecks: true
  disallowPrivileged: true
  disallowHostNetwork: true
  disallowDockerSocketMount: true
  disallowLatestTag: true
EOF
```

`requireNonRoot` is `false` in this example because database images often manage their own internal user switching. High-risk options such as privileged containers, host networking, Docker socket mounts, and `latest` tags are still blocked or reported.

## 3. Add values and schema

Create `values.yaml`:

```bash
cat > ./team-dashboard/values.yaml <<'EOF'
app:
  image: ghcr.io/example/team-dashboard
  tag: "1.0.0"
  port: 8080

database:
  image: postgres
  tag: "16.4"
  name: dashboard
  user: dashboard
  password: change-me-in-prod
  port: 5432
EOF
```

Create `values.schema.json`:

```bash
cat > ./team-dashboard/values.schema.json <<'EOF'
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["app", "database"],
  "properties": {
    "app": {
      "type": "object",
      "required": ["image", "tag", "port"],
      "properties": {
        "image": {
          "type": "string",
          "minLength": 1,
          "description": "Application image repository."
        },
        "tag": {
          "type": "string",
          "minLength": 1,
          "description": "Application image tag."
        },
        "port": {
          "type": "integer",
          "minimum": 1,
          "maximum": 65535,
          "description": "Host port exposed by the application."
        }
      },
      "additionalProperties": false
    },
    "database": {
      "type": "object",
      "required": ["image", "tag", "name", "user", "password", "port"],
      "properties": {
        "image": {
          "type": "string",
          "minLength": 1,
          "description": "PostgreSQL image repository."
        },
        "tag": {
          "type": "string",
          "minLength": 1,
          "description": "PostgreSQL image tag."
        },
        "name": {
          "type": "string",
          "minLength": 1,
          "description": "Application database name."
        },
        "user": {
          "type": "string",
          "minLength": 1,
          "description": "Application database user."
        },
        "password": {
          "type": "string",
          "minLength": 12,
          "description": "Application database password.",
          "x-dockyard-sensitive": true
        },
        "port": {
          "type": "integer",
          "minimum": 1,
          "maximum": 65535,
          "description": "PostgreSQL internal port."
        }
      },
      "additionalProperties": false
    }
  },
  "additionalProperties": false
}
EOF
```

## 4. Generate an operator values file

Production values belong outside the reusable package.

```bash
mkdir -p ./deploy-values

./bin/dockyard values template ./team-dashboard \
  -o ./deploy-values/dashboard-prod.yaml
```

The generated file is meant for the operator to edit. It uses comments from `values.schema.json` and masks sensitive fields.

Example output:

```yaml
app:
  # Application image repository.
  image: "ghcr.io/example/team-dashboard"
  # Host port exposed by the application.
  port: 8080
  # Application image tag.
  tag: "1.0.0"

database:
  # PostgreSQL image repository.
  image: "postgres"
  # Application database name.
  name: "dashboard"
  # Application database password.
  # Sensitive value. Keep this file private and do not commit production secrets.
  password: ""
  # PostgreSQL internal port.
  port: 5432
  # PostgreSQL image tag.
  tag: "16.4"
  # Application database user.
  user: "dashboard"
```

Edit the generated file and set a real password:

```bash
$EDITOR ./deploy-values/dashboard-prod.yaml
```

Validate it:

```bash
./bin/dockyard values validate ./team-dashboard \
  -f ./deploy-values/dashboard-prod.yaml
```

## 5. Add the Compose template

Create `compose.yaml`:

```bash
cat > ./team-dashboard/compose.yaml <<'EOF'
services:
  app:
    image: "${app.image}:${app.tag}"
    restart: unless-stopped
    depends_on:
      db:
        condition: service_healthy
    environment:
      DATABASE_URL: "postgres://${database.user}:${database.password}@db:5432/${database.name}?sslmode=disable"
    ports:
      - "${app.port}:8080"
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 5
    security_opt:
      - no-new-privileges:true
    cap_drop:
      - ALL

  db:
    image: "${database.image}:${database.tag}"
    restart: unless-stopped
    environment:
      POSTGRES_DB: "${database.name}"
      POSTGRES_USER: "${database.user}"
      POSTGRES_PASSWORD: "${database.password}"
    volumes:
      - postgres-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${database.user} -d ${database.name}"]
      interval: 10s
      timeout: 5s
      retries: 5
    security_opt:
      - no-new-privileges:true

volumes:
  postgres-data:
EOF
```

## 6. Lint and render

```bash
./bin/dockyard lint ./team-dashboard \
  -f ./deploy-values/dashboard-prod.yaml
```

Scan package defaults for accidental secrets:

```bash
dockyard secrets scan ./team-dashboard --strict
```

Run explicit policy checks:

```bash
dockyard policy check ./team-dashboard \
  -f ./deploy-values/dashboard-prod.yaml
```

Render and explain placeholders:

```bash
./bin/dockyard render ./team-dashboard \
  -f ./deploy-values/dashboard-prod.yaml \
  --explain
```

Sensitive values such as passwords are masked in diagnostics.

## 7. Validate Compose compatibility

Before installing, ask Docker Compose to validate the final rendered file:

```bash
./bin/dockyard config ./team-dashboard \
  -f ./deploy-values/dashboard-prod.yaml
```

This command renders the Dockyard package to a temporary Compose file and runs:

```bash
docker compose config
```

It does not create a release and does not start containers.

You can use the same check for a packaged archive:

```bash
./bin/dockyard config team-dashboard-0.1.0.dockyard.tgz \
  -f ./deploy-values/dashboard-prod.yaml \
  --require-lock
```

## 8. Create a lockfile

```bash
./bin/dockyard lock ./team-dashboard \
  -f ./deploy-values/dashboard-prod.yaml
```

This creates:

```text
team-dashboard/dockyard.lock
```

The lockfile records values digest, rendered Compose digest, package file digests, service image references, and image digests when image references are already pinned with `@sha256:...`.

Dockyard v0.6.2 does not resolve image tags to registry digests. It only records what appears in rendered Compose.

## 9. Package and verify

Create a locked package archive:

```bash
./bin/dockyard package ./team-dashboard \
  --locked \
  -f ./deploy-values/dashboard-prod.yaml \
  -o team-dashboard-0.1.0.dockyard.tgz
```

Verify it:

```bash
./bin/dockyard verify team-dashboard-0.1.0.dockyard.tgz \
  -f ./deploy-values/dashboard-prod.yaml \
  --require-lock
```

The archive includes:

```text
Dockyard.yaml
values.yaml
values.schema.json
compose.yaml
dockyard.lock
SHA256SUMS
package.provenance.json
```

The archive should not include `.env`, private keys, `.git/`, `.dockyard/`, `node_modules/`, or `vendor/`.

## 10. Install from the archive

Dry-run first:

```bash
./bin/dockyard install dashboard-prod team-dashboard-0.1.0.dockyard.tgz \
  -f ./deploy-values/dashboard-prod.yaml \
  --require-lock \
  --dry-run
```

Install:

```bash
./bin/dockyard install dashboard-prod team-dashboard-0.1.0.dockyard.tgz \
  -f ./deploy-values/dashboard-prod.yaml \
  --require-lock
```

Dockyard verifies the archive, extracts it to a temporary directory, validates the lockfile, renders Compose, writes release state, and runs:

```bash
docker compose -p dashboard-prod -f <rendered-compose-file> up -d
```

Release state is stored under:

```text
$DOCKYARD_HOME/releases/dashboard-prod/
```

## 11. Status, inspect, and list

```bash
./bin/dockyard status dashboard-prod
./bin/dockyard status dashboard-prod --compose-ps
./bin/dockyard inspect dashboard-prod
./bin/dockyard list
```

## 12. Upgrade

Change the app tag to `1.1.0` in a new values file:

```bash
cp ./deploy-values/dashboard-prod.yaml ./deploy-values/dashboard-prod-v1.1.0.yaml
```

Edit:

```yaml
app:
  tag: "1.1.0"
```

Regenerate the lockfile for the new render:

```bash
./bin/dockyard lock ./team-dashboard \
  -f ./deploy-values/dashboard-prod-v1.1.0.yaml
```

Package the new render:

```bash
./bin/dockyard package ./team-dashboard \
  --locked \
  -f ./deploy-values/dashboard-prod-v1.1.0.yaml \
  -o team-dashboard-0.1.0-v1.1.0-render.dockyard.tgz
```

Preview:

```bash
./bin/dockyard diff dashboard-prod team-dashboard-0.1.0-v1.1.0-render.dockyard.tgz \
  -f ./deploy-values/dashboard-prod-v1.1.0.yaml \
  --require-lock
```

Upgrade:

```bash
./bin/dockyard upgrade dashboard-prod team-dashboard-0.1.0-v1.1.0-render.dockyard.tgz \
  -f ./deploy-values/dashboard-prod-v1.1.0.yaml \
  --require-lock
```

## 13. Roll back

```bash
./bin/dockyard rollback dashboard-prod 1
```

## 14. Uninstall

Dry-run:

```bash
./bin/dockyard uninstall dashboard-prod --dry-run
```

Uninstall while preserving named volumes:

```bash
./bin/dockyard uninstall dashboard-prod
```

Remove named volumes too:

```bash
./bin/dockyard uninstall dashboard-prod --volumes
```

Remove Dockyard metadata after uninstall:

```bash
./bin/dockyard uninstall dashboard-prod --purge
```

## 15. Push and install from an OCI registry

Dockyard v0.6.2 can distribute the same verified `.dockyard.tgz` archive through OCI registries.

This MVP delegates registry authentication and transport to the `oras` CLI. For example, with GitHub Container Registry:

```bash
oras login ghcr.io
```

Push the package:

```bash
./bin/dockyard push team-dashboard-0.1.0.dockyard.tgz \
  oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.0
```

Pull the package somewhere else:

```bash
./bin/dockyard pull oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.0 \
  -o team-dashboard-0.1.0.dockyard.tgz
```

Install directly from the OCI reference:

```bash
./bin/dockyard install dashboard-prod \
  oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.0 \
  -f ./deploy-values/dashboard-prod.yaml \
  --require-lock
```

Preview and upgrade to a newer OCI package:

```bash
./bin/dockyard diff dashboard-prod \
  oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.1 \
  -f ./deploy-values/dashboard-prod.yaml \
  --require-lock

./bin/dockyard upgrade dashboard-prod \
  oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.1 \
  -f ./deploy-values/dashboard-prod.yaml \
  --require-lock
```

Release metadata records the package source as `oci`, so `dockyard inspect dashboard-prod` shows where the deployed revision came from.

## 16. Advanced: use a Compose overlay when values are not enough

Most deployments should start without overlays. Use an overlay when the Compose structure changes.

For example, add this to `Dockyard.yaml`:

```yaml
compose:
  base: compose.yaml
  overlays:
    prod: compose.prod.yaml
```

Create `compose.prod.yaml`:

```bash
cat > ./team-dashboard/compose.prod.yaml <<'EOF'
services:
  app:
    restart: always
    logging:
      options:
        max-size: "10m"
        max-file: "5"

  db:
    restart: always
EOF
```

Then pass `--overlay prod` explicitly:

```bash
./bin/dockyard install dashboard-prod team-dashboard-0.1.0.dockyard.tgz \
  -f ./deploy-values/dashboard-prod.yaml \
  --overlay prod \
  --require-lock
```

`-f ./deploy-values/dashboard-prod.yaml` chooses production values. `--overlay prod` applies the structural production Compose override.

## Operator workflow

For a normal install:

```bash
./bin/dockyard doctor
./bin/dockyard verify team-dashboard-0.1.0.dockyard.tgz -f ./deploy-values/dashboard-prod.yaml --require-lock
./bin/dockyard install dashboard-prod team-dashboard-0.1.0.dockyard.tgz -f ./deploy-values/dashboard-prod.yaml --require-lock
./bin/dockyard status dashboard-prod --compose-ps
```

For a normal upgrade:

```bash
./bin/dockyard verify team-dashboard-0.1.0-v1.1.0-render.dockyard.tgz -f ./deploy-values/dashboard-prod-v1.1.0.yaml --require-lock
./bin/dockyard diff dashboard-prod team-dashboard-0.1.0-v1.1.0-render.dockyard.tgz -f ./deploy-values/dashboard-prod-v1.1.0.yaml --require-lock
./bin/dockyard upgrade dashboard-prod team-dashboard-0.1.0-v1.1.0-render.dockyard.tgz -f ./deploy-values/dashboard-prod-v1.1.0.yaml --require-lock
./bin/dockyard status dashboard-prod --compose-ps
```

## Mental model

Dockyard does not replace Docker Compose.

Dockyard adds the missing package-manager layer:

```text
Dockyard package
  ↓
schema validation
  ↓
template rendering
  ↓
lockfile verification
  ↓
security linting
  ↓
package/archive verification
  ↓
release revision storage
  ↓
docker compose up/down
```

Values files change settings. Overlays change Compose structure. OCI stores and distributes the same verified package archive.


## Windows note

Windows PowerShell users can test HTTP endpoints with `curl.exe http://localhost:8080` or `Invoke-WebRequest http://localhost:8080 -UseBasicParsing`.


## Environment-variable secrets with v0.8

For production secrets, prefer environment references in the deployment values file instead of literal secrets.

Example `deploy-values/dashboard-prod.yaml`:

```yaml
database:
  image: postgres
  tag: "16.4"
  name: dashboard
  user: dashboard
  password: "${DASHBOARD_DB_PASSWORD}"
  port: 5432
```

Generate an env template for operators:

```bash
dockyard env template ./team-dashboard \
  --sensitive-only \
  -o ./deploy-values/dashboard-prod.env.example
```

PowerShell deployment:

```powershell
$env:DASHBOARD_DB_PASSWORD = "use-a-long-random-password"

.\bin\dockyard.exe install dashboard-prod .\team-dashboard `
  -f .\deploy-values\dashboard-prod.yaml
```

Validate the env example before committing it:

```bash
dockyard env check ./deploy-values/dashboard-prod.env.example
```

