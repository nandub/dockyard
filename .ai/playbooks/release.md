# Release Playbook

## Purpose

Prepare, validate, or document a Dockyard release while keeping code, docs, examples, CI, and artifacts aligned.

## When to Use It

Use for version changes, changelog updates, release candidates, tags, release workflows, binaries, SBOMs, checksums, and publication steps.

## Required Reading

- `AGENTS.md`
- `PLANS.md`
- `CHANGELOG.md`
- `docs/release-candidate-checklist.md`
- `docs/release-engineering.md`
- `docs/v1-readiness.md`
- `.github/workflows/release.yml`
- `.github/workflows/ci.yml`
- `.github/workflows/security.yml`

## Preconditions

- Target version and publish intent are clear.
- Git status and tag state inspected.

## Procedure

1. Confirm preparation versus publishing.
2. Inspect existing tags and release notes.
3. Keep version references, changelog, README, docs, examples, and workflows aligned.
4. Run release validation commands appropriate to scope.
5. Do not rewrite tags unless explicitly authorized.
6. Do not publish release assets unless explicitly requested.
7. Provide exact manual publish commands when publication remains user-owned.

## Validation

- `make verify`
- `make build`
- `dockyard --help`
- `dockyard doctor` when runtime checks are in scope.
- `dockyard compat ./examples/nginx --strict`
- `dockyard package lint ./examples/nginx --strict`
- `dockyard package test ./examples/nginx --strict`

## Completion Checklist

- Version references are consistent.
- Changelog/docs updated when appropriate.
- Release validation run or skipped with reasons.
- Generated release artifacts are not accidentally committed.
- Publish state is explicit.

## Escalation Conditions

- Tag rewrite is needed.
- GitHub publication or credentials are required.
- Release workflow fails in CI.

## Required Completion Report

Report summary, files changed, design decisions, tests changed, validation results, unverified items, and remaining risks.
