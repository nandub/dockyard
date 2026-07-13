# Repository Cleanup Playbook

## Purpose

Remove stale, duplicated, generated, or misleading repository material without deleting source of truth or user work.

## When to Use It

Use for cleanup of docs, generated outputs, obsolete onboarding material, unused examples, stale references, or duplicated instructions.

## Required Reading

- `AGENTS.md`
- `docs/index.md`
- `.gitignore`
- Relevant docs, source, scripts, and workflows.

## Preconditions

- Git status inspected.
- Cleanup target and deletion permission are clear.

## Procedure

1. Classify targets as source, docs, generated output, cache, artifact, vendored code, temporary data, or historical onboarding.
2. Search references before removal.
3. Consolidate useful duplicated content before deleting copies.
4. Do not delete unrelated user changes.
5. Keep removals scoped.
6. Report exact paths removed.

## Validation

- `git status --short`
- Reference search for removed paths.
- Link/path review for docs cleanup.
- `go test ./...` only when source, tests, examples, or package data changed.

## Completion Checklist

- Useful content consolidated.
- Removed paths are obsolete.
- References updated or verified absent.
- User-owned changes preserved.

## Escalation Conditions

- A file might be generated but source of truth is unclear.
- Cleanup would remove tracked source, tests, examples, or docs with unclear replacement.

## Required Completion Report

Report summary, files removed or moved, files merged, validation results, unverified items, and remaining risks.
