# AI Documentation Migration Report

## Files Moved

- `.agents/workflows/*` -> `.ai/playbooks/*`
- `docs/onboarding/*` -> `.ai/onboarding/reports/*`

## Files Merged or Rewritten

- `AGENTS.md` was rewritten as the concise AI entry point.
- `PLANS.md` was rewritten as execution-plan requirements.
- `docs/index.md` was rewritten as the durable documentation entry point.
- Earlier workflow files were rewritten as `.ai/playbooks/*`.
- Onboarding discovery findings were summarized into durable `docs/*` technical documents.

## Files Removed

- `.agents/` was removed after useful workflow content was migrated.
- `docs/onboarding/` was removed after historical reports were moved to `.ai/onboarding/reports/`.

## Corrected Claims

- `.agents/workflows/` is not treated as a standard Codex feature.
- `.agents/skills/` is not treated as a standard Codex feature.
- The repository now uses `.ai/playbooks/` for reusable procedures.
- Dockyard is documented as Docker Compose-oriented.

## Remaining Issues

See `.ai/onboarding/open-questions.md` and `.ai/onboarding/known-risks.md`.
