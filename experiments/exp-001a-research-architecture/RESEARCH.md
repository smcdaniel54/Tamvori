# Research dossiers

Eight systems. Four categories. Open-source projects are research inputs. Nothing here is a specification to port.

Categories:

- A. Autonomous software or agent execution: OpenHands, Browser Use.
- B. Workflow and orchestration: n8n, Temporal. Temporal is the stronger reference for deterministic durable execution. n8n is the stronger reference for integration-graph products and the complexity that follows.
- C. Agentic media production: OpenMontage, video-use, ComfyUI.
- D. Structured documents: Docling.

Each dossier answers the ten research questions. Claims are tagged **FACT**, **INFERENCE**, or **TAMVORI DESIGN DECISION**.

---

## 1. OpenHands

Repository: https://github.com/OpenHands/OpenHands
Observed release: v1.24.0 (2026-09-25), commit `7dc6805406ea3c76cb4a3ce407c3c72d481b0ac6`.
License: GitHub metadata MIT. Project README text also states that an `enterprise/` directory is excluded from MIT.

### Problem

Automate software engineering work: read a codebase, edit it, run commands, and produce a reviewable result.

### Fundamental idea

Separate the agent loop from the place where actions actually run. A sandbox executes bash, file, and browser actions. An event stream is the conversation between the model and those actions.

**FACT.** Tag 1.6.0 `system-architecture.md` names Server, Runtime, Action Execution Server, and EventStream, and lists Docker, remote, and Modal runtimes. The same path was absent on tag v1.24.0.

**FACT.** Current `main` documentation describes Agent Canvas in this repository, the agent server and SDK in `OpenHands/software-agent-sdk`, a TypeScript client repository, and an automation repository. The automation service decides when work runs. The agent server decides what runs inside a conversation.

**INFERENCE.** Between 1.6.0 and v1.24.0 the product split from one repository's server-plus-runtime into a multi-repository control plane (canvas, automation, SDK, client) plus a sandbox. That split is a scaling and product move. It is also a complexity move.

### What it does exceptionally well

The boundary between "the model decided" and "the sandbox did it" is explicit. Events are a log a person can inspect. Multiple runtimes mean the same agent contract can sit on a laptop or a remote sandbox.

### Learn

- Agent and execution environment are different components.
- An append-only event record is more honest than a hidden chat transcript.
- Scheduling work and performing work can be different authorities. The automation repository versus the agent server is a public example.

### Do not reproduce

- A desktop control center plus a separate SDK, client, automation service, and enterprise tree as the minimum viable system.
- Skills, hooks, MCP surfaces, and model-profile caches as core concepts. v1.24.0 release notes show product work on skills, hooks, tools dialogs, MCP OAuth, and LLM profiles. Those are product surfaces, not the kernel of governed production.
- Sandbox fleets, warm pools, and direct-to-sandbox WebSockets. Useful at their scale. Unnecessary for a single-process governor.

### Where complexity accumulated

**INFERENCE.** The 1.6.0 document already had three deployment shapes. v1.24.0 adds a GUI application shipped as deb, dmg, exe, and AppImage, and contributor docs that route changes across four repositories. Complexity accumulated in product surface and repository boundaries, not only in the agent loop.

### Public interface

A person meets a canvas, conversations, settings, automations, skills, and tools. The architectural idea (evented agent in a sandbox) is smaller than the interface required to operate the product.

### Clean sheet

One process owns run state. A worker is a contract that returns an artifact. The event log is the provenance log. Sandboxes are an adapter for untrusted code, introduced when a worker needs them, not as the platform.

### Go

**INFERENCE.** OpenHands' advantage is Python's agent ecosystem, not its orchestration model. Tamvori's governor does not need that ecosystem. A Go process can supervise a worker subprocess, hash artifacts, and append a log with the standard library. The model client, if any, is a small adapter.

### Generalizes

The agent/sandbox split generalizes to any worker that must not be trusted with the governor's state: a browser, a document parser, a renderer, or a procurement tool. The software-engineering product does not generalize into the core.

---

## 2. Browser Use

