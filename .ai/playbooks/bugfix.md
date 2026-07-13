# Bug Fix Playbook

## Purpose

Fix a defect with the smallest safe change and regression coverage where practical.

## When to Use It

Use for reported bugs, failing tests, regressions, incorrect docs examples, or inconsistent CLI behavior.

## Required Reading

- `AGENTS.md`
- `docs/runtime.md`
- `docs/testing.md`
- Relevant source and tests.
- Related docs if user-facing behavior is affected.

## Preconditions

- Reproduction or evidence of the defect is understood.
- Expected behavior is known or documented as Unknown.

## Procedure

1. Reproduce or inspect the defect with the least destructive command.
2. Locate the owning package.
3. Add or identify a regression test when feasible.
4. Fix the root cause without broad cleanup.
5. Preserve behavior outside the defect.
6. Update docs only when documented behavior was wrong.
7. Run focused validation before broader validation.

## Validation

- Focused package tests.
- `go test ./...` for shared behavior.
- `go fmt ./...` if Go files changed.
- Runtime/package checks only when the defect touches Docker, Compose, OCI, or package validation.

## Completion Checklist

- Defect cause identified.
- Regression coverage added or reason documented.
- Fix is scoped.
- Validation results reported.
- Remaining environmental dependencies are called out.

## Escalation Conditions

- Fix requires contract changes.
- Reproduction requires Docker or registry access.
- The bug appears to be in Docker Compose, registry infrastructure, or another delegated system.

## Required Completion Report

Report summary, files changed, design decisions, tests changed, validation results, unverified items, and remaining risks.
