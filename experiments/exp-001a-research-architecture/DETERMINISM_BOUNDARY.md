# Determinism boundary

The boundary is a rule about state, not a catalog of features that feel deterministic.

## Rule

A transition in the governor is a pure function:

```text
projection' = fold(projection, event, policy)
```

`policy` and the prior projection are already recorded. The event is the only new input. The function does not read the clock, the network, a model, or the filesystem, except for the artifact bytes named by hashes already in the event.

Anything that needs the clock, the network, a model, randomness, or a subprocess is a worker or an external actor. It contributes events. It is not inside `fold`.

This is the part worth taking from durable-execution systems, without taking their cluster or their workflow SDK. The fold is ordinary code. It stays deterministic because the impure world is not allowed to call it with hidden inputs.

## On the deterministic side of the fold

- Whether a step may start, given the pinned plan and dependency hashes
- Retry counts and exhaustion
- Budget comparisons the governor can observe: elapsed allowance already recorded, byte sizes, event counts
- Artifact identity: hashing bytes and refusing a store when the hash does not match
- Schema checks and other predicates that are pure functions of artifact bytes plus pinned criteria
- The producer-versus-evaluator identity check
- Whether release is authorized
- Hash-chain verification
- Projection of run status

Injected time is deterministic once it is an event. The governor does not call the wall clock inside `fold`. A supervisor records `time_observed` events, or the policy uses logical counters (attempts, bytes) so that tests do not depend on the clock at all. Prefer counters until a real timeout requires a recorded timestamp.

## On the generative side

These produce artifacts and claims. They do not fold state.

- Drafting a plan
- Interpreting an ambiguous objective
- Creative choices (structure, wording, shot intent, editorial taste)
- Research and synthesis
- Proposed patches and proposed process improvements
- A critic model's opinion

A critic opinion is evidence of low rank. See [VERIFICATION_MODEL.md](VERIFICATION_MODEL.md).

## Deterministic workers are still outside the fold

A packager, a hasher, a test runner, a schema linter, and a future FFmpeg invocation are deterministic only in the empirical sense: same inputs and same program version, same output hash. They still run outside `fold`, because they touch processes and files. Their outputs enter as events. If a supposed deterministic worker changes output without a version change, provenance shows it, and the worker loses the deterministic label. The governor does not "fix" that inside the fold.

## What replay means

| Replay target | Behavior |
| --- | --- |
| Governor projection | Re-fold the log. Must match. This is a test. |
| Deterministic worker | May re-execute. Output hash must match the recorded artifact, or the mismatch is a finding. |
| Generative worker | Do not re-call the model to recover. Reuse the recorded artifact. A new call is a new attempt with a new hash. |
| Human approval | Reuse the recorded approval event. Do not ask again unless a new run is started. |

Bit-identical regeneration of model output is not a requirement and must not be faked by temperature tricks. Where a seed exists and the worker claims determinism, record the seed and test it. Where it fails, label the worker generative.

## Boundary cases, decided

| Case | Decision |
| --- | --- |
| Planning | Generative proposal. Pinning the plan hash into the run is a deterministic transition. |
| "The model decided the workflow failed" | A claim. The fold ignores it. |
| Retry | A policy decision in the fold. The re-execution is a worker. |
| Choosing a model | Worker configuration, hashed into the worker version on the event. The fold does not switch on provider names. |
| Skipping unchanged work | Deterministic: if input hashes and worker version already have a passing verdict, policy may skip. The skip is an event. |
| Random tie-breaks | Forbidden in the fold. If a worker needs a choice, the choice is in the artifact. |
| Editing acceptance criteria to make a run pass | Not a transition available to workers or to the fold. A new criteria version is an authorizer event and does not alter pinned hashes. |
| Media encode | Future deterministic worker if the binary and flags are pinned. Generative models that emit frames stay generative workers. |

## What this rejects from the obvious split

A list that says "hashing is deterministic" and "planning is generative" is true and incomplete. Hashing inside a function that also reads the clock is not a deterministic transition. Planning that directly marks a run accepted is not safely "just generative." The boundary is the fold. Creative work stays outside it. Bookkeeping stays inside it. Workers of both kinds cross the boundary only by submitting events.
