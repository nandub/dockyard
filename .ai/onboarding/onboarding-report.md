# Onboarding Report

The original detailed onboarding reports were preserved under `.ai/onboarding/reports/`.

## Summary

Observed repository characteristics:

- Go module: `github.com/nandub/dockyard`.
- CLI entry point: `cmd/dockyard/main.go`.
- CLI framework: Cobra.
- Runtime delegation: Docker Compose through `internal/runner/`.
- OCI integration: external `oras` CLI through `internal/oci/` and `internal/catalog/`.
- Package model: `Dockyard.yaml`, values, schema, rendered Compose, archive, lockfile, provenance metadata, catalog metadata, and release state.

## Historical Reports

- `.ai/onboarding/reports/architecture.md`
- `.ai/onboarding/reports/component-map.md`
- `.ai/onboarding/reports/dependency-graph.md`
- `.ai/onboarding/reports/plugin-model.md`
- `.ai/onboarding/reports/development.md`
- `.ai/onboarding/reports/build.md`
- `.ai/onboarding/reports/ci.md`
- `.ai/onboarding/reports/release.md`
- `.ai/onboarding/reports/runtime.md`
- `.ai/onboarding/reports/configuration.md`
- `.ai/onboarding/reports/environment.md`
- `.ai/onboarding/reports/testing.md`
- `.ai/onboarding/reports/validation.md`
- `.ai/onboarding/reports/security.md`
- `.ai/onboarding/reports/threat-model.md`

Treat historical reports as discovery notes. Verify drift-prone claims against source, tests, Makefile, and CI.
