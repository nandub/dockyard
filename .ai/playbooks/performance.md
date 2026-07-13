# Performance Playbook

## Purpose

Investigate or improve Dockyard performance without weakening validation or correctness.

## When to Use It

Use for slow commands, package rendering, archive operations, schema validation, catalog/OCI paths, or subprocess overhead.

## Required Reading

- `AGENTS.md`
- `docs/runtime.md`
- `docs/repository-map.md`
- Relevant source and tests.

## Preconditions

- Slow path and reproduction command are known.
- Baseline measurement is available or can be collected.

## Procedure

1. Measure current behavior.
2. Separate Dockyard time from Docker, Compose, ORAS, registry, and network time.
3. Inspect repeated filesystem, archive, render, validation, or subprocess work.
4. Prefer targeted changes over caches that weaken trust or correctness.
5. Add benchmark or regression tests when practical.

## Validation

- Baseline and after timing.
- Focused package tests.
- `go test ./...`
- Benchmarks with `go test -bench` when benchmarks are added or exist.

## Completion Checklist

- Baseline is recorded.
- Improvement source is explained.
- Correctness validation still passes.
- External bottlenecks are separated from Dockyard bottlenecks.

## Escalation Conditions

- Measurement needs Docker, registry, or network access.
- Optimization would skip validation or security checks.

## Required Completion Report

Report summary, files changed, design decisions, tests changed, validation results, unverified items, and remaining risks.
