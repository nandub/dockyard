# Compose Overlays

Dockyard has two separate configuration layers:

- values files, passed with `-f`
- Compose overlays, passed with `--overlay`

Most packages should start with values files only. Add overlays when values are not enough.

## Values files answer: what settings should this environment use?

Example:

```bash
dockyard install dashboard-prod ./team-dashboard \
  -f ./deploy-values/dashboard-prod.yaml \
  --require-lock
```

The values file changes placeholder values such as:

```yaml
app:
  tag: "1.0.0"
  port: 8080

database:
  password: "set-outside-the-package"
```

Use values for:

- image tags
- ports
- hostnames
- usernames
- database names
- feature flags
- resource-size values
- environment-specific secrets or secret references

## Overlays answer: should the Compose structure change?

Example:

```bash
dockyard install dashboard-prod ./team-dashboard \
  -f ./deploy-values/dashboard-prod.yaml \
  --overlay prod \
  --require-lock
```

The overlay applies a named Compose file from `Dockyard.yaml`:

```yaml
compose:
  base: compose.yaml
  overlays:
    prod: compose.prod.yaml
```

Use overlays for structural differences such as:

- adding or removing services
- adding reverse-proxy labels
- changing volume layouts
- changing network topology
- adding production logging options
- changing restart policies
- adding production-only security hardening
- adding or removing exposed host ports

## Recommended default

Prefer this for most examples:

```bash
dockyard install dashboard-prod ./team-dashboard \
  -f ./deploy-values/dashboard-prod.yaml \
  --require-lock
```

Use this only when the package has a real structural production override:

```bash
dockyard install dashboard-prod ./team-dashboard \
  -f ./deploy-values/dashboard-prod.yaml \
  --overlay prod \
  --require-lock
```

A file named `dashboard-prod.yaml` does not automatically activate `compose.prod.yaml`. This is intentional. Values and overlays are explicit so operators can combine them safely.