Repository: https://github.com/browser-use/browser-use
Observed release: 0.13.10 (2026-09-04), commit `5c892e013a73e6622e6f50336e1eb0aa2c4405f2`.
License: MIT.

### Problem

Let a language model complete a task in a real browser.

### Fundamental idea

An agent loop observes a structured page (DOM and accessibility information, with screenshots available) and calls a small tool set over the Chrome DevTools Protocol. The browser session is infrastructure. The model is the policy for the next click.

**FACT.** Project `AGENTS.md` describes that loop: task in, Chromium via CDP, repeated model calls, agent marks completion. `CLAUDE.md` names Agent, BrowserSession, Tools, DomService, and an LLM adapter over several providers.

**FACT.** `CLOUD.md` describes a hosted product in which, after the agent marks completion, an independent strict judge examines the trajectory and returns true or false.

### What it does exceptionally well

It makes a messy environment legible. The model sees a structured observation instead of a raw browser. Provider adapters sit behind one call shape. The cloud writeup states an independent judge, which is the right authority split even though it is described as a product behavior rather than as the library's completion rule.

### Learn

- Observations should be structured artifacts, not raw environment dumps.
- Tool contracts can be small and closed.
- Completion claimed by the worker and completion judged by someone else are different events. The cloud document already separates them.

### Do not reproduce

- An open-ended step loop whose success signal is the same agent that chose the actions. The library guide says the agent marks the task complete.
- Per-step model calls as the only way to make progress. A deterministic worker should do the steps whose result is a function of inputs.
- Browser-specific session, DOM, and stealth machinery in the core. A browser worker is an adapter.

### Where complexity accumulated

**INFERENCE.** The project grew from a library loop into cloud browsers, MCP server and client modes, many model adapters, and a large settings surface. Release 0.13.10's notes are about pinning dependencies and MCP error handling, which is the maintenance cost of that surface. Each integration is locally reasonable. Together they make the framework the product.

### Public interface

Library users configure a browser profile, an LLM, tools, and output schemas. Hosted users submit a prompt. The conceptual task is one sentence. The operational surface is the browser plus the model plus the agent's own stopping rule.

### Clean sheet

A browser worker accepts a bounded contract: start URL, allowed actions, output schema, step budget. It returns an artifact (extracted data, screenshot hashes, final URL). A verifier outside the worker checks the artifact against criteria pinned before the run. The governor never imports CDP.

### Go

**INFERENCE.** CDP clients exist in Go, but Tamvori should not own a browser stack in the core module. Go's advantage is supervising that worker as a process with a timeout, a byte budget, and a captured stdout artifact. The browser toolkit can stay a separate program.

### Generalizes

Structured observation plus a closed action set is the pattern for any environment: a document, a ticket system, a machine, a timeline. The independent judge generalizes directly. The browser does not.

---

## 3. n8n

Repository: https://github.com/n8n-io/n8n
Observed release: `n8n@2.41.3` (2026-09-25), target `release/2.41.3`.
License: Sustainable Use License for non-`.ee` code (`LICENSE.md` on `master`). GitHub SPDX `NOASSERTION`.

### Problem

Connect many systems into a visually authored workflow, including AI steps, and run those workflows on a schedule, a webhook, or a manual start.

### Fundamental idea

A workflow is a graph of nodes. Data flows along edges. A node is both the integration and the step. Execution walks the graph and records run data.

**FACT.** Repository description: fair-code workflow automation with native AI capabilities and "400+ integrations." Topics include workflow automation, low-code, and MCP. A rendered README excerpt also advertises human approvals and model switching.

**INFERENCE.** Public architecture summaries describe a frontend, REST API, workflow engine, node registry, credential store, and an optional queue mode that adds Redis and worker processes. That queue-mode detail comes from secondary architecture writeups, not from a license or release note read in full here. Treat the distributed mode as a reported shape, not as a line-by-line confirmation.

### What it does exceptionally well

A graph makes control flow visible. Human approval can be a node rather than an afterthought. Credentials and integrations are isolated enough that a workflow author does not reimplement HTTP for each vendor. Partial execution and pinned data make debugging a graph practical.

