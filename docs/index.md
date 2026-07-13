# Documentation Index

This is the entry point for Dockyard technical documentation.

## Where to Begin

- `README.md` - user-facing project overview.
- `docs/getting-started.md` - first-run guide.
- `docs/operator-guide.md` - operator workflows.
- `docs/command-reference.md` - command reference.
- `AGENTS.md` - AI agent entry point.
- `.ai/README.md` - AI operating documentation layout.

## Technical Architecture

- `docs/repository-map.md` - repository structure and major files.
- `docs/architecture.md` - high-level architecture and package boundaries.
- `docs/domain-model.md` - Dockyard domain concepts.
- `docs/runtime.md` - startup, command dispatch, package loading, OCI, and shutdown model.
- `docs/configuration.md` - configuration inputs and precedence.
- `docs/package-lifecycle.md` - package lifecycle from source package to deployment.
- `docs/cli.md` - Cobra hierarchy, commands, flags, and output notes.
- `docs/oci.md` - OCI package, catalog, ORAS, and registry model.

## Development and Validation

- `docs/development.md` - local development workflow.
- `docs/build.md` - build, format, test, package, and release commands.
- `docs/testing.md` - testing strategy and gaps.
- `docs/validation.md` - recommended validation pipeline.
- `docs/glossary.md` - repository, architecture, command, and terminology glossary.

## Security

- `SECURITY.md` - vulnerability reporting.
- `docs/security.md` - user-facing security guidance.
- `docs/threat-model.md` - trust boundaries, assets, and risks.
- `.ai/checklists/security-review.md` - concise security review checklist.

## Packaging and Operations

- `docs/packaging-and-distribution.md` - package authoring and distribution.
- `docs/compose-compatibility.md` - Compose compatibility notes.
- `docs/real-world-example.md` - larger example walkthrough.
- `docs/support-policy.md` - support policy.
- `docs/upgrade-policy.md` - upgrade policy.

## Release Documentation

- `CHANGELOG.md` - release history.
- `docs/release-candidate-checklist.md` - release candidate checklist.
- `docs/release-engineering.md` - release engineering notes.
- `docs/v1-readiness.md` - v1 readiness notes.

## AI Operating Material

- `.ai/playbooks/` - reusable development procedures.
- `.ai/prompts/` - ready-to-use prompts.
- `.ai/templates/` - execution-plan and decision-log templates.
- `.ai/checklists/` - concise verification checklists.
- `.ai/onboarding/` - historical onboarding reports, migration notes, known risks, and open questions.

Historical discovery reports live in `.ai/onboarding/reports/`. Treat them as evidence notes, not current product documentation. Verify drift-prone claims against source, tests, Makefile, and CI before changing behavior.

## Documentation Maintenance

- Keep durable technical docs in `docs/`.
- Keep AI procedures, prompts, templates, checklists, and historical reports under `.ai/`.
- When command behavior changes, update `README.md`, `docs/command-reference.md`, related docs, and `CHANGELOG.md` when release-facing.
- When file formats change, update `internal/format`, `docs/v1-readiness.md`, and compatibility docs where applicable.
- Do not duplicate full procedures from `.ai/playbooks/` into technical docs.
