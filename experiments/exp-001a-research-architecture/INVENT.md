# INVENT

Capabilities that come from Tamvori's requirements. Some combine ideas evaluated in [LEARN.md](LEARN.md). The combination, the ownership rules, and the complexity rule are the part that has to be designed here rather than imported.

This document makes no patent or copyright claim. See [ORIGINALITY.md](ORIGINALITY.md).

## A. Deterministic governance of generative work

Generative workers propose plans, text, structures, and other artifacts. They emit events. The governor is the only component that changes run state, and it does so by folding an event with the current policy. A model response is data. It is not a branch in the governor.

This is stricter than "the application calls a model inside a workflow." The model is outside the state machine.

## B. Independent verification

No worker is the sole authority that its own output is correct. The producer writes an artifact and may write a claim. A verdict is a separate record naming the artifact hash, the acceptance-criteria hash, the verifier identity, and the result. The step's producer identity and the verdict's authority identity must differ. Policy rejects a verdict that fails that check.

Deterministic checks outrank model judgment. The hierarchy is in [VERIFICATION_MODEL.md](VERIFICATION_MODEL.md).

## C. Tamper-evident provenance

The provenance log is append-only. Each event includes the hash of the previous event, the hashes of artifacts it names, and the identity of the actor. Rewriting history breaks the chain. The log is the state. Derived status (running, accepted, rejected) is a projection that a pure function can recompute.

"Tamper-evident" here means a local hash chain a later reader can check. It does not mean a blockchain, a trusted timestamp service, or multi-party signatures. Those are future options if a threat model requires them.

## D. Replay where it is technically real

Replay of the governor means: given the same log, the fold produces the same projection. That is testable and does not call a model.

Replay of a deterministic worker means: given the same input artifact hashes and the same worker version, the output hash matches. When it does not, the worker is not deterministic and must be labeled as such.

Replay of a generative worker means: return the stored artifact. A fresh model call is a new execution, even if the prompt matches. Tamvori does not claim bit-reproducible language models.

## E. Explicit human approval gates

Policy names the steps that require an approval event before release or before a side effect that leaves the run (spending money, publishing, sending). The approval event names a human identity, the artifact hash, and the criteria hash. A worker, a verifier model, and a missing field cannot fill that event in.

Approval is a kind of verdict issued by a human. It is not a second product concept.

## F. Bounded retry and recovery

Every step has a retry budget in the policy: count, and which failures are retryable. Exhaustion is a recorded failure. Recovery after a crash resumes at the last complete event. In-flight work whose output was not recorded is rerun only if the worker is deterministic or the policy explicitly allows a new generative attempt. Idempotency keys for external side effects are part of the step contract when a worker can affect the outside world.

## G. Complexity budgets

Complexity is an acceptance dimension. Counts of dependencies, packages, exported API, config keys, services, and core concepts have numeric ceilings. Crossing a ceiling fails the build the way a failing test does. Budgets are in [COMPLEXITY_BUDGET.md](COMPLEXITY_BUDGET.md).

The aim is the minimum complexity that delivers the capability, not the smallest numbers that can be written down. A budget may rise, by a human decision recorded as a criteria change, when a capability cannot be delivered inside it.

## H. Continuous simplification

The system must be able to get smaller. A simplification proposal deletes or merges code, config, dependencies, or concepts. It is accepted when the pinned acceptance suite still passes, the targeted complexity metric drops, and no protected outcome metric regresses. Adding a feature is not the only way to improve.

## I. Continuous improvement against a baseline

A candidate, including a simplification, is a proposal artifact. It is executed as its own run under the same pinned acceptance criteria as the baseline. Acceptance requires the criteria to pass and a measured benefit, or a complexity reduction with no outcome regression. The candidate does not edit the criteria. Comparison evidence is kept in the provenance of the decision.

## J. Domain-independent orchestration

The governor's nouns are objective, run, step, worker, artifact, verdict, policy, and provenance. Media, software, documents, and procurement show up as artifact schemas and workers. The state machine does not branch on domain.

## K. Swappable generative providers

A generative worker declares a provider configuration outside the core types. Swapping a model changes worker configuration and therefore the worker version recorded in provenance. It does not change policy code. Two providers can be bound to the same step contract in two runs, and their artifacts compared under the same criteria.

## L. Local deterministic workers with no model

A worker may be an ordinary program: a schema validator, a packager, a hasher, a test runner, a document converter, later an FFmpeg invocation. The governor treats it as a worker. It does not have a special "tool" type. The first vertical slice must include at least one worker that cannot be satisfied by a language model, so the architecture cannot collapse into "just call the model."

## M. Artifact-first architecture

Work products are artifacts. Plans, claims, evidence, measurements, and improvement proposals are artifacts with schemas. The run stores hashes, not a zoo of parallel object models. Bytes are the identity. Names are labels.

## N. Observable cost, time, and resource budgets

Policy carries ceilings: model tokens or money if a worker reports them, wall time, process count, artifact bytes, retry count. The governor enforces ceilings it can observe itself (time, bytes, retries, process exit). Self-reported money is a claim and is recorded, and policy may refuse another paid step when the claimed spend reaches the ceiling. A ceiling hit is an event, not a log line.

## O. Four authorities

| Authority | May do | May not do |
| --- | --- | --- |
| Proposer | Emit a proposal artifact (plan, patch, creative choice, improvement) | Change run state, write acceptance criteria, issue a verdict |
| Executor | Produce an output artifact from a contract | Grade that artifact as accepted |
| Evaluator | Issue a verdict on an artifact against a pinned criteria hash | Modify the artifact or the criteria |
| Authorizer | Issue an approval, release, or criteria change | Be the worker that produced the artifact under review |

A single human can hold authorizer and, on a different run, proposer. A single identity cannot be executor and evaluator for the same artifact. The governor enforces the separation. It is not a convention in a prompt.

## P. Acceptance criteria outside worker control

Criteria are versioned artifacts. A run pins their hash at start. Worker contracts receive them read-only. The store denies worker identities write permission to the criteria path. A criteria change is a new version created by an authorizer. In-flight runs keep the hash they pinned. Past runs are not re-graded under newer criteria unless a human starts an explicit re-grade run, which is a new run.

This is the rule that makes improvement and verification meaningful. Detail is in [CONTINUOUS_IMPROVEMENT.md](CONTINUOUS_IMPROVEMENT.md).

## Additional capability: claims are not verdicts

A worker may attach a claim artifact ("I believe this passed, here is my check"). The projection ignores claims when computing acceptance. Claims are retained because they are useful evidence and because hiding them would encourage the worker to stuff the same content into the output. The schema distinguishes `claim` from `verdict` so a later reader cannot confuse them.

## Additional capability: release is a separate transition

Producing a good artifact and releasing it are different events. Release is allowed only when required verdicts, and approval if policy demands it, are present for the pinned criteria hash. A media render, a merged patch, a sent purchase order, or a published document are release effects in adapters. The core only records whether release was authorized.

## What this phase explicitly does not invent

- A new model, a new workflow language, or a new agent prompt format
- A distributed consensus protocol
- A visual programming environment
- A media engine
- An implementation

The inventions are ownership rules and a small kernel. The next experiment tests whether that kernel can execute one run.
