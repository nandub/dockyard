# Documentation Playbook

## Purpose

Keep user-facing and AI-facing documentation accurate, consolidated, and evidence-based.

## When to Use It

Use for README, `docs/*`, `.ai/*`, `AGENTS.md`, `PLANS.md`, prompts, templates, checklists, and onboarding reports.

## Required Reading

- `AGENTS.md`
- `docs/index.md`
- Relevant source, tests, Makefile, and workflows for factual claims.
- Existing docs in the same topic area.

## Preconditions

- Audience is known: user, operator, package author, maintainer, or AI agent.
- Facts are classified as Verified, Observed, Inferred, or Unknown when needed.

## Procedure

1. Determine the correct location: durable docs in `docs/`, procedures in `.ai/playbooks/`, prompts in `.ai/prompts/`, templates in `.ai/templates/`, checklists in `.ai/checklists/`, historical reports in `.ai/onboarding/`.
2. Verify commands against source, Makefile, and workflows.
3. Avoid duplicating full procedures across docs and playbooks.
4. Cross-link related docs.
5. Update `docs/index.md` when adding durable docs.
6. Preserve useful historical onboarding information under `.ai/onboarding/`.

## Validation

- Check local links and paths.
- Search for stale paths.
- No build/test run is required for documentation-only changes unless examples or generated output are being verified.

## Completion Checklist

- Documentation is in the correct location.
- Commands and paths are accurate or marked Unknown.
- No production source changed unintentionally.
- Indexes are updated.

## Escalation Conditions

- Docs contradict source or CI.
- A requested doc would require inventing behavior.
- User-facing examples require Docker, registry, or network validation.

## Required Completion Report

Report summary, files changed, docs consolidated, validation results, unverified items, and remaining risks.
