# Known Risks

Risks preserved from onboarding discovery:

- Docker daemon access is a high-trust boundary.
- OCI package and catalog signature verification was not observed.
- Release checksums and SBOM generation were observed, but signing was not observed.
- Mutable package, catalog, and image tags can move.
- Archive checksums inside an archive do not prove publisher identity.
- The external `docker` binary is trusted from `PATH`.
- Docker, registry, and network validation is not always available locally or in CI.
- Security tooling in workflows may depend on externally resolved tool versions.

See `docs/threat-model.md` and `.ai/onboarding/reports/security.md` for supporting evidence.
