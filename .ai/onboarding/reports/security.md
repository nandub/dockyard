# Security Analysis

This document analyzes Dockyard security posture and trust boundaries from static repository inspection. It records risks only; it does not prescribe or apply fixes.

## Summary

Dockyard wraps Docker Compose, package archives, local release state, and OCI package distribution. Its most important trust boundaries are:

- Untrusted package directories and archives.
- Untrusted OCI artifacts and catalog metadata.
- Operator-owned values and env files.
- Docker Compose execution against the host Docker daemon.
- External `oras` authentication and transport.
- GitHub Actions release artifacts.

Dockyard has meaningful protections for path containment, archive extraction, secret-like file rejection, policy checks, digest manifests, and release-state file permissions. It does not currently provide a full cryptographic supply-chain story: no OCI signature verification, no release asset signing, no image digest enforcement, and no in-process registry/TLS verification beyond what ORAS, Docker, Go, GitHub Actions, and the OS provide.

## Credential Usage

Dockyard does not appear to accept registry usernames, passwords, or tokens directly.

Observed credential-related behavior:

- OCI registry authentication is delegated to `oras`.
- Docker registry authentication is delegated to Docker and Docker Compose.
- `--env-file` can provide secrets to Docker Compose subprocesses.
- Env-file values are appended to subprocess environment and are not written into Dockyard release metadata.
- Release metadata stores the env-file path.
- Render diagnostics and generated templates attempt to mask or avoid populated secret-like values.

Evidence:

- `internal/oci/oci.go`
- `internal/runner/docker.go`
- `internal/envfile/envfile.go`
- `internal/cli/common.go`
- `docs/security.md`
- `docs/operator-guide.md`

Risks:

- A release record stores the env-file path, which may reveal deployment structure or secret file locations.
- Environment values are still visible to the subprocess environment and may be visible through OS process/environment inspection depending on platform and privileges.
- Docker Compose may resolve and expose environment-backed values according to Compose behavior outside Dockyard's control.
- Secret detection is heuristic and can miss secrets with uncommon names.

## Token Storage

No Dockyard-specific token storage was found.

Credential/token storage is external:

- ORAS stores and uses registry authentication according to ORAS behavior.
- Docker stores and uses registry authentication according to Docker configuration.
- GitHub Actions release permissions are managed by workflow permissions.

Risks:

- External credential stores are outside Dockyard validation.
- Compromised `oras`, Docker config, or CI credentials could affect package publishing or pulling.

## OCI Authentication

OCI push/pull is implemented by shelling out to `oras`.

Dockyard checks:

- `oras` exists in `PATH`.
- OCI references start with `oci://`.
- OCI references include an explicit tag or digest.
- Pulled package artifacts contain exactly one archive.

Dockyard does not directly handle:

- Registry login.
- Credential storage.
- TLS configuration.
- Registry trust roots.
- Signature verification.

Risks:

- Trust depends on the user's `oras` binary and its configuration.
- A malicious or shadowed `oras` binary earlier in `PATH` could be executed.
- A valid tag is mutable; using tags alone does not guarantee immutability.
- Digest-form OCI references are accepted, but not required for package sources generally.

## TLS and Certificate Validation

No Dockyard-specific TLS or certificate validation code was found.

TLS/certificate behavior is delegated to:

- ORAS for OCI registry communication.
- Docker/Docker Compose for image pulls and registry operations.
- GitHub Actions and GitHub release infrastructure for release asset publishing.
- Containers defined by Compose for application TLS behavior.

Repository TLS examples:

- `examples/nginx-tls-mounted-certs`
- `examples/caddy-letsencrypt`
- `examples/traefik-letsencrypt`

These model TLS with Compose; Dockyard does not manage certificates itself.

Risks:

- Dockyard cannot enforce registry TLS policy beyond ORAS/Docker behavior.
- Dockyard does not validate operator-provided TLS certificates or private key correctness.
- Mounted certificate/key paths in packages are host filesystem trust inputs.

## Filesystem Trust

Dockyard reads and writes local files extensively:

- Package directories.
- Package archives.
- Values files.
- Env files.
- Release state under Dockyard home.
- Temporary extraction directories.
- Catalog cache under `~/.dockyard/cache/catalogs`.

Protections found:

