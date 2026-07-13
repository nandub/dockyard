# Onboarding Playbook

## Purpose

Discover repository structure, architecture, workflows, risks, and unknowns without changing production behavior.

## When to Use It

Use for repository discovery, onboarding reports, architecture inventories, or historical documentation refreshes.

## Required Reading

- `AGENTS.md`
- `docs/index.md`
- `.ai/onboarding/` historical reports when relevant.
- README, Makefile, workflows, and source entry points.

## Preconditions

- Git status inspected.
- Scope is discovery-only unless the user requests documentation updates.

## Procedure

1. Inventory source, tests, docs, scripts, workflows, examples, and generated/artifact locations.
2. Classify claims as Verified, Observed, Inferred, or Unknown.
3. Prefer `rg` and `rg --files` for repository searches.
4. Compare README/docs commands with Makefile and CI.
5. Record contradictions and unknowns.
6. Put historical reports under `.ai/onboarding/`.
7. Put durable technical docs under `docs/`.

## Validation

- Confirm referenced paths exist.
- Search for stale or contradicted claims.
- No build/test run is required for static discovery unless command verification is requested.

## Completion Checklist

- Repository map updated or confirmed.
- Commands documented with source evidence.
- Unknowns and risks recorded.
- No source behavior changed.

## Escalation Conditions

- Discovery requires Docker, registry, network, or credentials.
- Existing docs contradict source or CI.

## Required Completion Report

Report summary, files changed, evidence used, validation results, unverified items, and remaining risks.
