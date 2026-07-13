# Runtime Model

## Startup

Observed startup sequence:

1. `cmd/dockyard/main.go` calls `cli.NewRootCommand().Execute()`.
2. `internal/cli/root.go` constructs the Cobra root command.
3. Subcommands are registered.
4. Cobra parses args and dispatches to command handlers.
5. Errors return to `main`, which exits non-zero.

No daemon, server, or background worker was observed.

## Configuration and Package Loading

Runtime commands load package manifests, values files, schemas, env files, lockfiles, archives, catalog metadata, or OCI artifacts depending on command.

## External Execution

Dockyard delegates:

- Docker and Docker Compose operations to external Docker/Compose commands.
- OCI package push/pull and catalog artifact pull through the embedded ORAS Go client.

## Lifecycle

Release lifecycle commands include install, upgrade, rollback, status, uninstall, inspect, list, diff, and prune.

## Shutdown

No long-running shutdown lifecycle was observed. Commands return after completing their subprocess work.

Historical detail: `.ai/onboarding/reports/runtime.md`.
