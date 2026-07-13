# Open Questions

These questions come from historical onboarding reports and remain unverified in the current migration.

- Should official OCI package or catalog artifacts require signature verification?
- Should release assets and `SHA256SUMS` be signed?
- Should CI pin GitHub Actions and Go security tools to immutable versions?
- Should archive extraction enforce total uncompressed size or file-count limits?
- Should catalog cache location follow `DOCKYARD_HOME`?
- Should the Docker executable path be configurable or verified?
- Should live OCI registry round-trip validation exist in CI?
- Should coverage reporting or a coverage gate be added?
