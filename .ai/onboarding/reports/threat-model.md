# Threat Model

This document describes Dockyard threat boundaries, assets, actors, and risks discovered during static inspection. It does not apply fixes.

## Scope

In scope:

- Dockyard CLI runtime.
- Local package directories.
- `.dockyard.tgz` archives.
- OCI package and catalog references.
- Release state under Dockyard home.
- Docker/Docker Compose subprocess execution.
- ORAS subprocess execution.
- GitHub Actions release and security workflows.

Out of scope or delegated:

- Docker daemon security model.
- ORAS credential storage and TLS implementation.
- Registry authorization internals.
- Container image publisher security.
- Application TLS behavior inside containers.
- Operating system temp directory security.

## Assets

Important assets:

- Operator secrets in env files, shell environment, CI secret stores, or private values files.
- Docker daemon access.
- Docker registry credentials.
- ORAS registry credentials.
- Dockyard release metadata.
- Package archives.
- Catalog metadata.
- Release binaries.
- Release checksums and SBOM.
- Source repository integrity.
- Package lockfiles and provenance metadata.

## Trust Boundaries

### User to Dockyard CLI

The user invokes Dockyard with flags and paths.

Boundary inputs:

- Package source path.
- Values file path.
- Env-file path.
- Release name.
- Home directory.
- OCI/catalog references.
- Risk-bypass flags.

Primary risks:

- Dangerous package source.
- Wrong env-file or values file.
- Accidental use of bypass flags.

### Dockyard to Local Filesystem

Dockyard reads and writes package files, archives, release state, temp files, and cache files.

Trust concerns:

- Path traversal.
- Symlinks.
- Secret-like files.
- Existing file overwrite.
- Temp data leakage.

Protections:

- `SafeJoin`.
- Archive path checks.
- Regular-file-only archive extraction.
- Secret-like file rejection.
- Restrictive file modes where practical.
- `O_EXCL` for archive output and extraction.

Residual risks:

- No archive decompression size limits found.
- Temp data can remain after process crash.
- Local package directories can still contain semantically dangerous Compose definitions.

### Dockyard to Docker Compose

Dockyard delegates runtime behavior to Docker Compose.

Trust concerns:

- Containers can affect the host through Docker features.
- Host-path mounts, Docker socket mounts, privileged mode, and host networking are high-risk.
- Docker Compose interprets the final rendered YAML.

Protections:

- Policy checks.
- High-severity install block unless `--allow-risk`.
- Optional `docker compose config` validation.

Residual risks:

- Policy checks are not a full Compose security model.
- `--skip-policy`, `--allow-risk`, and `--skip-compose-config` can reduce protections.
- Docker daemon access itself is powerful.

### Dockyard to ORAS and OCI Registry

Dockyard delegates OCI network operations to ORAS.

Trust concerns:

- Registry authentication.
- Registry TLS.
- Mutable tags.
- Artifact substitution.
- Catalog metadata manipulation.

Protections:

- OCI references must include tag or digest.
- ORAS must exist in `PATH`.
- Package archive verification after pull.
- Catalog API version validation.
- Catalog cache TTL of five minutes.

Residual risks:

- No signature verification.
- No publisher identity verification.
- Tag references can move.
- Digest references are accepted but not required everywhere.
- `oras` binary and credential state are trusted.

### Dockyard to GitHub Actions

GitHub Actions builds and publishes releases.

Trust concerns:

- Workflow permissions.
- Action supply chain.
- Release asset integrity.
- Tool downloads.

Protections:

- Release workflow uses `contents: write` only for release publishing.
- Release workflow generates `SHA256SUMS`.
- Release workflow generates an SPDX SBOM with Syft.
- Security workflow runs govulncheck, Staticcheck, and Semgrep.

Residual risks:

- Release artifacts are checksummed but not signed.
- SBOM is unsigned.
- GitHub Actions are tag-pinned, not SHA-pinned.
- Go security/staticcheck tools are fetched with `@latest`.

