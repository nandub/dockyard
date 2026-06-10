# Changelog

## v0.14.1

- Normalized `--strict` behavior for package quality commands: warnings now fail by default.
- Added `--allow-advisory` to `dockyard package lint` and `dockyard package test`.
- Marked missing package-local `LICENSE` as an advisory warning that can be allowed for private/internal packages.
- Added package-local `LICENSE` files to public examples so strict release gates can pass cleanly.
- Added tests for strict/advisory quality behavior.
- Updated command reference, packaging docs, v1 readiness docs, README, and AGENTS.md.

## v0.14.0

Release-candidate preparation pass.

- Added `dockyard compat --strict` to treat compatibility warnings as failures.
- Expanded v1.0 readiness documentation with a release-candidate gate.
- Updated command reference with strict compatibility usage.
- Updated README and AGENTS.md with release-candidate validation guidance.

## v0.13.1

- Removed duplicate `Package: name@version` output from `dockyard package test`.
- Added Docker preflight checks before `dockyard package test --smoke`.
- Improved smoke test errors so Docker Desktop / Docker daemon issues point users to `dockyard doctor`.
- Documented that `--smoke` requires a reachable Docker daemon.

## v0.13

- Added `dockyard package test PACKAGE_SOURCE` for package-author validation pipelines.
- Added optional `dockyard package test --smoke` for examples that can be safely started and stopped with Docker Compose.
- Added `--env-file`, `--require-lock`, `--overlay`, `--strict`, and JSON support to package tests.
- Updated command reference and packaging docs with package test workflows.


## v0.12.0

- Added `dockyard package lint PACKAGE_DIR`.
- Added package quality checks for recommended docs, forbidden local artifacts, values/schema validation, schema descriptions, sensitive schema markers, default rendering, and configured policy findings.
- Added `--strict` and `--json` output for package quality checks.
- Added `internal/quality` with unit tests.
- Updated command reference, packaging documentation, README, and AGENTS.md.


## v0.11.0

- Added `dockyard compat` for format status and package/release compatibility checks.
- Added `internal/format` to centralize Dockyard format API versions.
- Added `apiVersion` to newly written `release.json` records.
- Kept backwards compatibility for legacy release records without `apiVersion`.
- Added `docs/v1-readiness.md`.
- Updated command reference and README compatibility documentation.

## v0.10.4

- Fixed `make fmt-check` on Windows by replacing shell-specific `gofmt -l` logic with a small Go-based formatter check.
- Updated `make fmt` to run `go fmt ./...` so formatting covers all Go packages, including developer tools.
- Kept `make build` non-mutating and `make dev-build` as the local convenience target.


## v0.10.3

Developer experience patch.

- Added `make dev-build` for local development builds that run `go mod tidy`, `gofmt`, and `make build`.
- Added `make verify` for pre-commit/CI checks.
- Added `make fmt-check` and `make tidy-check`.
- Kept `make build` non-mutating so builds do not silently edit `go.mod` or `go.sum`.
- Updated README development target documentation.

## v0.10.2

- Fixed Windows Makefile build metadata generation.
- Escaped Git pretty-format `%cI` correctly for Make.
- Avoided Unix-only `/dev/null` and `date -u` usage in Makefile.
- Added Windows-compatible `release-snapshot` and `clean` targets.

## v0.10.1

Examples expansion release.

- Added `examples/caddy-letsencrypt` for automatic HTTPS with Caddy.
- Added `examples/nginx-tls-mounted-certs` for operator-provided certificate and key files.
- Added `examples/traefik-letsencrypt` for Docker-label routing and Let's Encrypt with Traefik.
- Updated README and getting-started docs to explain example package choices.
- Kept TLS behavior modeled through standard Docker Compose; Dockyard does not manage certificates itself.

## v0.10.0

- Added `dockyard version` with build metadata and JSON output.
- Added Dockyard CLI version to release revision metadata.
- Added GitHub Actions release workflow for Windows, Linux, and macOS binaries.
- Added release checksums and SBOM generation to the release workflow.
- Added `docs/command-reference.md` and `docs/release-engineering.md`.
- Added runnable examples under `examples/nginx` and `examples/postgres-app`.
- Updated README install instructions for release artifacts.
- Updated Makefile build flags and cross-platform snapshot build target.


## v0.9.3

- Added `AGENTS.md` with repository guidance for code agents and automation.

## v0.9.2

Documentation consolidation release.

- Consolidated `docs/` from many feature-specific files into six maintained guides:
  - `getting-started.md`
  - `operator-guide.md`
  - `packaging-and-distribution.md`
  - `security.md`
  - `compose-compatibility.md`
  - `real-world-example.md`
- Removed duplicate docs for overlays, secrets, env files, OCI, pruning, Windows smoke tests, and hardening.
- Updated `README.md` to reference only the consolidated docs.
- Kept generated smoke-test artifacts outside the repository in examples.

## v0.9.1

- Updated docs to keep generated smoke-test packages, deployment values, env files, rendered files, and package archives outside the Dockyard repository.
- Updated Windows smoke-test commands to use `../dockyard-work` and `../deploy-values`.
- Added `../dockyard-artifacts` examples for generated `.dockyard.tgz` archives.
- Expanded `.gitignore` for common local smoke-test artifacts if they are accidentally created inside the repo.

