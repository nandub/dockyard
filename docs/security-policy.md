# Security Policy

Dockyard includes a built-in Compose policy linter. The linter is intentionally conservative: it focuses on common high-impact risks while still allowing Docker Compose to remain the runtime source of truth.

## Commands

List built-in checks:

```bash
dockyard policy list
```

Run policy checks against a package directory:

```bash
dockyard policy check ./team-dashboard -f ./deploy-values/dashboard-prod.yaml
```

Run policy checks against an archive:

```bash
dockyard policy check team-dashboard-0.1.0.dockyard.tgz \
  -f ./deploy-values/dashboard-prod.yaml
```

Run policy checks against an OCI package:

```bash
dockyard policy check oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.0 \
  -f ./deploy-values/dashboard-prod.yaml
```

## Manifest configuration

Policy behavior is configured in `Dockyard.yaml`:

```yaml
security:
  requireNonRoot: true
  requireHealthchecks: true
  requireReadOnlyRootFilesystem: false
  requireNoNewPrivileges: true
  requireCapDropAll: true
  disallowPrivileged: true
  disallowHostNetwork: true
  disallowDockerSocketMount: true
  disallowHostPathMounts: false
  disallowLatestTag: true
```

## Recommended defaults

For application services, prefer:

```yaml
services:
  app:
    image: ghcr.io/example/app:1.0.0
    user: "10001:10001"
    read_only: true
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 5
    security_opt:
      - no-new-privileges:true
    cap_drop:
      - ALL
```

For stateful infrastructure services such as databases, some checks may need to be disabled or documented because upstream images often manage their own user and filesystem behavior.

## Compatibility boundary

Dockyard does not model every Compose security setting. Advanced Compose fields may work at runtime while not yet being fully analyzed by the Dockyard policy engine.

Use `dockyard config` or `dockyard render --validate-compose` to validate the final rendered Compose file with Docker Compose itself.
