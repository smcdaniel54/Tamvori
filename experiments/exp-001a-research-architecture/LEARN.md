# LEARN

Ideas observed in other systems, judged against Tamvori's requirements. Inclusion here means the idea survives evaluation. It does not mean Tamvori copies a design.

Sources and the fact/inference split are in [RESEARCH.md](RESEARCH.md).

## Kept

### 1. Proposer and executor are different roles

OpenHands separates the agent loop from the sandbox that performs actions. Temporal separates deterministic workflow code from activities. video-use separates an edit decision list from FFmpeg. Docling separates a canonical document from whatever generative system consumes it.

Tamvori keeps this as worker versus governor. The governor applies events. Workers return artifacts. A generative worker is a proposer. A deterministic worker is an executor. Neither writes run state.

### 2. History is the state

Temporal's premise is the strongest version: an append-only history, replayed to recover. OpenHands' event stream is a weaker, still useful version: the run is inspectable as events rather than as a chat bubble. n8n stores execution data so a run can be examined after the fact.

Tamvori keeps an append-only provenance log as the run's state. Recovery reads the log. It does not ask a model to remember.

### 3. Side effects are recorded, then replay reuses the record

Temporal replay returns stored activity results instead of repeating the side effect. ComfyUI skips nodes whose inputs have not changed. video-use caches transcripts per source file.

Tamvori keeps content-addressed artifacts. A step whose input hashes already have a verified output does not need to run again. A generative call is never assumed to be repeatable. Its output bytes are the thing that is replayed.

### 4. Structured observations, not raw environments

Browser Use sends a DOM-scale observation to the model. video-use sends a packed transcript and, only at decision points, a timeline image. Docling emits one document model.

Tamvori keeps this as an adapter duty. The core stores artifacts. Adapters owe the core a structured artifact a later worker can read without the original environment. Media owes a production package or a decision list, not a framebuffer. Documents owe a canonical document, not a PDF byte string, once parsing has happened.

### 5. Closed tool contracts and budgets

Browser Use maps model decisions onto a registry of browser actions. OpenMontage discovers tools and expects a plan and a cost before paid calls. video-use refuses to cut until a person confirms the strategy.

Tamvori keeps bounded worker contracts: allowed inputs, declared output schema, step budget, and money or time limits. Discovery is allowed before a run is frozen. After freeze, the worker set and the acceptance criteria hash do not change.

### 6. Human approval is a pause, not a feeling

n8n can place a human in the graph. OpenMontage instructs the agent to present work when a stage requires approval. video-use requires plain-language confirmation before the cut.

Tamvori keeps approval as a policy state. The run waits for an approval event from an identity that is not the worker. Silence is not approval. The worker cannot synthesize the approval event.

### 7. Fail closed on deterministic checks

video-use's observed HEAD stops the render when a declared subtitles file is missing, instead of warning and continuing. Docling's v2.131.0 notes include reporting a failed remote model call as an inference error rather than an empty success. Temporal treats a command mismatch on replay as a failure of the workflow task.

Tamvori keeps this bias. A missing required artifact, a hash mismatch, a schema failure, or an exhausted retry budget ends in a recorded failure. It does not end in a best-effort success.

### 8. Workflow as data, for hashing and explanation

n8n and ComfyUI store graphs as data. OpenMontage stores pipelines as YAML. Data can be hashed, diffed, and pinned to a run.

Tamvori keeps the step list of a run as data inside the provenance log. It does not keep a visual editor, and it does not keep a reusable workflow product in the first slice. A reusable template is future work. The pinned step list is enough.

### 9. Domain knowledge may be documents; control flow may not

OpenMontage and video-use show that a great deal of production knowledge fits in readable instructions. That is a good input to a generative worker.

Tamvori allows a worker to read pinned knowledge artifacts. The governor does not consult those documents to decide the next state. If the knowledge matters to acceptance, the relevant predicate is in the acceptance criteria, which are code or structured assertions evaluated by the verifier.

### 10. One canonical artifact per domain object

Docling's document type is why dozens of formats do not become dozens of downstream products. video-use's `edl.json` is why taste does not have to live inside the renderer.

Tamvori's core artifact is content-addressed bytes plus a media type and a schema name. Domain objects (document, media package, patch, purchase proposal) are schemas in adapters.

### 11. Local execution is a legitimate default

Docling treats local, air-gapped conversion as a feature. ComfyUI runs models on the operator's machine. Temporal allows a self-hosted cluster, at a much higher operational cost.

Tamvori's first governor is one local process with no required external service. Remote models and remote workers are adapters that must still return artifacts into the local log.

### 12. Determinism can be tested

Temporal's replay tests and workflow determinism checker show that "deterministic" can be a failing test rather than a slogan.

Tamvori will test the governor by replaying a fixed log and requiring the same projection. Generative workers are excluded from that test. Their recorded outputs are inputs to it.

## Evaluated and dropped

These appeared in the research brief as examples. They do not earn a core concept.

| Idea | Judgment |
| --- | --- |
| Graph workflows as the authoring UI | Data dependencies between steps are enough. A canvas is optional and out of scope. |
| Skill systems as the control plane | Skills may be pinned inputs. They are not the scheduler. |
| Plugin and node marketplaces | Catalog growth is the main complexity failure mode across n8n, ComfyUI, and OpenMontage. |
| Agents that choose their own tools without a frozen contract | Discovery before freeze is useful. Unbounded tool use during a run is not. |
| Model-specific architecture | Provider adapters sit behind a worker. The core has no model SDK. |
| Distributed worker fleets and warm pools | OpenHands and Temporal justify these at their scale. Tamvori's first slice is one process. |
| Self-review loops | A worker may attach a claim. Advancement requires an independent verdict. |

## What "learn" does not authorize

No file in any studied repository is an implementation source. The next experiment implements the Tamvori concepts in [ARCHITECTURE.md](ARCHITECTURE.md), using the Go standard library, from the acceptance criteria of that experiment. It does not translate these projects.
