# Next experiment

`TAMVORI-EXP-001B-KERNEL-SLICE`

Do not start it from this document's existence alone. EXP-001A stops at the architecture freeze. EXP-001B starts when a human authorizes implementation.

## Objective

Prove one governed run, in Go, with no third-party modules:

```text
objective
  -> pinned plan artifact
  -> generative-shaped worker (fixture)
  -> deterministic worker
  -> independent verifier
  -> artifact
  -> provenance record
```

The artifact is a structured media production package: JSON describing intended assets and their hashes. No frames, no FFmpeg, no model SDK.

## Why this slice

It exercises every core concept once. The fixture worker uses the generative contract (it may emit invalid bytes on a negative test) without making the test depend on a provider. The deterministic worker assembles the package from fixtures. The verifier is a different identity and checks schema, hashes, and the pinned criteria hash. The log can be replayed to the same projection.

## Acceptance criteria, owned outside worker code

Proposed location: `experiments/exp-001b-kernel-slice/acceptance/`. Worker packages must not be the tests' source of truth. A test fails if a worker identity writes that directory.

Minimum checks:

1. A successful fixture run ends `accepted` only after a verdict whose evaluator name differs from the producer name.
2. A verdict event that reuses the producer name does not move the step to `passed`.
3. A worker claim artifact alone does not move the step to `passed`.
4. Tampering with one log line fails hash-chain verification on replay.
5. Re-folding an untouched log yields the same projection.
6. The media package schema is parsed only in an adapter package. The core fold does not mention media fields.
7. `go.mod` has zero requirements.
8. The run resumes after a simulated crash once the output artifact event is durable, and does not invent a result that was never written.

Criteria text for EXP-001B should be copied into that experiment's acceptance file at the start of the phase and then pinned. This page is the proposal. It is not yet the pinned criteria.

## Out of scope for EXP-001B

Real LLM calls, FFmpeg, HTTP APIs, databases, authentication, UI, plugin loaders, reusable workflow editors, and a second domain. If any of those seem necessary to make the slice pass, stop and amend the experiment. Do not add them silently.

## Stop conditions

Stop if the slice needs a third-party library, if media fields enter the core fold, if the verifier must be the same process identity as the producer with no inequality check, or if the only way to pass is to weaken the acceptance file during the run.

## Later phases, not authorized

- EXP-001C: one real generative provider behind the same worker contract, still emitting a package or a plan, still no renderer.
- EXP-001D: a deterministic render worker, likely FFmpeg, as an adapter.
- A document-to-action slice that reuses the kernel with a different adapter, to test domain independence.
- Only after the kernel is boring: proposals that modify Tamvori itself under pinned tests.

## Repository note

The Tamvori directory was not a git repository at the start of EXP-001A. Initializing git and committing these documents is a human decision. See the phase report for the recommendation.