### Learn

- Explicit steps and explicit data passing.
- A human approval can be a first-class pause in the run.
- Execution history is part of the product, not a log someone might grep.

### Do not reproduce

- One node type per vendor. Hundreds of integrations inside the core product is how a workflow tool becomes an integration company.
- Visual authoring as the source of truth. Graphs are useful for explanation. They are a poor kernel for an AI-built system whose workflows should be code, reviewed like code, and tested like code.
- A credentials-and-nodes platform, queue brokers, and a UI as prerequisites for the first run.

### Where complexity accumulated

**INFERENCE.** The node registry is the accumulation point. Each integration adds configuration, credentials, versioning, and failure modes. Queue mode then adds a second execution topology to keep that catalog available at scale. The architecture grows by addition of nodes, not by consolidation of concepts.

### Public interface

The canvas, node picker, credential forms, and per-node option panels are the product. The execution engine underneath is smaller than the configuration required to use it.

### Clean sheet

A step has an input artifact, a worker name, and an output artifact. Integrations live in workers, behind a contract, and only the workers a domain needs are present. There is no node marketplace in the core. Approval is a policy state, not a special node package.

### Go

**INFERENCE.** n8n's TypeScript and UI stack fit a node product. They do not fit a minimum governor. Go can represent a step graph as data and execute it in one process. It should not grow an integration framework to match n8n's catalog.

### Generalizes

Graph-shaped work, pauses for approval, and stored executions generalize to every domain Tamvori might enter. The integration catalog does not. It is the domain, repeated hundreds of times.

---

## 4. Temporal

Repository: https://github.com/temporalio/temporal
Observed release: v1.32.0 (2026-09-11). Language: Go. License: MIT.
SDK note: `go.temporal.io/sdk` was listed as v1.49.0 on pkg.go.dev during this research. That is a different module from the server release.

### Problem

Run business logic that must continue correctly across process crashes, deploys, and retries, for a very large number of concurrent executions.

### Fundamental idea

Event sourcing. Each workflow execution has an append-only history. Workflow code is deterministic and side-effect free. Activities perform the side effects and must be idempotent or explicitly non-retryable. After a failure, a worker replays history. Commands must match. Recorded activity results are reused.

**FACT.** `docs/architecture/README.md` at v1.32.0 states those premises. The cluster is split into frontend, history, matching, and internal workers. User workers poll task queues. History shards own workflow executions. The system is designed to scale to arbitrarily many concurrent executions. Users may operate the cluster and its database, or use Temporal Cloud.

**FACT.** Public SDK documentation states that workflow code must produce the same commands given the same input, and that a random branch inside workflow code breaks replay.

### What it does exceptionally well

It takes determinism seriously enough to make replay a test. It separates orchestration code from side effects. Recovery is a function of history, not of a model remembering where it left off. Time, randomness, and I/O have explicit substitutes inside workflow code.

### Learn

- Append-only history is the state. Current status is a projection.
- Orchestration must be deterministic. Side effects live elsewhere.
- Replay tests catch non-determinism before production does.
- Activities have retry semantics that are policy, not hope.
- A recorded result is what replay returns. The activity does not have to run again for the workflow to recover.

### Do not reproduce

- A multi-service cluster and a database as the smallest durable executor.
- Workflow-as-framework, where application code must be written against a replay-sensitive SDK (`workflow.Sleep`, `workflow.Now`, and a determinism checker). That is justified when the user code *is* the workflow and must survive years of deploys. Tamvori's governor can be ordinary Go if it only appends events and folds them.
- Task-queue polling, shards, and matching services. Those exist to scale to huge concurrency. Tamvori's first runs are few and local.
- The assumption that all workflow code is trusted developer code. Generative workers are untrusted proposers. Temporal activities are still "our code." Tamvori needs a verdict the worker cannot issue.

### Where complexity accumulated

**FACT.** The architecture document's own requirement is arbitrary scale, and the design answer is a cluster with sharded history and a matching service.

