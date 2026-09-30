# Evidence log

Observation window: 2026-09-29 America/Denver. GitHub API `updated_at` values are UTC and often read 2026-09-30.

Labels used in [RESEARCH.md](../RESEARCH.md):

- **FACT** — directly observed in a repository API response, license file, release, or project document fetched in this window.
- **INFERENCE** — a conclusion drawn from those facts. It is not a claim that the project authors stated it.
- **TAMVORI DESIGN DECISION** — a choice for this architecture. It is not a claim about another project.

No source code was copied into this repository.

## Systems given a full dossier

| Project | Repository | Version observed | License observed |
| --- | --- | --- | --- |
| OpenHands | https://github.com/OpenHands/OpenHands | Release v1.24.0, published 2026-09-25, target commit `7dc6805406ea3c76cb4a3ce407c3c72d481b0ac6` | GitHub metadata: MIT. README text carves out `enterprise/` under a separate license. |
| Browser Use | https://github.com/browser-use/browser-use | Release 0.13.10, published 2026-09-04, target commit `5c892e013a73e6622e6f50336e1eb0aa2c4405f2` | MIT |
| n8n | https://github.com/n8n-io/n8n | Release `n8n@2.41.3`, published 2026-09-25, target `release/2.41.3` | `LICENSE.md` on `master`: Sustainable Use License, with `.ee` files excluded. GitHub SPDX: `NOASSERTION`. |
| Temporal | https://github.com/temporalio/temporal | Release v1.32.0, published 2026-09-11, target `release/v1.32.x` | MIT. `docs/architecture/README.md` fetched at tag v1.32.0. |
| OpenMontage | https://github.com/calesthio/OpenMontage | No GitHub release. HEAD commit `08e2151fa02de28a5d6a312b3d575692bf147ad7` (2026-09-06). Agent guide blob `61c704149e4f81d653fe7150981b0946d47efeee`. | GitHub metadata: AGPL-3.0 |
| video-use | https://github.com/browser-use/video-use | No release tag consulted. HEAD commit `b877063835e6ea6e457124da7e28a0ae26691dc3` (2026-09-24). | MIT |
| ComfyUI | https://github.com/Comfy-Org/ComfyUI | Release v0.38.0, published 2026-09-29, target branch `master` | GPL-3.0 |
| Docling | https://github.com/docling-project/docling | Release v2.131.0, published 2026-09-29 | MIT. Architecture page fetched at tag v2.92.0. |

Star counts below are GitHub API `stargazers_count` at observation time. They are popularity context, not architectural evidence.

| Project | Stars | Primary language | Default branch |
| --- | --- | --- | --- |
| OpenHands | 89563 | TypeScript | main |
| Browser Use | 116762 | Python | main |
| n8n | 206317 | TypeScript | master |
| Temporal | 23377 | Go | main |
| OpenMontage | 61911 | Python | main |
| video-use | 27672 | Python | main |
| ComfyUI | 135539 | Python | master |
| Docling | 68199 | Python | main |

## Additional observations used

- Temporal Go SDK module page `https://pkg.go.dev/go.temporal.io/sdk` reported module version v1.49.0 during this research. That version is the SDK, not server v1.32.0.
- OpenHands `openhands/architecture/system-architecture.md` was fetched from tag 1.6.0. The same path returned 404 for tag v1.24.0. Current README and `AGENTS.md` on `main` describe Agent Canvas in this repository and agent execution in `OpenHands/software-agent-sdk`, with separate automation and TypeScript client repositories.
- An earlier OpenHands `AGENTS.md` blob `15423fbe4b53adc61097b80919b6204cefd2bc19` describes an in-repository `enterprise/` tree under the Polyform Free Trial License.
- Browser Use `CLOUD.md` (blob prefix `d19ec6ef`) describes a hosted product in which an independent judge scores an agent trajectory. `AGENTS.md` (blob `d1690e510a311cd69ad437ab53a024c86cf2a138`) describes the library loop: the agent decides the next action until it marks the task complete.
- n8n repository description says "400+ integrations." A rendered README excerpt during search claimed a larger integration count and "human approvals." The API description is the figure used as fact. The larger marketing figure is not treated as settled.
- Secondary writeups (DeepWiki, third-party ComfyUI tours) were used only to form inferences, and those inferences are labeled.

## Considered and not given a full dossier

Dify (`https://github.com/langgenius/dify`) was surveyed at headline level. GitHub metadata: license `Other`, description as an agentic-workflow and RAG platform. A public architecture overview describes a multi-service deployment (API, worker, plugin daemon, data stores). It repeats the plugin-platform lesson already taken from n8n, so a ninth full dossier would not change the architecture. Flowise, Langflow, and Crawl4AI were not studied in depth. Crawl4AI's absence is acceptable because Docling covers structured document processing, which is the relevant lesson for document-to-action work.

## What this log does not prove

It does not prove internal implementation details that were not read. It does not prove that a blob used for a guide is the same tree as the named release unless the fetch URL says so. Design decisions in the other documents are Tamvori choices justified by this evidence, not descriptions of what those projects intended.
