# Support Policy

Dockyard is a Docker Compose package manager and release tracker. Docker Compose remains the runtime source of truth.

## Supported platforms

Dockyard release builds target:

```text
windows/amd64
linux/amd64
linux/arm64
darwin/amd64
darwin/arm64
```

Windows support is tested with Docker Desktop and PowerShell examples.

## Supported Docker Compose behavior

Dockyard renders standard Compose YAML and delegates runtime behavior to the Docker Compose CLI.

Dockyard does not attempt to implement the full Compose specification internally. Advanced Compose features may work when the rendered file is valid, but Dockyard's package verification, policy checks, and lockfile behavior may understand only a subset.

Use:

```bash
dockyard config PACKAGE_SOURCE
dockyard render PACKAGE_DIR --validate-compose
```

to verify rendered Compose output.

## Security support

Dockyard provides policy checks and package-quality checks, but it is not a vulnerability scanner for every container image, application, or Compose feature.

Package authors should also run image scanning and dependency scanning in their own CI.

## OCI support

OCI push and pull currently use the external `oras` CLI. Run:

```bash
dockyard doctor
```

to check whether `oras` is available.

## Best-effort compatibility

Dockyard will try to keep package manifests stable across `v1.x`.

Experimental files such as lockfiles, provenance metadata, and release-state metadata may evolve until they are explicitly marked stable.
