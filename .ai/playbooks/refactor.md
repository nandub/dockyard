# Refactor Playbook

## Purpose

Improve internal structure without changing user-visible behavior.

## When to Use It

Use for behavior-preserving code movement, duplication removal, naming cleanup, or package boundary improvements.

## Required Reading

- `AGENTS.md`
- `PLANS.md`
- `docs/architecture.md`
- `docs/repository-map.md`
- Affected source and tests.

## Preconditions

- Existing behavior is understood.
- An execution plan exists for substantial refactoring.

## Procedure

1. Identify the behavior that must remain unchanged.
2. Keep changes within the smallest package boundary.
3. Avoid unrelated style churn.
4. Preserve CLI output, JSON output, file formats, and errors unless authorized.
5. Preserve Windows path behavior.
6. Run tests before and after when risk is non-trivial.

## Validation

- `go fmt ./...`
- Focused package tests.
- `go test ./...`
- `make verify` for broad refactors.

## Completion Checklist

- No intentional behavior changes.
- Tests still cover the affected behavior.
- No unrelated files rewritten.
- Any behavior drift is reported.

## Escalation Conditions

- Public behavior changes become necessary.
- Refactor touches persistent state, OCI, archive, catalog, or CLI contracts.

## Required Completion Report

Report summary, files changed, design decisions, tests changed, validation results, unverified items, and remaining risks.
