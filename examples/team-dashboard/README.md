# Team dashboard dependency example

This example demonstrates Dockyard package dependency metadata.

The package itself is a small runnable nginx-based dashboard placeholder. Its `Dockyard.yaml` declares a metadata-only dependency on a PostgreSQL package:

```yaml
dependencies:
  - name: postgres
    alias: db
    source: oci://ghcr.io/nandub/dockyard/postgres:0.1.0
```

Dockyard v1.2 validates and displays this dependency, but does not automatically install it.

```bash
dockyard package deps ./examples/team-dashboard
dockyard package lint ./examples/team-dashboard --strict
```