## Actors

### Legitimate Operator

Capabilities:

- Runs Dockyard.
- Chooses packages and values.
- Provides env files.
- Has Docker daemon access.

Risks:

- Accidental deployment of dangerous Compose configuration.
- Accidental secret inclusion in package values or archives.
- Using mutable tags unintentionally.

### Malicious Package Author

Capabilities:

- Provides a Dockyard package directory or archive.
- Controls Compose YAML, values schema, defaults, and metadata.

Possible attacks:

- Request privileged containers.
- Mount host paths or Docker socket.
- Use malicious images.
- Attempt archive path traversal.
- Hide secrets or generated artifacts.
- Abuse Compose features outside Dockyard's policy coverage.

Protections:

- Archive/path checks.
- Forbidden file checks.
- Policy checks.
- Package quality checks.
- Compose config validation.

Residual risks:

- Review is still required; Dockyard should not be used to run untrusted packages blindly.

### Compromised Registry or Publisher Account

Capabilities:

- Replace mutable tags.
- Serve malicious catalog metadata.
- Serve malicious package artifacts.

Possible attacks:

- Redirect catalog package names to malicious OCI sources.
- Replace package tags.
- Publish malicious archives with internally consistent checksums.

Protections:

- Explicit tag/digest syntax required for OCI references.
- Package archive verification.
- Catalog API version validation.

Residual risks:

- No signature verification.
- No trusted publisher policy.
- No mandatory digest pinning.

### Local Attacker

Capabilities:

- May manipulate PATH, temp directories, package files, env files, or Docker/ORAS credentials depending on local privileges.

Possible attacks:

- Shadow `docker` or `oras` binary.
- Modify local package source before packaging/install.
- Read temp files or env files if permissions allow.
- Tamper with catalog cache or release state if filesystem permissions allow.

Protections:

- User-owned state directories and restrictive modes.
- Temp directories created by OS temp APIs.
- File digest checks when lockfiles or archives are used.

Residual risks:

- Same-user local attackers remain powerful.
- PATH trust is broad.

## Data Flow: Install From Local Package

1. User runs `dockyard install`.
2. Dockyard resolves home.
3. Dockyard loads package manifest.
4. Dockyard loads and merges values.
5. Dockyard validates schema.
6. Dockyard renders Compose.
7. Dockyard optionally verifies lockfile.
8. Dockyard runs policy checks.
9. Dockyard writes pending release state.
10. Dockyard optionally runs `docker compose config`.
11. Dockyard runs `docker compose up -d`.
12. Dockyard marks release deployed or failed.

Trust points:

- Package source.
- Values/env input.
- Docker daemon.
- Compose interpretation.
- Release state filesystem.

## Data Flow: Install From OCI Package

1. User supplies `oci://...` or catalog shorthand.
2. Dockyard resolves catalog source if needed.
3. Dockyard runs `oras pull`.
4. Dockyard locates pulled archive.
5. Dockyard verifies/extracts archive.
6. Dockyard follows normal package install flow.

Trust points:

- Catalog metadata.
- ORAS binary.
- Registry account and transport.
- Pulled archive.
- Docker daemon.

## Data Flow: Package and Publish

1. User runs `dockyard lock`.
2. Dockyard records values/render/file/image/dependency digests.
3. User runs `dockyard package --locked`.
4. Dockyard verifies lock when requested.
5. Dockyard creates archive with provenance and `SHA256SUMS`.
6. User runs `dockyard push`.
7. Dockyard runs `oras push`.

Trust points:

- Local package source.
- Lockfile correctness.
- Archive integrity.
- ORAS credentials.
- Registry write permissions.

## Risk Register

### R1: Unsigned OCI Packages

Risk:

- Dockyard verifies archive consistency but not signer identity.

Impact:

- Malicious registry content can be accepted if it passes internal archive checks.

Evidence:

