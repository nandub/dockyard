# Release Checklist

- Git status inspected.
- Target version and release scope confirmed.
- Existing tags checked before any tag operation.
- README, CHANGELOG, release docs, examples, and workflows are aligned when affected.
- `make verify` run or skipped with reason.
- `make build` run or skipped with reason.
- Package validation run or skipped with reason.
- Runtime checks run only when Docker is available and in scope.
- No generated release artifacts committed accidentally.
- Publish state is explicit.
