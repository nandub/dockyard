# Release Checklist

Use this checklist before tagging a release candidate such as `v1.0.0-rc.1` or a final release such as `v1.0.0`.

## Local verification

```bash
go mod tidy
make verify
make dev-build
dockyard version
dockyard compat
```

## Example package verification

```bash
dockyard lock ./examples/nginx
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

When Docker is available:

```bash
dockyard package test ./examples/nginx --smoke
```

Repeat the non-smoke gate for each public example package.

For OCI/catalog changes, run the public registry smoke gate with artifacts outside the repository:

```powershell
New-Item -ItemType Directory -Force ..\dockyard-artifacts | Out-Null
dockyard catalog list
dockyard pull redis -o ..\dockyard-artifacts\redis-0.2.0.dockyard.tgz
dockyard verify ..\dockyard-artifacts\redis-0.2.0.dockyard.tgz
dockyard package test redis --strict
```

## Release build

For local release snapshots:

```bash
make release-snapshot VERSION=v1.0.0
```

For GitHub releases, push a tag:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The GitHub release workflow builds cross-platform binaries, checksums, and an SBOM.

## Documentation review

Before tagging:

```text
README.md
docs/getting-started.md
docs/operator-guide.md
docs/packaging-and-distribution.md
docs/security.md
docs/compose-compatibility.md
docs/v1-readiness.md
docs/upgrade-policy.md
docs/support-policy.md
docs/release-candidate-checklist.md
AGENTS.md
CHANGELOG.md
```

## Compatibility decision

For `v1.0.0`:

```text
Stable:
  Dockyard.yaml apiVersion dockyard.dev/v1alpha1

Experimental:
  dockyard.lock
  package.provenance.json
  release.json
```
