# v1.0 readiness

Dockyard v1.0 establishes the package manifest compatibility contract for the v1.x line.

`Dockyard.yaml` with `apiVersion: dockyard.dev/v1alpha1` is the supported package manifest format. Other Dockyard-generated formats remain explicitly experimental so they can continue to evolve.

## Format stability

Dockyard currently uses these format versions:

| Format | API version | Stability |
| --- | --- | --- |
| Package manifest | `dockyard.dev/v1alpha1` | Stable |
| Lockfile | `dockyard.dev/lockfile/v1alpha1` | Experimental |
| Package provenance | `dockyard.dev/provenance/v1alpha1` | Experimental |
| Release state | `dockyard.dev/release/v1alpha1` | Experimental |

`Dockyard.yaml` is the primary v1.x compatibility contract. Other formats may still receive compatible or documented changes while they remain experimental.

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

## v1.x compatibility checklist

Before making changes in the v1.x line, review:

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


## Package quality gate

Example and catalog-ready packages should pass:

```bash
dockyard compat PACKAGE_DIR
dockyard package lint PACKAGE_DIR --strict
```

`compat` checks format support. `package lint` checks package quality and publication readiness.


## v1.0 release gate

Before cutting a `v1.0.0` tag or later v1.x release tag, run this checklist from a clean checkout:

```bash
go mod tidy
make verify
make dev-build
dockyard version
dockyard compat
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

When Docker is available, also run:

```bash
dockyard package test ./examples/nginx --smoke
```

On Windows, use the generated executable:

```powershell
.\bin\dockyard.exe compat .\examples\nginx --strict
.\bin\dockyard.exe package test .\examples\nginx --smoke
```

### Strict compatibility checks

`dockyard compat --strict`, `dockyard package lint --strict`, and `dockyard package test --strict` treat warnings as failures. Use `--allow-advisory` with package lint/test only when private packages intentionally allow advisory warnings such as a missing package-local `LICENSE`.

Warnings are still useful during normal development; strict mode is intended for packages or release state that should be ready for publishing.


## v1.0.0-rc.1 scope

`v1.0.0-rc.1` is a release preparation release, not a feature-expansion release.

The release candidate freezes the intended package manifest contract:

```text
Dockyard.yaml
apiVersion: dockyard.dev/v1alpha1
```

The following formats remain experimental during the release period:

```text
dockyard.lock
package.provenance.json
release.json
```

## Release-candidate gate

For public packages:

```bash
dockyard lock ./examples/nginx
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

For runnable examples:

```bash
dockyard package test ./examples/nginx --smoke
```

For private packages that intentionally rely on repository-level licensing:

```bash
dockyard package lint ./private-package --strict --allow-advisory
dockyard package test ./private-package --strict --allow-advisory
```