- `dockpkg.SafeJoin` rejects absolute package paths and path escapes.
- Archive extraction rejects absolute entries and parent traversal patterns.
- Archive extraction rejects non-regular tar entries.
- Archive creation rejects symlinks.
- Archive packaging rejects common secret-like and local artifact names.
- Release state and generated sensitive files are generally written with `0o600`; directories often use `0o700`.
- Archive output uses `O_EXCL` to avoid overwriting existing archives.

Risks:

- Local package directories are trusted inputs; malicious Compose files can still request dangerous host behavior.
- Some generated files can contain rendered Compose with sensitive values if values contain literal secrets.
- Catalog cache is under OS user home, not necessarily the `--home`/`DOCKYARD_HOME` state directory.
- Temporary directories are created under the OS temp location and cleaned best-effort with deferred `RemoveAll`.
- Archive extraction does not appear to limit total uncompressed size, file count, or per-file size beyond tar header behavior.

## Package Verification

Package verification exists in `internal/archive`.

Verification checks:

- Archive can be extracted safely.
- Forbidden files are absent.
- `SHA256SUMS` parses.
- File digests match `SHA256SUMS`.
- No extra files are present outside `SHA256SUMS`.
- Package provenance, when present, has the expected API version and package identity.
- Locked package provenance checks the lockfile digest.
- `Dockyard.yaml` loads and validates.
- Optional caller-provided extraction lint can run.

Risks:

- `SHA256SUMS` inside the archive is self-contained; without a signature or external trusted digest, it protects against internal inconsistency, not malicious archive authors.
- `package.provenance.json` is generated metadata, not signed provenance.
- Lockfiles verify package files, values digest, and rendered Compose digest, but do not prove publisher identity.

## OCI Signature Verification

No OCI signature verification was found.

No use of:

- Cosign.
- Notation.
- Sigstore.
- Rekor.
- ORAS signature verification.
- Public key or certificate policy.

Risks:

- A package pulled from OCI is trusted based on registry access, tag/digest selection, and archive verification only.
- Mutable tags can be replaced by an attacker with registry push access.
- There is no built-in publisher identity verification.

## SBOM Generation

SBOM generation exists in the release workflow:

```bash
syft dir:. -o spdx-json=dist/dockyard-source.spdx.json
```

Release workflow publishes:

- `dockyard-source.spdx.json`
- `SHA256SUMS`
- Cross-platform binaries.

Risks:

- SBOM is generated for source directory, not clearly for each final binary artifact.
- SBOM is not signed.
- `SHA256SUMS` is not signed.
- Local `make release-snapshot` does not generate SBOM or checksums.

## Dependency Verification

Go dependency integrity:

- Go modules use `go.sum`.
- CI runs `go mod tidy`.
- Security workflow runs `govulncheck`.
- Staticcheck and Semgrep run in workflows.

Package dependency metadata:

- Dependency source is validated for presence.
- OCI dependency sources must include an explicit tag or digest.
- Dependency names and aliases are validated.

Image dependency handling:

- Lockfiles record rendered image references and digest values only when images are already pinned with `@sha256:...`.
- Dockyard does not resolve tags to digests.
- Dockyard does not enforce digest-pinned images globally, unless package policy flags catch `latest`/untagged patterns.

Risks:

- Go module downloads rely on Go module ecosystem trust and configured module proxy/sumdb behavior.
- Staticcheck and govulncheck are invoked via `go run ...@latest`, which fetches current tool versions at runtime.
- Container image tags remain mutable unless operators use digest-pinned references.
- Dependency package sources can use tags; tags can move.

## Supply-Chain Protections

Protections found:

- `go.sum` dependency checksums.
- `go mod tidy -diff` in `make verify`.
- `govulncheck`, Staticcheck, and Semgrep workflows.
- Release binaries built in GitHub Actions from tags.
- Version metadata embedded in release binaries.
- Release checksums.
- Release SBOM.
- Package archive checksums and provenance metadata.
- Optional lockfile verification.
- Package quality checks and policy lint.

Risks:

- No signed commits/tags requirement found.
- No signed release artifacts found.
- No signed package archives found.
- No OCI signature verification found.
- No SLSA provenance workflow found.
- No pinned GitHub Action SHAs; workflows use version tags such as `actions/checkout@v4`.
- Security tools invoked with `@latest` are not version-pinned.

