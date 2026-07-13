# Architecture

Dockyard is a Go CLI layered on Docker Compose. It renders, validates, packages, distributes, and records release state for Compose-based applications. Docker Compose remains the runtime source of truth.

Evidence:

- `cmd/dockyard/main.go`
- `internal/cli/root.go`
- `internal/runner/`
- `internal/archive/`
- `internal/oci/`
- `internal/catalog/`

## Major Boundaries

- CLI layer: `internal/cli/`
- Package manifest layer: `internal/dockpkg/`
- Values/schema layer: `internal/values/`
- Render layer: `internal/render/`
- Policy and compatibility layer: `internal/policy/`
- State layer: `internal/state/`
- Docker/Compose runner layer: `internal/runner/`
- Archive and lock layer: `internal/archive/`, `internal/lock/`
- OCI/catalog layer: `internal/oci/`, `internal/catalog/`

## Architectural Rules

- Keep CLI code thin.
- Put testable behavior in `internal/*`.
- Do not execute arbitrary package hooks or scripts.
- Keep OCI authentication outside Dockyard; current behavior delegates to `oras`.
- Preserve Windows-safe path handling.
- Treat package archives and package paths as untrusted input.

## Historical Detail

See `.ai/onboarding/reports/architecture.md` and `.ai/onboarding/reports/component-map.md` for the original static discovery notes.
