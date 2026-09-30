# EXP-001A — Research and architecture freeze

Phase: `TAMVORI-EXP-001A-RESEARCH-ARCHITECTURE-FREEZE`

Observed: 2026-09-29 (America/Denver). GitHub API timestamps in the evidence fall on 2026-09-30 UTC.

## Question

Can Tamvori learn from successful open-source systems, reject their accumulated complexity, and define a simpler Go-native architecture for governed autonomous production?

## Result

The architecture can stay domain-independent. Eight concepts are enough for the core. Media production stays in an adapter. No production code is required to answer the question.

## Read in this order

1. [RESEARCH.md](RESEARCH.md) — eight systems, with fact, inference, and design decision separated.
2. [LEARN.md](LEARN.md) — ideas worth keeping, evaluated rather than adopted wholesale.
3. [REJECT.md](REJECT.md) — complexity and patterns Tamvori will not reproduce, and why.
4. [INVENT.md](INVENT.md) — capabilities that have to be designed from Tamvori's requirements.
5. [ARCHITECTURE.md](ARCHITECTURE.md) — minimal architecture, concept justifications, Go evaluation, core versus media.
6. [DETERMINISM_BOUNDARY.md](DETERMINISM_BOUNDARY.md)
7. [VERIFICATION_MODEL.md](VERIFICATION_MODEL.md)
8. [CONTINUOUS_IMPROVEMENT.md](CONTINUOUS_IMPROVEMENT.md)
9. [COMPLEXITY_BUDGET.md](COMPLEXITY_BUDGET.md)
10. [ORIGINALITY.md](ORIGINALITY.md)
11. [NEXT_EXPERIMENT.md](NEXT_EXPERIMENT.md)
12. [evidence/SOURCES.md](evidence/SOURCES.md)

## Hypothesis under test

A continuously AI-developed system can stay simple, deterministic, auditable, and reliable when generative AI proposes and creates, deterministic software governs, workers stay inside bounded contracts, independent verification is mandatory, acceptance criteria sit outside worker control, complexity is measured, provenance is first-class, and a change must show value before acceptance.

EXP-001A defines that architecture. It does not implement it.

## Scope held

No production packages, `go.mod`, application scaffold, UI, API server, database, containers, plugin framework, model integration, or media renderer. No other repository was modified.
