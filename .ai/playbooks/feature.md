# Feature Playbook

## Purpose

Add a user-visible Dockyard capability without changing unrelated behavior.

## When to Use It

Use for new commands, flags, package behavior, validation behavior, or documented workflows that users can observe.

## Required Reading

- `AGENTS.md`
- `PLANS.md`
- `docs/architecture.md`
- `docs/domain-model.md`
- Relevant `internal/*` package.
- Existing tests for the affected package.

## Preconditions

- Git status inspected.
- Current behavior understood from source and tests.
- Execution plan created if the feature crosses components or changes contracts.

## Procedure

1. Identify the owning package boundary.
2. Find the closest existing implementation pattern.
3. Keep Cobra command code focused on parsing and presentation.
4. Put reusable logic in `internal/*`.
5. Preserve JSON output shape unless a change is explicitly requested.
6. Update docs and examples for user-visible behavior.
7. Add tests for new behavior and important failure paths.

## Validation

- `go fmt ./...`
- Focused `go test ./internal/PACKAGE`
- `go test ./...`
- `go build -o bin/dockyard$(go env GOEXE) ./cmd/dockyard`
- Package or runtime validation when affected.

## Completion Checklist

- Feature is scoped to the request.
- Tests cover the new behavior.
- Docs match actual commands.
- Generated artifacts are not committed.
- Skipped validation is explained.

## Escalation Conditions

- Feature changes file formats, state layout, OCI behavior, or CLI contracts.
- Docker, registry, or network access is required.
- Requirements conflict with existing docs or tests.

## Required Completion Report

Report summary, files changed, design decisions, tests changed, validation results, unverified items, and remaining risks.