**INFERENCE.** Most of Temporal's complexity is the distributed system required by that scale requirement, plus the SDK discipline required to keep user code replayable. Both are real. Neither is the minimum complexity for a single-machine governed run.

### Public interface

Developers learn workflows, activities, workers, task queues, signals, queries, retries, and versioning. The concepts are coherent. The operational surface (cluster, database, workers) is large before the first workflow does anything.

### Clean sheet

Keep history-as-state and the split between deterministic transitions and side-effecting work. Implement history as an append-only file. Implement the fold as a pure function in the governor. Do not import a workflow SDK. Do not run a cluster. A crash resumes by reading the log and continuing from the last recorded event. Generative calls are activities whose outputs are stored. Replay does not call the model again.

### Go

**FACT.** The Temporal server is Go. The Go SDK exists and includes a workflow determinism checker.

**TAMVORI DESIGN DECISION.** Go is a good language for Tamvori's governor. Temporal's server is evidence that Go can host durable orchestration. It is also evidence that a Go architecture can become a distributed system if scale is taken as a requirement too early. Tamvori uses Go's standard library, not Temporal.

### Generalizes

Durable, replayable orchestration generalizes to every domain on the list. The cluster topology does not. Media, documents, and software production all need the history idea. None of them need matching-service shards in the first slice.

---

## 5. OpenMontage

Repository: https://github.com/calesthio/OpenMontage
Observed HEAD: `08e2151fa02de28a5d6a312b3d575692bf147ad7` (2026-09-06). No GitHub release.
Agent guide observed at blob `61c704149e4f81d653fe7150981b0946d47efeee`. That blob was not re-fetched at HEAD. HEAD's commit only changes `README.md`.
License: AGPL-3.0 (GitHub metadata).

### Problem

Turn a plain-language video request into a finished production: research, script, assets, edit, and render, using an AI coding assistant the user already runs.

### Fundamental idea

The agent is the orchestrator. Python provides tools and persistence. Pipelines, stage direction, review, and checkpoint policy live in YAML and Markdown instructions that the agent is told to read.

**FACT.** The agent guide states: "The AI agent IS the intelligence." It states that Python contains no orchestration logic, creative decisions, review logic, or checkpoint policy. The loop is: read manifest, read stage skill, call tools, self-review, checkpoint, present for human approval. It requires every production request to go through `pipeline_defs/` and stage director skills. Repository description: 12 pipelines, 100+ tools, 700+ agent skill and production-knowledge files.

**FACT.** The guide also requires announcing provider, model, and cost before consequential generation, and asking before major creative changes. Human approval is part of the instructed loop when the stage says so.

### What it does exceptionally well

It treats a video as a production with stages, gates, and a decision log, rather than as one prompt to a video model. Tool discovery is explicit. Cost is supposed to be visible before a paid call. Human checkpoints exist. The instructions are readable, which is better than hidden prompts buried in a service.

### Learn

- Production is a sequence of stages with artifacts, not a single generation call.
- Tool availability should be discovered and then frozen for the run.
- Decisions (provider, model, cost) should be recorded before money is spent.
- Human approval points are legitimate.
- Domain knowledge can live in documents a worker reads, as long as those documents are not the control plane.

### Do not reproduce

- The agent as the orchestrator. If the model skips a file, the pipeline, the review, and the checkpoint policy are suggestions.
- Self-review as the quality gate. The guide's loop says the agent self-reviews using a reviewer skill, then may present to a human. The reviewer and the producer are the same authority unless a separate verifier runs.
- Hundreds of skill files and a dozen pipelines as the way to encode a domain. That is a second program, written in natural language, with no compiler.
- Media pipelines, Remotion, Blender, and model vendors inside the core. The repository topics are the media industry.

### Where complexity accumulated

**FACT.** The project's own description counts 12 pipelines, 100+ tools, and 700+ skill and knowledge files.

**INFERENCE.** Complexity moved out of Python and into the instruction corpus. That looks simple in the code layout and is still a large system: the agent must select a pipeline, read the right skills, and obey them. Failure mode is silent non-compliance, which a deterministic governor would treat as a failed transition.

