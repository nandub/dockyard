# Catalog Change Playbook

## Purpose

Change Dockyard catalog source, resolution, cache, or shorthand behavior safely.

## When to Use It

Use for `DOCKYARD_CATALOG`, `catalog://NAME[:VERSION]`, bare install shorthand, registry-prefix compatibility, or catalog metadata loading.

## Required Reading

- `AGENTS.md`
- `docs/oci.md`
- `docs/package-lifecycle.md`
- `internal/catalog/*`
- `internal/cli/catalog*.go`
- Catalog-related tests.

## Preconditions

- Desired catalog behavior and precedence are clear.
- Execution plan exists for resolution or cache behavior changes.

## Procedure

1. Preserve precedence for explicit local paths, archives, and `oci://` sources over catalog shorthand.
2. Preserve `DOCKYARD_CATALOG` behavior unless explicitly changed.
3. Avoid hardcoding package names.
4. Keep JSON output quiet and machine-readable.
5. Add tests for source normalization, catalog lookup, cache behavior, and errors.
6. Document registry-dependent validation that could not be run.

## Validation

- Focused catalog tests.
- `go test ./...`
- CLI build.
- Registry-backed catalog validation only when permitted and credentials/network are available.

## Completion Checklist

- Resolution precedence is preserved.
- Catalog tests cover changed behavior.
- Docs reflect actual source forms.
- Registry validation status is explicit.

## Escalation Conditions

- Live registry access is required.
- Catalog format compatibility changes.
- JSON output may change.

## Required Completion Report

Report summary, files changed, design decisions, tests changed, validation results, unverified items, and remaining risks.
