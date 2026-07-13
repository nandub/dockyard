# CLI Change Playbook

## Purpose

Add or change CLI behavior while preserving command contracts and output expectations.

## When to Use It

Use for commands, flags, arguments, help text, JSON output, exit behavior, shell completion, or command documentation.

## Required Reading

- `AGENTS.md`
- `docs/cli.md`
- `docs/command-reference.md`
- `docs/runtime.md`
- `internal/cli/*`

## Preconditions

- Existing command behavior is inspected.
- Execution plan exists for CLI contract changes.

## Procedure

1. Find the closest existing Cobra command pattern.
2. Define validation, errors, and output before editing.
3. Keep command code thin.
4. Preserve machine-readable JSON output where applicable.
5. Add `--dry-run` or `--force` safety only where it matches existing patterns.
6. Update command docs for user-visible changes.
7. Add tests for command parsing and internal behavior.

## Validation

- Focused CLI/package tests.
- `go test ./...`
- `go build -o bin/dockyard$(go env GOEXE) ./cmd/dockyard`
- Manual `dockyard COMMAND --help` check when help changes.

## Completion Checklist

- Command validates inputs clearly.
- Output contract is documented.
- Docs match actual flags.
- Tests cover success and failure paths.

## Escalation Conditions

- JSON output shape changes.
- Command mutates Docker, release state, archives, or registry state.
- Backward compatibility is unclear.

## Required Completion Report

Report summary, files changed, design decisions, tests changed, validation results, unverified items, and remaining risks.