### Public interface

The interface is the coding assistant plus a required reading list. Powerful for an expert user who trusts the agent to follow instructions. Opaque as an API: there is no small set of typed operations whose preconditions the runtime enforces.

### Clean sheet

The governor reads a step graph and invokes workers. A media adapter may ship a stage list as data the governor executes. Skills, if they exist, are inputs to a generative worker, hashed and pinned, not the source of control flow. Self-review can attach a claim. Only an independent verdict advances the run.

### Go

**INFERENCE.** OpenMontage uses the coding agent as a universal orchestrator because that agent can already read files and run Python. Tamvori inverts that. Go is advantageous exactly where OpenMontage refuses to put logic: state transitions, checkpoint policy, and review authority belong in compiled code that the model cannot skip.

### Generalizes

Staged production, frozen tool plans, cost announcement, and human gates generalize. The video ontology and the instruction-driven control plane do not.

---

## 6. video-use

Repository: https://github.com/browser-use/video-use
Observed HEAD: `b877063835e6ea6e457124da7e28a0ae26691dc3` (2026-09-24). No release tag was consulted.
License: MIT.

### Problem

Edit real footage with a coding agent: transcribe, cut, grade, subtitle, and render, without a traditional editor UI.

### Fundamental idea

The model does not watch the video as its primary representation. It reads a packed transcript and, at decision points, a rendered timeline image. Cut decisions become an edit decision list (`edl.json`). Deterministic helper scripts and FFmpeg render the result. Hard correctness rules are distinguished from taste.

**FACT.** `SKILL.md` states those principles, including strategy confirmation before any cut, outputs confined to an `edit/` directory, and "Verify your own output before showing it to the user." Hard rules cover subtitle order, concat strategy, audio fades, word-boundary cuts, and transcript caching. The directory layout includes `edl.json`, transcripts, `preview.mp4`, and `final.mp4`.

**FACT.** HEAD's commit message adds a critic sub-agent inside self-eval, and changes render behavior so a missing subtitles file fails the render instead of warning.

### What it does exceptionally well

