# Dockyard Glossary

This glossary is for maintainers and AI agents. It defines repository terminology observed in source, documentation, workflows, and onboarding discovery.

## Repository Glossary

`cmd/dockyard/`
: Go CLI entry point. `main.go` constructs and executes the Cobra root command.

`internal/cli/`
: Cobra command wiring and CLI-facing behavior. Keep this layer thin where practical.

`internal/dockpkg/`
: Dockyard package manifest loading, package path handling, and path safety helpers.

`internal/values/`
: Values loading, merging, defaults, and JSON Schema validation.

`internal/render/`
: Placeholder rendering and render diagnostics for Compose templates.

`internal/policy/`
: Security and compatibility policy checks against rendered Compose data.

`internal/state/`
: Dockyard home, release records, revisions, and metadata persistence.

`internal/runner/`
: Docker and Docker Compose subprocess integration.

`internal/archive/`
: `.dockyard.tgz` package creation, extraction, and verification.

`internal/lock/`
: `dockyard.lock` creation and verification.

`internal/oci/`
: OCI package push/pull integration through the embedded ORAS Go client.

`internal/catalog/`
: Catalog source resolution and OCI/local catalog metadata loading.

`internal/envfile/`
: Env-file template and env-file validation helpers.

`internal/quality/`
: Package quality checks used by package validation flows.

`internal/format/`
: Centralized format/API version constants.

`internal/version/`
: CLI version metadata.

`tools/fmtcheck/`
: Cross-platform formatting check helper used by `make fmt-check`.

`examples/`
: Dockyard package examples used for documentation, package quality checks, and validation.

`.ai/onboarding/`
: AI-facing historical repository discovery notes. These are not user-facing product docs.

## Architecture Glossary

Dockyard package
: A directory or archive containing `Dockyard.yaml` and related package files such as values, schema, Compose templates, README, lockfile, or provenance metadata.

Package manifest
: `Dockyard.yaml`. The stable observed API version is `dockyard.dev/v1alpha1`.

Values file
: `values.yaml` plus optional override files passed by the operator.

Values schema
: `values.schema.json`, used to validate merged values.

Rendered Compose
: Docker Compose YAML produced by Dockyard after applying values to package templates.

Docker Compose runtime
: Docker Compose remains the runtime source of truth. Dockyard renders and validates Compose, then delegates runtime behavior.

Release
: A named deployment tracked under Dockyard state.

Revision
: A persisted release revision under `releases/<release>/revisions/<revision>/`.

Dockyard home
: State root resolved by `--home`, then `DOCKYARD_HOME`, then `~/.dockyard`.

Release metadata
: `release.json` stored in release state. It records deployment metadata and should not store secret values.

Lockfile
: `dockyard.lock`, generated and verified by `internal/lock`.

Package archive
: `.dockyard.tgz` archive produced and verified by `internal/archive`.

Provenance metadata
: `package.provenance.json`, package metadata generated for archives. It is not an observed signature mechanism.

Catalog
: Metadata index used to resolve `catalog://NAME[:VERSION]` and bare package shorthand.

OCI package
: Dockyard package archive pushed to or pulled from an OCI registry with the embedded ORAS Go client.

ORAS
: OCI artifact tooling and Go library used by Dockyard for registry push and pull operations.

Policy check
: Dockyard checks for selected Compose risks such as privileged mode, host networking, Docker socket mounts, host paths, and image tag concerns.

Strict mode
: Mode where warnings fail for compatibility and package-quality commands.

Advisory warning
: Warning category that can be allowed only where commands explicitly support advisory exceptions.

Dependency metadata
: `Dockyard.yaml` dependency declarations used by dependency commands and explicit dependency install planning.

Install plan
: Read-only dependency/deployment plan generated before installation.

## Command Glossary

`go mod tidy`
: Reconcile Go module files. Mutates `go.mod` or `go.sum` when dependencies drift.

`go fmt ./...`
: Format Go source.

`go test ./...`
: Run Go tests across the module.

`go build -o bin/dockyard$(go env GOEXE) ./cmd/dockyard`
: Build the platform-native Dockyard binary.

`make tidy`
: Wrapper for Go module tidying.

`make fmt`
: Wrapper for formatting Go source.

`make fmt-check`
: Cross-platform formatting verification using `go run ./tools/fmtcheck`.

`make tidy-check`
: Go module tidy verification using `go mod tidy -diff`.

`make test`
: Wrapper for `go test ./...`.

`make build`
: Non-mutating platform-native binary build.

`make dev-build`
: Local convenience target that may tidy, format, and build.

`make verify`
: CI-style local verification: tidy-check, fmt-check, and tests.

`make clean`
: Remove build output under `bin/`.

`dockyard doctor`
: Check local runtime prerequisites.

`dockyard init`
: Generate a starter Dockyard package.

`dockyard values template`
: Generate operator values from package defaults.

`dockyard lint`
: Validate a package with values and policy checks.

`dockyard render`
: Render package Compose YAML.

`dockyard render --validate-compose`
: Render Compose and validate it with Docker Compose while keeping rendered YAML output quiet.

`dockyard config`
: Print Docker Compose normalized config output.

`dockyard install`
: Install a release.

`dockyard install --with-dependencies`
: Explicitly install declared dependencies before the root package.

`dockyard install --dry-run`
: Plan install behavior without writing release state or mutating Docker.

`dockyard install-plan`
: Read-only install/dependency planning command.

`dockyard upgrade`
: Upgrade an existing deployed release.

`dockyard status`
: Show release status.

`dockyard status --compose-ps`
: Include Docker Compose `ps` status.

`dockyard uninstall`
: Uninstall a release.

`dockyard package lint`
: Run package quality checks.

`dockyard package test`
: Run package validation. Default mode is non-destructive.

`dockyard package test --smoke`
: Run smoke validation for examples safe to start and stop locally.

`dockyard compat`
: Run compatibility checks.

## Common Terminology

Authoritative command
: The command treated as the source of truth for a workflow. CI workflows and Makefile targets usually provide the strongest evidence.

Generated file
: A file produced by build, render, package, smoke-test, coverage, or release tooling. Do not edit generated files manually.

Vendored code
: Third-party source copied into the repository. No active vendored dependency tree was identified during onboarding discovery.

Artifact
: Build output, package archive, release binary, checksum, SBOM, rendered Compose, or validation output.

Cache
: Reusable local data such as catalog cache or Go/tool caches. Do not treat cache contents as source.

Temporary directory
: Directory created with `os.MkdirTemp` or external tooling for transient package extraction, render, pull, config, or test work.

Smoke test
: End-to-end runtime check that may interact with Docker and should keep generated packages, values, and artifacts outside the repository.

Machine-readable output
: JSON or other structured output intended for automation. Keep quiet OCI pull paths and avoid human-only noise in JSON modes.

Source of truth
: The layer responsible for final behavior. For runtime container behavior, Docker Compose is the source of truth.

Trust boundary
: A place where Dockyard accepts input from an operator, package, archive, registry, subprocess, or external system.

Secret-like value
: A value or key that appears to contain credentials, tokens, passwords, API keys, or private material and must not be printed in clear text.
