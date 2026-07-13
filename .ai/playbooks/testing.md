# Testing Playbook

## Purpose

Add, update, or analyze tests and validation coverage for Dockyard.

## When to Use It

Use for Go tests, package validation, smoke tests, CI validation, coverage, benchmarks, or missing validation analysis.

## Required Reading

- `AGENTS.md`
- `docs/testing.md`
- `docs/validation.md`
- Relevant source and existing tests.
- `.github/workflows/ci.yml`
- `.github/workflows/security.yml`

## Preconditions

- Target behavior and validation level are clear.
- Runtime prerequisites such as Docker or ORAS are known.

## Procedure

1. Identify whether the needed coverage is unit, integration-like, package validation, smoke, performance, or CI.
2. Prefer fast deterministic Go tests for internal logic.
3. Keep Docker and registry tests optional unless the workflow explicitly requires them.
4. Avoid writing generated packages, values, archives, or rendered files into the repository.
5. Add regression coverage for bug fixes when practical.
6. Update validation docs when the validation strategy changes.

## Validation

- Focused package tests.
- `go test ./...`
- `make verify`
- Package validation commands when examples or package behavior change.
- Docker smoke tests only when available and safe.

## Completion Checklist

- Test scope matches risk.
- Tests are deterministic where possible.
- Runtime-dependent checks are clearly marked.
- Validation results are reported.

## Escalation Conditions

- Docker, registry, or network access is required.
- A test would be flaky due to external services.
- Coverage requires changing package formats or runtime behavior.

## Required Completion Report

Report summary, tests changed, validation results, skipped checks, unverified items, and remaining risks.
