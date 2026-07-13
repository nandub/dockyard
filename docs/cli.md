# CLI Model

Dockyard uses Cobra for command dispatch.

Evidence:

- `cmd/dockyard/main.go`
- `internal/cli/root.go`
- `internal/cli/*.go`

## Observed Commands

Command groups include:

- Core package operations: `init`, `lint`, `render`, `config`.
- Catalog operations: `catalog list`, `catalog info`.
- Release lifecycle: `install`, `install-plan`, `upgrade`, `rollback`, `uninstall`, `status`, `inspect`, `list`, `diff`, `prune`.
- Packaging and distribution: `lock`, `package`, `verify`, `push`, `pull`.
- Diagnostics and policy: `doctor`, `policy`, `secrets`, `compat`, `version`.
- Values and env helpers: `values`, `env`.

## Output

Machine-readable output must remain machine-readable. JSON output paths should avoid human-only noise, especially when OCI pulls are required.

## Completion

No dedicated shell-completion documentation was created during onboarding. Verify current Cobra setup before changing completion behavior.

Historical detail: `.ai/onboarding/reports/architecture.md`.
