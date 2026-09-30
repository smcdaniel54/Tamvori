# Originality

Engineering provenance for EXP-001A. This is not a legal opinion, a copyright opinion, or a patent claim. It records where ideas came from so later work does not launder a port into an "original" design, and does not pretend a well-known pattern was invented here.

Tamvori can be explained without naming another project's implementation: a human pins an objective and acceptance criteria; a governor folds events; workers return hashed artifacts; a different identity issues a verdict; the log is the state; a change is accepted only if those criteria pass and the change shows a benefit or a real simplification. Media is one producer of artifacts.

## 1. Commonly known

These are part of ordinary software practice. Using them is not a distinguishing claim.

- Separating a plan from an execution
- Jobs, steps, retries, and timeouts
- Content hashes as identities
- Append-only logs
- Tests and schema validation
- Human approval before a dangerous action
- Adapters at the edges of a core
- "Do not trust user input," applied here to model output

## 2. Learned from the systems studied

Valuable, and traceable to specific observations in [RESEARCH.md](RESEARCH.md). Tamvori did not originate these.

- Agent loop versus sandbox execution (OpenHands)
- Structured page observations and closed browser actions (Browser Use)
- An independent judge of an agent trajectory, described in Browser Use's cloud product document
- Node graphs, stored executions, and human-in-the-loop pauses (n8n)
- Event-sourced replay, workflow code versus activities, recorded side effects (Temporal)
- Instruction-readable production stages, tool preflight, and cost announcement (OpenMontage)
- A model-legible proxy plus an edit decision list plus a deterministic render, and fail-closed render checks (video-use)
- Partial re-execution when inputs are unchanged, workflows stored as data (ComfyUI)
- One canonical document produced by format adapters (Docling)

## 3. Deliberately rejected

See [REJECT.md](REJECT.md). Rejection is part of the design record. In particular: agent-as-orchestrator, self-certification, integration and node catalogs in the core, visual editors as the source of truth, prose as policy, workflow clusters, and media types inside the state machine.

## 4. Independently derived requirements

These follow from the problem statement (governed autonomous production, AI-built software, mistakes that cost money, evidence that must survive the producer) rather than from a feature in one of the eight repositories.

- The producer of an artifact is forbidden, by the state machine, from being the evaluator of that artifact. Some products bolt a judge or a human on later. Here the inequality is a fold precondition for every accepting transition.
- Acceptance criteria are versioned artifacts, pinned by hash to the run, and outside the worker write set. A run cannot be made to pass by editing the rubric.
- Generative output is untrusted input to a deterministic fold. The model is not a hidden branch in the governor.
- Replay is defined only where it is true: re-fold the log; re-execute only workers that earn a deterministic label; never pretend a fresh model call is a replay.
- The first proof workload must be expressible as an adapter schema so the core stays usable for documents, software, and other expensive workflows.
- The development process that builds Tamvori is the same authority split the product enforces. That bootstrap is a requirement, not a feature copied from an agent framework.

## 5. New combinations

None of the pieces below are unique alone. The combination is the architecture:

1. Temporal-style history as a local file fold, without a cluster and without a replay SDK.
2. Agent-style workers that may be generative, treated as untrusted activity executors.
3. A verification rank order that refuses to let a critic model outvote a deterministic predicate.
4. Content-addressed skip, taken as an idea from media graphs, applied to any artifact.
5. Domain catalogs forced out into adapters, with a numeric complexity budget that can fail the build.
6. Improvement proposals, including deletions, competing against a baseline under criteria they cannot edit.

An architecture review that says "this is a workflow engine plus agents plus evals" is describing the ingredients. The design constraint is which ingredients were refused and who is allowed to decide success.

## 6. Potentially novel capabilities

"Potentially" means: this research did not find the same rule as a kernel invariant in the eight systems. It does not mean a search of all prior art.

- Complexity ceilings are acceptance criteria, and a proposal whose only benefit is a lower ceiling can win against a feature proposal.
- Simplification proposals are a required improvement class, not a cleanup someone might do later.
- Claims and verdicts are different event types, and the projection is specified to ignore claims.
- Criteria changes are new hashes that do not re-pin in-flight or historical runs.
- Core concept count is itself budgeted, so a new noun requires a recorded human decision.

If a later phase finds a system that already does one of these, record it in the research log and narrow the claim. Do not expand the claim.

## How to explain Tamvori without a reference implementation

Use section 4 and the eight concepts in [ARCHITECTURE.md](ARCHITECTURE.md). Name studied systems only when discussing provenance of an idea, as this file does. Do not describe the product as a Go port, a rewrite, or a thinner version of any dossier in the research log.
