# Validation Checklist

- Changed files match the requested scope.
- No generated, vendored, binary, archive, or local state files were edited manually.
- Focused tests were run for the affected package when code changed.
- `go fmt ./...` was run when Go files changed.
- `go test ./...` was run or skipped with a reason.
- Build command was run or skipped with a reason.
- Package validation was run or skipped with a reason when package behavior changed.
- Docker, registry, or network-dependent checks are clearly marked as run or not run.
- Registry smoke validation is reported separately when OCI/catalog behavior changes.
- Remaining unknowns are documented.
