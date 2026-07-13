# Threat Model

This threat model summarizes observed trust boundaries. The detailed historical review is preserved at `.ai/onboarding/reports/threat-model.md`.

## Assets

- Operator secrets in env files, shell environment, CI secrets, or private values files.
- Docker daemon access.
- Docker-compatible registry credentials.
- Package archives and catalog metadata.
- Release state.
- Release binaries, checksums, and SBOM.

## Trust Boundaries

- User input to Dockyard CLI.
- Local package directories and archives.
- Values and env files.
- Dockyard release state and temp files.
- Docker Compose subprocess execution.
- Embedded ORAS Go client registry communication.
- GitHub Actions release and security workflows.

## Observed Controls

- Path traversal checks for package paths and archive entries.
- Rejection of non-regular archive entries.
- Forbidden secret-like file checks during packaging.
- Policy checks for selected Compose risks.
- Lockfile and archive digest verification.
- Registry authentication read from Docker-compatible credential configuration.
- Restrictive file permissions are used for many generated files and release-state files where practical.
- Temporary extraction and render directories are created with `os.MkdirTemp` and removed with `defer os.RemoveAll` on observed command paths.

## Known Risks

- OCI package and catalog artifacts were not observed to have signature verification.
- Release checksums and SBOM are generated, but signing was not observed.
- Mutable tags can be used unless operators choose digest-pinned references.
- Docker daemon access is a high-trust boundary.
- The external `docker` binary is trusted from `PATH`.
- Policy and archive verification can be explicitly bypassed by flags such as `--skip-policy` and `--skip-verify`.
- Catalog metadata can be loaded from OCI, file paths, `file://` paths, or cache. Operators should trust the configured catalog source.
- Catalog cache storage is implemented under the operating system user home, not under `DOCKYARD_HOME`.

## Unknowns

- Whether official package artifacts are intended to require signatures in the future.
- Whether CI should enforce registry round-trip validation.
- Whether catalog cache behavior should move under Dockyard home.