It finds a small, model-legible representation of a huge binary (word-level transcript, about a dozen kilobytes in the project's own description of the packed view). Taste stays free. Correctness is written as rules a render step can enforce. The edit decision list is an artifact a different program could execute. Caching transcripts on source identity avoids repeated paid calls.

### Learn

- Large binary artifacts need a canonical structured proxy before a model should reason about them.
- An edit decision list separates creative choice from rendering.
- Some checks belong in the renderer and should fail closed (missing subtitles file exits with an error).
- A human confirms the plan before the expensive or destructive step.
- Cache keys should be the source artifact, not the session.

### Do not reproduce

- Self-evaluation, including a critic sub-agent spawned by the same skill, as the gate before a human sees the work. A critic that shares the producer's instructions and goal is still on the producer's side. It can be evidence. It cannot be the verdict.
- The skill file as an ever-growing constitution of FFmpeg recipes. The hard rules are domain adapter knowledge. They will rot inside a general core.
- "Invent freely" plus shell access as the execution model. Bounded workers with declared outputs are the alternative.

### Where complexity accumulated

**INFERENCE.** The skill accumulates worked examples, animation stacks (HyperFrames, Remotion, Manim), and correctness rules in one document an agent must follow. Each rule is locally earned (a pop, a caption drift, a silent missing subtitle). The accumulation pattern is the same as OpenMontage's skill corpus, at a smaller scale: production knowledge encoded as instructions the model may skip, plus a growing helper toolbox.

### Public interface

Conversation. The user confirms a plain-English strategy. There is no typed API. That is a virtue for a skill and a defect for a governed system, which needs a plan artifact with a hash.

### Clean sheet

A generative worker may propose an edit decision list. A deterministic worker renders it or, before FFmpeg exists, validates and packages it. A verifier checks schema, hashes, word-boundary constraints, and any other predicate written in the pinned acceptance file. The model's self-check is stored as a claim artifact and does not flip the run to accepted.

### Go

**INFERENCE.** Rendering will eventually be `os/exec` around FFmpeg, which Go does directly. The governor does not need a Python skill runtime. The media proxy (transcript, decision list) is an adapter schema. Go's role is to hash the source, refuse to render if the decision list fails schema, and record the render inputs.

### Generalizes

Proxy-then-decide, decision-list-then-execute, and fail-closed deterministic checks generalize to documents (canonical structure, then proposed action), software (patch, then tests), and procurement (structured quote, then purchase decision). The caption and fade rules do not.

---

## 7. ComfyUI

Repository: https://github.com/Comfy-Org/ComfyUI
Observed release: v0.38.0 (2026-09-29). License: GPL-3.0. Language: Python.

### Problem

Give practitioners precise, repeatable control over diffusion and related generative models, locally, without hiding every parameter behind one prompt box.

### Fundamental idea

A node graph is the workflow. A prompt queue feeds an executor. Nodes run when their inputs are ready. Work is skipped when inputs have not changed.

**FACT.** Repository description: modular diffusion GUI, API, and backend with a graph/nodes interface. README text states an asynchronous queue, JSON save/load of workflows, and that only changed parts of a graph re-execute between runs. README text also states that only the subgraph with satisfied inputs runs.

**INFERENCE.** Third-party source tours attribute cache keys to input signatures and describe custom-node expansion, lazy evaluation, and model memory management. Those details match the public partial-reexecution claim, but this research did not read `execution.py`. Cache-key identity is treated as an inference supported by the README's re-execution claim, not as a quoted implementation.

### What it does exceptionally well

Partial re-execution. If the upstream artifact is unchanged, downstream work can be skipped. Workflows are data (JSON), so they can be saved, diffed, and submitted to an API. The graph makes a generative pipeline inspectable.

### Learn

- Artifact identity should decide whether work needs to run again.
- A workflow submitted as data is easier to hash and store than a workflow that exists only as UI state.
- A queue separates submission from execution.

### Do not reproduce

- The custom-node ecosystem as an extension model. v0.38.0's changelog is dominated by model-specific nodes, partner-vendor nodes, and memory optimizations. That is the correct focus for a diffusion engine. It is the wrong center of gravity for a domain-independent governor.
- A GPU memory manager, sampling loop, or model patcher anywhere in the core.
- GPL as a design input. The license is noted so Tamvori does not copy ComfyUI code. The architectural lesson does not require it.
- A visual graph editor. JSON graphs are enough if something must be a graph. The first Tamvori runs can be a list of steps with explicit dependencies.

### Where complexity accumulated

**INFERENCE.** Complexity sits in the node catalog and in per-model execution details. The executor's idea (ready inputs, cached outputs) is stable. The product grows by adding nodes. Partner nodes for external vendors show the same integration pressure as n8n, inside a media engine.

### Public interface

A canvas of nodes with model-specific widgets, plus an HTTP/WebSocket prompt API. The API is smaller than the canvas. Both are larger than "submit a graph of artifact-producing functions."

### Clean sheet

Content-addressed artifacts. A step runs when its input hashes are not already mapped to a verified output. No node class hierarchy. No UI. Model graphs, if a media adapter needs them later, are a worker's private representation. The governor sees input hashes, output hashes, and a verdict.

### Go

**INFERENCE.** ComfyUI is Python because PyTorch is Python. Tamvori's core never loads a model. Go is a better fit for the queue, the hash map, and the process supervisor. Calling into a Python model worker later is an adapter boundary, not a reason to write the governor in Python.

### Generalizes

Content-addressed skip and "workflow is data" generalize to every domain. Diffusion nodes, VRAM policy, and the canvas do not.

---

## 8. Docling

Repository: https://github.com/docling-project/docling
Observed release: v2.131.0 (2026-09-29). License: MIT for the codebase. Model weights have their own licenses, per the project README.
Architecture page observed at tag v2.92.0, not re-fetched at v2.131.0.

### Problem

Convert messy documents (PDF and many other formats) into one structured representation that downstream AI systems can use, with local execution as a supported path.

### Fundamental idea

A format-specific backend parses. A pipeline enriches. Both produce a single `DoclingDocument`. Export and chunking happen after that canonical document exists.

**FACT.** The v2.92.0 architecture page says the document converter selects a backend and a pipeline per format, and the conversion result contains the Docling document. The project README describes many input formats, a unified document representation, Markdown/HTML/JSON export, local execution, and optional integrations (LangChain, LlamaIndex, MCP, an API server).

**FACT.** The v2.131.0 release notes still add and fix format-specific backends and pipelines (Keynote, PPTX, DOCX, HTML, LaTeX, USPTO, and others).

**INFERENCE.** The backend-plus-pipeline-plus-canonical-document split is still the architecture at v2.131.0, because the release continues to use those words. This research did not confirm that the v2.92.0 diagram is unchanged.

### What it does exceptionally well

One document model is the product. Formats are adapters. Local execution is a first-class mode, so a sensitive document does not have to leave the machine. Export is downstream of structure, so Markdown is a view, not the source of truth.

### Learn

- A canonical structured artifact should sit between raw input and any generative use.
- Format support is a set of adapters behind one model, not a set of products.
- Deterministic parsing and model-based enrichment can share a pipeline if their outputs land in the same artifact and their failures are distinguishable. The v2.131.0 notes include a fix to report failed remote VLM calls as inference errors rather than empty success. Fail closed.

### Do not reproduce

- A universal parser with every format, OCR engine, and VLM inside the first Tamvori binary. Docling's format list is the point of Docling. Tamvori should call a document worker when a domain needs one.
- Framework integrations as core dependencies. LangChain and LlamaIndex adapters are distribution choices.
- Inheritance-heavy backend hierarchies as the Tamvori extension model. A worker contract is enough.

### Where complexity accumulated

**INFERENCE.** Each new format and each enrichment model adds a backend or a pipeline option. The v2.131.0 changelog is long because the format surface is the product. The central document type prevents that surface from fragmenting the output, which is why the complexity stays contained. Tamvori should copy the containment, not the format catalog.

### Public interface

A converter API with pipeline options, plus CLI, MCP, and an optional serve API. The simple path (convert a file, get a document) is clear. The option space for OCR, VLM, and format backends is large, which is justified for a parser toolkit and unjustified for a governor.

### Clean sheet

Documents enter Tamvori as artifacts. A document adapter, possibly Docling invoked as a worker, emits a canonical structured artifact plus a hash. Later, a generative worker proposes an action. A verifier checks the action against the document and against acceptance criteria. The core has no PDF code.

### Go

**INFERENCE.** Docling is Python because its parsers and models are Python. Go does not need to reimplement them. Go is advantageous as the process that stores the canonical artifact, refuses to proceed without it, and records which parser version produced it. Reimplementing Docling in Go would be a port, which this project rejects.

### Generalizes

Canonical artifact before generative use is the document-to-action pattern, and the same pattern as video-use's transcript and ComfyUI's cached tensors. Format backends do not generalize into the core. They generalize as the definition of an adapter.

---

## Cross-system reading

**INFERENCE.** The eight systems cluster into three mechanisms that keep showing up:

1. A boundary between a proposer and an executor (OpenHands sandbox, Temporal activities, video-use decision list versus FFmpeg, Docling parser versus downstream AI).
2. A record that makes the run recoverable (OpenHands events, Temporal history, n8n execution data, ComfyUI cached outputs).
3. A pressure to grow by catalog (n8n nodes, ComfyUI nodes, OpenMontage skills, Docling formats, Browser Use model adapters).

**TAMVORI DESIGN DECISION.** Adopt (1) and (2) in a single process with an append-only provenance log and content-addressed artifacts. Resist (3) by keeping catalogs in adapters and by giving complexity a budget. Add a fourth mechanism none of them put in the kernel: the worker that produced an artifact cannot be the authority that accepts it. Browser Use's cloud judge, OpenMontage's human gate, and video-use's fail-closed renderer each show a piece of that mechanism. None of them make it the center of a domain-independent core.
