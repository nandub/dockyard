# OCI Registry Support

Dockyard v0.6 supports OCI distribution through the `oras` CLI.

## Requirements

Install `oras` and authenticate to your registry before using Dockyard OCI commands:

```bash
oras login ghcr.io
```

Dockyard delegates authentication to `oras`. Do not pass credentials directly to Dockyard commands.

## Push

```bash
dockyard push team-dashboard-0.1.0.dockyard.tgz \
  oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.0
```

By default, Dockyard verifies the local package archive before pushing.

## Pull

```bash
dockyard pull oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.0 \
  -o team-dashboard-0.1.0.dockyard.tgz
```

By default, Dockyard verifies the pulled package archive.

## Install from OCI

```bash
dockyard install dashboard-prod \
  oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.0 \
  -f ./deploy-values/dashboard-prod.yaml \
  --require-lock
```

## Upgrade from OCI

```bash
dockyard upgrade dashboard-prod \
  oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.1 \
  -f ./deploy-values/dashboard-prod.yaml \
  --require-lock
```

## Reference rules

Dockyard OCI references must:

- start with `oci://`
- include an explicit tag or digest
- not contain whitespace

Examples:

```text
oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.0
oci://ghcr.io/nandub/dockyard/team-dashboard@sha256:<digest>
```


## Values versus overlays

A production values file such as `deploy-values/dashboard-prod.yaml` does not automatically activate a production Compose overlay.

Use:

```bash
dockyard install dashboard-prod   oci://ghcr.io/nandub/dockyard/team-dashboard:0.1.0   -f ./deploy-values/dashboard-prod.yaml   --require-lock
```

Add `--overlay prod` only when the package defines a structural Compose override in `Dockyard.yaml`.

See [Compose Overlays](overlays.md).
