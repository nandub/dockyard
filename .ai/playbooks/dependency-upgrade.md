# Dependency Upgrade Playbook

## Purpose

Upgrade Go modules, GitHub Actions, workflow tools, or external tool references without changing product behavior.

## When to Use It

Use for dependency version changes, vulnerability remediation, toolchain updates, or CI action updates.

## Required Reading

- `AGENTS.md`
- `docs/development.md`
- `docs/build.md`
- `docs/security.md`
- `go.mod`
- `go.sum`
- Relevant `.github/workflows/*`

## Preconditions

- Dependency type and reason for upgrade are clear.
- Network access requirements are known.

## Procedure

1. Identify whether the dependency is a Go module, workflow action, CLI tool, Docker image, or external runtime.
2. Inspect current usage in source, tests, docs, and CI.
3. Prefer the smallest version movement that satisfies the request.
4. Run `go mod tidy` for Go module changes.
5. Review `go.sum` for expected transitive movement.
6. Update docs only if commands or requirements changed.

## Validation

- `go mod tidy`
- `go test ./...`
- `go build -o bin/dockyard$(go env GOEXE) ./cmd/dockyard`
- `make verify` when Make is available.

## Completion Checklist

- Version movement is intentional.
- Module files are consistent.
- CI and docs are aligned when affected.
- New dependency is justified.

## Escalation Conditions

- Network access is required.
- Upgrade breaks public behavior or supported platforms.
- License, maintenance, or security posture is unclear.

## Required Completion Report

Report summary, files changed, design decisions, tests changed, validation results, unverified items, and remaining risks.
