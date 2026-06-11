# Contributing to Dockyard

Thanks for helping improve Dockyard.

## Local setup

```bash
go mod tidy
make dev-build
```

On Windows, `make build` writes `bin/dockyard.exe`. On Linux and macOS, it writes `bin/dockyard`.

## Before opening a pull request

Run:

```bash
make verify
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

For package or example changes, also run:

```bash
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

If Docker is available, run a smoke test for affected runnable examples:

```bash
dockyard package test ./examples/nginx --smoke
```

## Development guidelines

- Keep Docker Compose as the runtime source of truth.
- Do not add secrets, tokens, private keys, `.env` files, or operator-owned values to the repository.
- Keep generated test packages and deployment values outside the repo, for example under `../dockyard-work` and `../deploy-values`.
- Keep public docs consolidated. Prefer updating the existing guide files instead of adding one small doc per feature.
- Maintain Windows compatibility for paths, build output, and PowerShell examples.
- Keep `Dockyard.yaml` `dockyard.dev/v1alpha1` backward-compatible during v1.x.

## Security-sensitive changes

For changes touching archive extraction, path handling, environment files, OCI, or Docker command execution:

- add or update tests,
- avoid leaking secrets in errors or logs,
- preserve path traversal protections,
- avoid passing credentials to Dockyard commands when external tools can handle auth,
- prefer explicit allowlists over blocklists when possible.
