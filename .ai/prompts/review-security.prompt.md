# Prompt: Review Security

Read `AGENTS.md`, then use `.ai/playbooks/security-review.md`.

Perform a review-only security analysis unless remediation is explicitly requested. Inspect source, tests, workflows, `SECURITY.md`, `docs/security.md`, and `docs/threat-model.md`.

Identify trust boundaries, credential handling, token storage, OCI authentication, TLS delegation, filesystem trust, package verification, signature verification, SBOM generation, dependency verification, supply-chain protections, temporary files, privilege boundaries, shell execution, and network boundaries.

Classify findings as Verified, Observed, Inferred, or Unknown. Return risks with evidence and do not apply fixes.
