# v1.0 readiness

Dockyard v0.11 starts the v1.0 compatibility pass. The goal is to make file formats, command behavior, and release state predictable before declaring a stable v1.0.

## Format stability

Dockyard currently uses these format versions:

| Format | API version | Stability |
| --- | --- | --- |
| Package manifest | `dockyard.dev/v1alpha1` | Stable candidate |
| Lockfile | `dockyard.dev/lockfile/v1alpha1` | Experimental |
| Package provenance | `dockyard.dev/provenance/v1alpha1` | Experimental |
| Release state | `dockyard.dev/release/v1alpha1` | Experimental |

`Dockyard.yaml` is the most important compatibility contract. Other formats may still receive small changes before v1.0.

## Compatibility command

Print supported formats:

```bash
dockyard compat
```

Check a local package:

```bash
dockyard compat ./examples/nginx
```

Check a package archive:

```bash
dockyard compat nginx-0.1.0.dockyard.tgz
```

Check release state:

```bash
dockyard compat --release example
```

JSON output is available:

```bash
dockyard compat --json
dockyard compat ./examples/nginx --json
dockyard compat --release example --json
```

## Release state

New v0.11 release revisions include an `apiVersion` field in `release.json`:

```json
{
  "apiVersion": "dockyard.dev/release/v1alpha1",
  "dockyardVersion": "v0.11.0"
}
```

Dockyard can still read older release records that do not have `apiVersion`; they are treated as legacy v0.x records.

## v1.0 review checklist

Before v1.0, review:

- CLI argument order and flag names.
- `Dockyard.yaml` fields and defaults.
- `values.schema.json` expectations.
- `dockyard.lock` digest behavior.
- Package archive structure and forbidden-file rules.
- OCI push/pull behavior.
- Release state directory layout.
- Windows, Linux, and macOS behavior.
- Security policy defaults.
- Documentation examples and command reference.
