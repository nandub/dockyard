# PLANS.md

This file defines when Dockyard work requires an execution plan and what that plan must contain.

Plans are working documents for complex changes. They should be updated as facts change.

## When a Plan Is Required

Create an execution plan before editing when work involves any of the following:

- Cross-component changes.
- Public API or file-format changes.
- CLI contract changes, including flags, arguments, output, exit behavior, or JSON shape.
- Persistent state changes under Dockyard home.
- OCI layout, archive, provenance, lockfile, catalog, or registry behavior.
- Security-sensitive changes.
- Protocol or configuration changes.
- Migration work.
- Multi-stage implementation.
- Substantial refactoring.
- Release preparation or publishing.
- Any change where validation will require more than one command family.

For small documentation-only edits or narrow typo fixes, a brief note in the final response is enough.

## Required Plan Sections

Use these sections:

- Objective
- Current Behavior
- Scope
- Constraints
- Proposed Design
- Work Breakdown
- Validation
- Risks
- Progress
- Decisions
- Completion Summary

The reusable template lives at `.ai/templates/execution-plan.md`.

## Planning Rules

- Read `AGENTS.md` first.
- Use `.ai/playbooks/` for procedure selection.
- Use source, tests, CI, and scripts as evidence before historical onboarding notes.
- Mark facts as Verified, Observed, Inferred, or Unknown when accuracy matters.
- Keep the plan scoped to the user request.
- Do not hide contradictions between docs, source, tests, and CI.
- Update progress as steps are completed.
- Record design decisions that affect future maintainers.

## Completion

When the work is complete, summarize:

- What changed.
- What did not change.
- Validation commands and results.
- Any skipped validation and why.
- Remaining unknowns or risks.