- No signature verification code or Cosign/Notation usage found.

### R2: Mutable Tags

Risk:

- Package, dependency, catalog, and image references can use mutable tags.

Impact:

- Repeat installs may fetch different content.

Evidence:

- OCI references require tag or digest; digest is not mandatory.
- Lockfiles record image digests only when already pinned.

### R3: Docker Daemon Privilege

Risk:

- Docker daemon access is a powerful host boundary.

Impact:

- Malicious Compose can affect host files, networking, and processes.

Evidence:

- Runtime uses `docker compose up/down`.
- Security docs warn against untrusted packages.

### R4: Policy Bypass Flags

Risk:

- Users can pass `--allow-risk`, `--skip-policy`, or `--skip-compose-config`.

Impact:

- Dangerous Compose may be deployed.

Evidence:

- Install/upgrade/package test flags.

### R5: External Tool Trust

Risk:

- Dockyard executes `docker` and `oras` from `PATH`.

Impact:

- PATH hijacking or compromised binaries can affect runtime.

Evidence:

- `exec.CommandContext(ctx, "docker", ...)`
- `exec.CommandContext(ctx, "oras", ...)`

### R6: Self-Contained Archive Checksums

Risk:

- Archive `SHA256SUMS` detects tampering relative to the archive manifest, but a malicious package author can generate valid checksums.

Impact:

- Verification can pass for malicious content.

Evidence:

- `SHA256SUMS` is included inside archive and checked during verification.

### R7: Temporary Data Exposure

Risk:

- Temporary directories can hold extracted packages, rendered Compose, pulled archives, and inline dependency values.

Impact:

- Same-host users or post-crash cleanup failures may expose data.

Evidence:

- Multiple `os.MkdirTemp` flows and best-effort deferred cleanup.

### R8: Unsigned Release Assets

Risk:

- Release checksums and SBOM are generated but not signed.

Impact:

- Users can detect accidental corruption if they trust `SHA256SUMS`, but not independently verify publisher identity.

Evidence:

- Release workflow generates `SHA256SUMS` and SBOM; no signing step found.

### R9: Tooling Supply Chain Drift

Risk:

- CI invokes some Go tools with `@latest` and actions by tag.

Impact:

- Tool behavior can change unexpectedly.

Evidence:

- Security and release workflows use `go run ...@latest`.
- Actions use version tags.

## Existing Security Controls

Controls found:

- Archive path traversal rejection.
- Archive symlink rejection.
- Archive non-regular-entry rejection.
- Forbidden secret-like file names.
- Package quality checks.
- Policy checks for privileged mode, host networking, Docker socket mounts, host paths, missing hardening, and latest tags.
- Secret-like values scan.
- Sensitive render diagnostics masking.
- Env files passed only to subprocesses.
- Release metadata avoids secret env values.
- Restrictive file modes.
- Lockfile digest checks.
- Package archive checksums.
- Release checksums.
- Release SBOM.
- govulncheck, Staticcheck, and Semgrep workflows.

## Explicit Non-Goals or Delegations

Observed delegations:

- Dockyard is not a secret manager.
- Docker Compose remains runtime source of truth.
- ORAS handles OCI authentication.
- Docker handles image pulls and Docker authentication.
- TLS behavior is delegated to ORAS, Docker, registries, and containers.

## Open Questions

- Should package installs require digest-pinned OCI references for high-assurance workflows?
- Should official package/catalog artifacts be signed?
- Should release assets and checksums be signed?
- Should CI pin GitHub Actions and Go security tools to immutable versions?
- Should archive extraction enforce total size and file count limits?
- Should catalog cache respect `DOCKYARD_HOME` or remain under user home?
- Should Docker/ORAS executable paths be configurable or verified?

## Evidence

- `SECURITY.md`
- `docs/security.md`
- `docs/compose-compatibility.md`
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
- `internal/quality/quality.go`
- `internal/cli/install.go`
- `internal/cli/common.go`
