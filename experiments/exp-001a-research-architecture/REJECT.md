# REJECT

Patterns Tamvori will not reproduce. Each rejection is a complexity or correctness failure, not a complaint that the source project is bad at its own job. Those projects are successful at different jobs.

## 1. The agent as orchestrator

OpenMontage states that the agent is the intelligence and that Python does not own orchestration, review logic, or checkpoint policy. video-use's skill similarly asks the agent to follow hard rules and to verify its own output.

**Why rejected.** Instruction-following is probabilistic. A governed run needs transitions that happen if and only if the recorded event and the policy say so. If the model skips a checkpoint, the checkpoint did not exist. Tamvori puts transitions in the governor.

## 2. A worker certifying itself

Browser Use's library loop has the agent mark the task complete. OpenMontage's loop self-reviews with a reviewer skill. video-use tells the agent to verify its own output and, at the observed HEAD, adds a critic sub-agent inside that self-eval.

**Why rejected.** A producer that grades itself can move the goal, ignore a failed check, or rationalize a near miss. Claims are allowed. The accepting verdict must come from a different authority: a deterministic checker, a separate verifier, or a human. Browser Use's cloud document already describes an independent judge. That split belongs in the core, not in an optional hosted tier.

## 3. Catalog growth as the extension model

n8n's repository description cites 400+ integrations. ComfyUI's v0.38.0 changelog is largely new model and vendor nodes. OpenMontage's description cites 12 pipelines, 100+ tools, and 700+ skill files. Docling's value is a format catalog, which is appropriate for a parser and inappropriate as a template for a governor. Browser Use and OpenHands accumulate model, MCP, and tool settings.

**Why rejected.** Each addition is locally useful and globally expensive: more configuration, more failure modes, more concepts a new developer must hold. Tamvori extensions are adapters with a worker contract. The core's concept count is budgeted. A new vendor does not add a core noun.

## 4. Visual canvases and node pickers as the source of truth

n8n and ComfyUI are excellent canvases. Their execution engines are smaller than the UI required to author a graph by hand.

**Why rejected.** Tamvori is AI-built. The authoring surface that must stay simple is the contract a human and a governor can both check: objective, acceptance criteria, step list, artifacts. A canvas is a future view of that data, not the data model, and not part of EXP-001B.

## 5. YAML and Markdown as the control plane

OpenMontage stores stages, gates, and policy in manifests and skills. video-use stores a large fraction of production policy in `SKILL.md`.

**Why rejected.** Those files are good domain knowledge and bad enforcement. They have no type checker for "the agent actually did this." Policy that affects acceptance, retries, budgets, or approval is structured data evaluated by the governor. Prose remains an input artifact when a generative worker needs advice.

## 6. Framework-within-framework designs

Dify's headline architecture (surveyed, not fully dossiered) stacks an API, a job worker, a plugin daemon, a data store, and a visual builder. OpenHands at v1.24.0 spans a canvas repository, an SDK repository, a TypeScript client, and an automation service. Temporal splits the cluster into frontend, history, matching, and workers, then asks user code to use a replay-sensitive SDK.

**Why rejected.** Each layer exists to manage the previous layer's flexibility. Tamvori's first system is one Go process. Workers are functions or subprocesses. There is no plugin host, no RPC mesh, and no SDK that application steps must be written against.

## 7. Distributed systems adopted for their own sake

Temporal's cluster and database, n8n queue mode's reported Redis workers, and OpenHands warm pools all solve large concurrent scale.

**Why rejected for the core.** Tamvori's durability requirement is "a crashed local process resumes from the log," not "arbitrarily many concurrent executions across a fleet." A database, a broker, and a second service are complexity failures until a measured load requires them. The history idea is kept. The cluster is not.

## 8. Hidden mutable state

Chat memory, UI stores, node static data, and agent scratchpads can change a later step without appearing in the artifact hash.

**Why rejected.** Anything that affects the next transition is an event in the provenance log or a content-addressed artifact. Workers may use private scratch space. Scratch space that is not hashed is invisible to verification and cannot be required for acceptance.

## 9. Non-reproducible execution dressed up as recovery

Asking a model "where were we?" or re-running a generative step and hoping for the same bytes will not reconstruct a run.

**Why rejected.** Recovery replays recorded events and recorded artifacts. A new generative attempt is a new step execution, with a new artifact hash, and it needs its own verdict. Clocks and randomness inside the governor are injected and recorded.

## 10. Model-specific core

OpenHands, Browser Use, n8n, Dify, and ComfyUI all grow adapters for specific providers. That is reasonable at the edge.

**Why rejected in the core.** The governor speaks artifacts and verdicts. A provider change replaces a worker configuration. It does not change run states, policy evaluation, or provenance. No model SDK is a core dependency.

## 11. Media-specific core

OpenMontage, video-use, and ComfyUI are media systems. Their timelines, samplers, captions, and model graphs are the product.

**Why rejected in the core.** Media is the proving workload. A shot, a timeline, an FFmpeg filter, and a diffusion node are adapter concepts. If the core mentions them, the architecture has already failed the domain-independence test. Software patches, document actions, and purchase proposals must be expressible with the same eight concepts.

## 12. Giant configuration surfaces

Browser profiles, node credential forms, pipeline options, OCR and VLM switches, and model pickers are how these products become useful to many users.

**Why rejected as a default.** Configuration is a complexity metric. The first governor aims at zero required config files beyond the objective and the acceptance criteria, which are product inputs, not tuning. Optional keys need a budget and a reason.

## 13. Excessive indirection

Abstract runtimes, backend class hierarchies, node type registries, and plugin daemons make new formats and vendors possible.

**Why rejected.** Tamvori needs a handful of interfaces: store an artifact, run a worker, evaluate a policy, append an event, produce a verdict. New behavior is a new worker program, not a new abstract base class. Go interfaces are acceptable where a test double is required. They are not a layering strategy.

## 14. Features accumulated without consolidation

Release notes across OpenHands, ComfyUI, and Docling show a steady addition of surfaces. That is normal for those products.

**Why rejected as a process.** Tamvori treats an unused concept, an unused config key, an unused dependency, and an unused exported API as defects. Simplification is an improvement type. See [CONTINUOUS_IMPROVEMENT.md](CONTINUOUS_IMPROVEMENT.md) and [COMPLEXITY_BUDGET.md](COMPLEXITY_BUDGET.md).

## 15. Agents allowed to edit the definition of success

None of the studied systems were proven to lock acceptance criteria outside the worker's write path. OpenMontage puts review logic in skills the agent reads. video-use puts correctness rules in a skill the agent is asked to obey, while also telling it that artistic values are free.

**Why rejected.** If the same agent can change the rubric, the rubric does not constrain the agent. Acceptance criteria are versioned, hashed, and pinned to the run. Workers are not granted write access to them. A proposal that changes criteria is a different act, authorized by a human, and it does not rewrite the criteria of an in-flight run.

## Rejected dependencies and infrastructure, for the first implementation

These are consequences of the rejections above, stated so a later phase does not "prepare" them early:

- Workflow engines, including Temporal itself
- Web frameworks and frontend toolchains
- Databases and queue brokers
- Container orchestrators
- Plugin loaders
- Model SDKs
- FFmpeg, as a core dependency (allowed later inside a media render worker)
- Authentication platforms

A standard-library Go module is the default. A dependency requires a written reason and a budget change approved with the architecture.
