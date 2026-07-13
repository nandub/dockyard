# Security Review Playbook

## Purpose

Analyze security posture, trust boundaries, and risks without applying fixes unless requested.

## When to Use It

Use for trust-boundary reviews, credential handling, archive safety, OCI supply-chain questions, policy checks, temporary file handling, and CI security posture.

## Required Reading

- `AGENTS.md`
- `SECURITY.md`
- `docs/security.md`
- `docs/threat-model.md`
- `.github/workflows/security.yml`
- Relevant source packages.

## Preconditions

- Scope is clear.
- Review-only versus remediation is explicit.

## Procedure

1. Identify assets and trust boundaries.
2. Separate Dockyard-owned controls from Docker, Compose, ORAS, registry, OS, and CI responsibilities.
3. Inspect credential, token, TLS, filesystem, archive, subprocess, and network behavior.
4. Rate risks with source or workflow evidence.
5. Mark unknowns instead of guessing.
6. Do not edit code in review-only mode.

## Validation

- Static source/workflow inspection.
- Optional existing security workflow commands only when requested and available.

## Completion Checklist

- Trust boundaries listed.
- Risks include evidence.
- Delegated responsibilities are explicit.
- No fixes applied during review-only work.

## Escalation Conditions

- Finding requires immediate secret rotation.
- Verification needs registry, Docker, or network access.
- Review discovers source/doc contradictions.

## Required Completion Report

Report summary, files reviewed, findings, validation performed, unverified items, and remaining risks.
