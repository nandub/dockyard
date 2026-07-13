# OCI Package Playbook

## Purpose

Change OCI package, archive, lockfile, provenance, ORAS, or registry behavior safely.

## When to Use It

Use for `oci://` references, package push/pull, `.dockyard.tgz`, `dockyard.lock`, provenance metadata, ORAS command construction, or registry round trips.

## Required Reading

- `AGENTS.md`
- `docs/oci.md`
- `docs/package-lifecycle.md`
- `docs/security.md`
- `docs/threat-model.md`
- `internal/oci/*`
- `internal/archive/*`
- `internal/lock/*`
- `internal/catalog/*`

## Preconditions

- Registry/network requirements are known.
- Execution plan exists for layout, protocol, or registry behavior changes.

## Procedure

1. Preserve precedence for local paths, archives, explicit `oci://`, catalog references, and shorthand resolution.
2. Keep ORAS authentication external.
3. Do not pass absolute archive paths to ORAS.
4. Preserve archive path safety and forbidden-file checks.
5. Preserve quiet pull behavior for JSON output modes.
6. Add tests for reference parsing, archive verification, and command construction.
7. Document registry-dependent validation that could not be run.

## Validation

- Focused tests for `internal/oci`, `internal/archive`, `internal/lock`, or `internal/catalog`.
- `go test ./...`
- `go build -o bin/dockyard$(go env GOEXE) ./cmd/dockyard`
- Registry round-trip only when explicitly available and permitted.

## Completion Checklist

- Reference precedence preserved.
- ORAS auth remains outside Dockyard.
- Archive verification remains strict.
- JSON output remains machine-readable.
- Registry validation is complete or explicitly not run.

## Escalation Conditions

- Registry credentials or network access are required.
- Signature, provenance, or artifact media type semantics change.
- Existing package compatibility is at risk.

## Required Completion Report

Report summary, files changed, design decisions, tests changed, validation results, unverified items, and remaining risks.
