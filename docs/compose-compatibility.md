# Docker Compose Compatibility

Dockyard does not replace Docker Compose. Dockyard renders standard Docker Compose YAML and delegates runtime behavior to the Docker Compose CLI.

## Compatibility model

Dockyard has two layers:

```text
Dockyard package layer:
  Dockyard.yaml
  values.yaml
  values.schema.json
  dockyard.lock
  package archives
  release state
  policy checks

Docker Compose runtime layer:
  services
  networks
  volumes
  configs
  secrets
  profiles
  labels
  healthchecks
  ports
  environment
  and other Compose features
```

Most Compose features can be used in `compose.yaml` as long as the final rendered file is valid according to:

```bash
docker compose config
```

Dockyard v0.6.2 does not claim full internal awareness of every Compose feature. Advanced Compose fields may work at runtime while Dockyard's policy engine, lockfile generation, package verification, and diff output only reason about a focused subset.

## Validate rendered Compose

Use `dockyard config` to render a package and run Docker Compose validation without installing it:

```bash
dockyard config ./team-dashboard -f ./deploy-values/dashboard-prod.yaml
```

For archives:

```bash
dockyard config team-dashboard-0.1.0.dockyard.tgz \
  -f ./deploy-values/dashboard-prod.yaml \
  --require-lock
```

For OCI packages:

```bash
dockyard config oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.0 \
  -f ./deploy-values/dashboard-prod.yaml \
  --require-lock
```

You can also validate during rendering:

```bash
dockyard render ./team-dashboard \
  -f ./deploy-values/dashboard-prod.yaml \
  --validate-compose
```

## Recommended production package style

For production-friendly Dockyard packages, prefer:

- Prebuilt images over local `build` contexts.
- Explicit image tags or digests over `latest`.
- Named volumes over broad host-path mounts.
- Values files for environment-specific settings.
- Compose overlays only for structural changes.
- `docker compose config` validation before install.
- `dockyard lock` and `--require-lock` before installing packaged artifacts.

## Features that need extra care

These Compose features may work after rendering, but package authors should use them carefully:

| Compose feature | Guidance |
|---|---|
| `build` | Prefer prebuilt images for distributable packages. Build contexts can accidentally include large or sensitive files. |
| `env_file` | Avoid packaging `.env` files. Dockyard rejects common secret-like files from archives. |
| host-path `volumes` | Prefer named volumes. Host paths are environment-specific and harder to verify. |
| `extends` / `include` | Use cautiously. Make sure referenced files are included in the package and do not escape the package root. |
| `profiles` | Supported by Compose, but Dockyard does not yet provide first-class profile management. |
| `secrets` / `configs` | Supported by Compose, but Dockyard's secret-management story is intentionally conservative in this MVP. |
| `deploy` | Compose accepts some deploy fields depending on runtime mode. Validate with `dockyard config`. |

## Policy engine scope

Dockyard's policy linter currently checks selected high-value risks, including:

```text
privileged containers
host networking
Docker socket mounts
missing healthchecks
root users, when enabled
latest or implicit latest image tags
```

The linter is intentionally not a complete Compose security scanner yet.

## Validation output behavior

`dockyard render --validate-compose` validates the rendered Compose file quietly and then prints Dockyard's rendered YAML.

`dockyard config` intentionally prints Docker Compose's normalized `docker compose config` output. Use it when you want to see how Docker Compose interprets the final file.


## Multi-service and multi-application packages

Dockyard packages can contain any valid Docker Compose project. One package may include many services, such as a frontend, API, worker, database, cache, and reverse proxy. Dockyard renders and manages that Compose project as one Dockyard release.

Package dependency metadata is different from Compose services. A package can declare dependencies in `Dockyard.yaml` to document related packages, but v1.2 does not automatically install those dependency packages. Use one package with multiple Compose services for tightly coupled deployments. Use separate packages plus dependency metadata for independently managed components.