## Temporary File Handling

Temporary directories use `os.MkdirTemp`.

Observed temp prefixes:

- `dockyard-oci-*`
- `dockyard-src-*`
- `dockyard-dependency-values-*`
- `dockyard-catalog-*`
- `dockyard-verify-*`
- `dockyard-config-*`
- `dockyard-render-*`
- `dockyard-package-test-*`
- `dockyard-pull-*`
- `dockyard-name-*`

Protections:

- Most temp directories are cleaned with deferred `os.RemoveAll`.
- Extracted files are created with `O_EXCL` and mode `0o600`.
- Temp extraction directories are used for archives and OCI pulls.

Risks:

- Cleanup is best-effort and may leave files after crashes.
- Temp files may contain rendered Compose, extracted package contents, or inline dependency values.
- Temp location security depends on OS temp directory behavior and permissions.

## Privilege Boundaries

Dockyard runs as the invoking user.

Major privilege boundary:

- Docker daemon access. A user who can control Docker can often affect the host substantially.

Dockyard package risk:

- Compose packages can define containers, volumes, ports, images, host paths, Docker socket mounts, privileged mode, and host networking.

Protections:

- Policy checks can reject or warn on high-risk Compose patterns.
- Install refuses high-severity policy findings unless `--allow-risk` is set.
- Package quality checks include policy lint.

Risks:

- `--allow-risk` can bypass high-severity policy blocking.
- `--skip-policy` can skip policy checks.
- `--skip-compose-config` can skip Compose validation.
- Docker Compose remains capable of privileged host-affecting behavior.

## Unsafe APIs and Shell Execution

External command execution uses `exec.CommandContext`, not shell string execution.

Commands found:

- `docker`
- `docker compose ...`
- `oras`

Makefile uses shell features for build metadata:

- `git rev-parse`
- `git show`
- platform-specific release build command syntax.

Risks:

- Executable lookup uses `PATH`; malicious `docker` or `oras` binaries could be invoked.
- Docker and ORAS subprocesses inherit environment by default.
- Subprocess stderr is often passed through directly.
- Errors from Docker/ORAS are intentionally summarized, which reduces leakage but can also hide forensic detail.

## Network Boundaries

Network operations are delegated:

- ORAS performs OCI registry network I/O.
- Docker/Docker Compose perform image pulls and Docker daemon communication.
- GitHub Actions publish release assets.
- Go tooling may contact module proxy/sumdb/tool module sources during CI/security workflows.

Dockyard itself does not appear to open network sockets directly.

Risks:

- Registry trust is external.
- Catalog metadata pulled from OCI can influence package source resolution.
- Docker image pulls may fetch mutable tags.
- Go `@latest` tool invocations fetch mutable tool versions.

## Risk Register

### High

- Untrusted Compose packages can request host-affecting Docker behavior. Policy checks reduce risk but can be bypassed with flags or may not cover all Compose features.
- OCI package and catalog artifacts are not signature verified.
- Release assets and `SHA256SUMS` are not signed.
- Mutable tags are allowed for package sources, dependency sources, catalog entries, and container images.

### Medium

- Archive verification is self-contained and does not establish publisher identity.
- Temp files can contain rendered Compose and inline values; cleanup is best-effort.
- External `docker` and `oras` binaries are trusted from `PATH`.
- Catalog cache is outside configurable Dockyard home and is trusted for five minutes.
- Security tooling in workflows uses `@latest` for Go tools.
- GitHub Actions are referenced by tags rather than pinned SHAs.
- No dedicated CI registry round-trip verification was found.

### Low / Contextual

- Env-file path in release metadata may disclose local layout.
- Error summarization can hamper debugging.
- SBOM is source-level and unsigned.

## Evidence

- `SECURITY.md`
- `docs/security.md`
- `docs/release-engineering.md`
- `.github/workflows/release.yml`
- `.github/workflows/security.yml`
- `internal/archive/archive.go`
- `internal/lock/lock.go`
- `internal/oci/oci.go`
- `internal/catalog/catalog.go`
- `internal/runner/docker.go`
- `internal/envfile/envfile.go`
- `internal/policy/policy.go`
- `internal/cli/common.go`
- `internal/cli/install.go`
