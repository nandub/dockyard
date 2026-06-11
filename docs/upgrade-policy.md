# Upgrade Policy

Dockyard is approaching `v1.0.0`. This document describes how compatibility is handled for release and stable releases.

## Stable format for v1.0

The following format is treated as the stable v1.0 package authoring contract:

```text
Dockyard.yaml
apiVersion: dockyard.dev/v1alpha1
```

Dockyard `v1.0.0` is expected to read packages using `dockyard.dev/v1alpha1`.

## Experimental formats

The following formats remain experimental during the release period:

```text
dockyard.lock
apiVersion: dockyard.dev/lockfile/v1alpha1

package.provenance.json
apiVersion: dockyard.dev/provenance/v1alpha1

release.json
apiVersion: dockyard.dev/release/v1alpha1
```

These files are useful and supported by the current CLI, but their exact schema may still change before `v1.0.0`.

## Compatibility expectations

Patch releases should not break working packages.

Minor pre-1.0 releases may change experimental formats, but should provide clear migration notes.

The first stable `v1.0.0` release should preserve the `Dockyard.yaml` v1alpha1 contract unless a security issue requires a breaking change.

## Package upgrade guidance

Before publishing or upgrading a package, run:

```bash
dockyard compat ./path/to/package --strict
dockyard package lint ./path/to/package --strict
dockyard package test ./path/to/package --strict
```

When the package is safe to start locally, also run:

```bash
dockyard package test ./path/to/package --smoke
```

## Release-state upgrade guidance

Dockyard stores release state under:

```text
$DOCKYARD_HOME/releases/<release>/
```

Do not manually edit release state while a release is deployed.

Before testing a new Dockyard CLI against important releases, back up `$DOCKYARD_HOME`.