## v0.9

- Added `--env-file` support for Compose-facing commands.
- Added `dockyard prune` for release revision cleanup.
- Stored the env-file path in release metadata without storing secret values.
- Added docs for private dotenv workflows and release pruning.
- Kept Windows `.exe` Makefile build support from v0.8.1.


## v0.8.1

- Fixed `make build` so it uses `go env GOEXE` and produces `bin/dockyard.exe` on Windows.
- Changed the default `make` target to build the Dockyard binary instead of only running `gofmt`.


## v0.8

Secrets and environment ergonomics release.

- Added `dockyard env template PACKAGE_DIR`.
- Added `dockyard env check ENV_FILE`.
- Added support for generating `.env.example`-style files from Dockyard values.
- Added `--sensitive-only` for secret-oriented environment templates.
- Added `--prefix` for generated environment variable names.
- Added duplicate-key, malformed-line, and populated-secret checks for env files.
- Added `docs/operator-guide.md`.
- Updated README and real-world examples with environment-variable secret injection guidance.

## v0.7.3

Patch release based on Windows Docker Desktop lifecycle testing.

- Made the default `dockyard init` nginx example runnable with stock `nginx`.
- Kept starter hardening practical by defaulting to `no-new-privileges` without `read_only`, `user`, or `cap_drop: [ALL]`.
- Added `docs/security.md` for stricter image-specific hardening.
- Added `docs/getting-started.md`.
- Added `status --compose-ps --all` to include stopped containers.
- Improved missing/unsupported `apiVersion` manifest errors.
- Allowed `install` to create a new revision when the existing release status is `uninstalled`.

## v0.7.2

Patch release based on local Windows smoke testing.

- Fixed JSON Schema v6 loading by decoding `values.schema.json` into a JSON value before calling `Compiler.AddResource`.
- Kept the v0.7.1 Windows path containment fix.
- Added quiet Compose validation for `render --validate-compose`, `install`, and `upgrade`.
- Kept `dockyard config` as the command that intentionally prints normalized `docker compose config` output.

## v0.7

Security and operator-safety release.

- Added `dockyard policy list`.
- Added `dockyard policy check PACKAGE_SOURCE`.
- Added `dockyard secrets scan PACKAGE_DIR`.
- Added policy checks for `read_only`, `no-new-privileges`, `cap_drop: [ALL]`, and host path mounts.
- Added `docs/security.md`.
- Added `docs/security.md`.
- Updated the README and real-world guide with policy and secret-scanning examples.
- Kept Docker Compose as the runtime source of truth.

## v0.6.2

Compose compatibility and validation release.

- Added `docs/compose-compatibility.md`.
- Added `dockyard config PACKAGE_SOURCE` to render and run `docker compose config` without installing.
- Added `dockyard render --validate-compose`.
- Updated README and the real-world guide with Compose validation examples.
- Documented the boundary between Dockyard's package layer and Docker Compose runtime behavior.

## v0.6.1

Documentation-focused cleanup.

- Simplified beginner examples to use production values files without `--overlay prod`.
- Added `docs/operator-guide.md` explaining values versus Compose overlays.
- Added `docs/getting-started.md`.
- Updated OCI and real-world examples to use overlays only in the advanced structural-override section.
- Clarified that values files choose settings while overlays change Compose structure.

## v0.6

- Added OCI registry support through the `oras` CLI.
- Added `dockyard push PACKAGE_ARCHIVE OCI_REFERENCE`.
- Added `dockyard pull OCI_REFERENCE`.
- Added install, upgrade, and diff support for `oci://` package sources.
- Added OCI source metadata in release revisions.
- Updated `dockyard doctor` to report whether `oras` is available.
- Updated real-world documentation with GHCR-style OCI examples.

## v0.5.1

- Added `dockyard values` command group.
- Added `dockyard values template PACKAGE_DIR -o values.yaml`.
- Added `dockyard values validate PACKAGE_DIR -f values.yaml`.
- Added `dockyard values schema PACKAGE_DIR`.
- Added schema-description comments and sensitive-value masking for generated values templates.
- Updated real-world documentation with the operator values workflow.

## v0.5.0

- Added `dockyard lock`.
- Added `dockyard package --locked`.
- Added package provenance metadata via `package.provenance.json`.
- Made package archives more reproducible by normalizing tar/gzip metadata.
- Added install, upgrade, and diff support for local `.dockyard.tgz` archives.
- Added `--require-lock` for install, upgrade, and diff.
- Added release-state copying of `dockyard.lock` when present.
- Added a security workflow for `govulncheck`, `staticcheck`, and Semgrep.
- Updated the real-world documentation to use lock, package, verify, and install-from-archive flows.

## v0.4.1

Documentation-focused update.

- Added `docs/real-world-example.md`.
- Linked the real-world guide from `README.md`.
- Documented a realistic `team-dashboard` + PostgreSQL package workflow.
- Covered `doctor`, `lint`, `render`, `install`, `status`, `inspect`, `list`, `package`, `verify`, `diff`, `upgrade`, `rollback`, and `uninstall`.

## v0.4

- Added `doctor`.
- Added `inspect`.
- Added `package`.
- Added `verify`.
- Added safer placeholder parsing and render diagnostics.
- Added package archive integrity checks via `SHA256SUMS`.
