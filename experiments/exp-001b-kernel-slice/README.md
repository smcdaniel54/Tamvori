# EXP-001B — Kernel slice

Phase: `TAMVORI-EXP-001B-KERNEL-SLICE`

Observed: 2026-09-30 (America/Denver).

Baseline HEAD: `ee794abbc864a7ecaf0166e3d4db245f2f00b885` (EXP-001A frozen).

## Hypothesis

A single Go process can execute one governed run—objective → pinned plan → generative-shaped worker → deterministic worker → independent verifier → content-addressed artifact → provenance—with zero third-party modules, while keeping acceptance criteria outside worker control and media schema out of the governor.

## Result

**PASS.** `go test ./...` and `go vet ./...` succeed. A valid fixture run ends `accepted` only after a verifier identity distinct from the producer. Invalid and self-claiming fixtures are rejected. Replay of the frozen valid fixture yields identical artifact hashes across work directories.

## Implementation

| Path | Role |
| --- | --- |
| `go.mod` | Module `github.com/smcdaniel54/Tamvori`, no `require` |
| `cmd/tamvori` | Single CLI: `tamvori run` |
| `internal/kernel` | Artifact store, hash-chained log, fold, `Execute` |
| `internal/media` | Production-package schema, fixture propose, canonicalize, verify |
| `experiments/exp-001b-kernel-slice/acceptance/criteria.json` | Pinned acceptance criteria (not owned by workers) |

### Flow

```text
objective + criteria bytes
  -> criteria hash pinned on run_started
  -> plan artifact pinned
  -> propose (fixture) -> candidate artifact (+ optional claim)
  -> transform (sort scenes/assets, derive asset hashes, canonical JSON)
  -> verify (different identity) against pinned criteria bytes
  -> run_accepted | run_rejected
```

Claims are recorded and ignored by `Fold` for acceptance. A verdict whose actor equals the producer of the artifact is ignored for acceptance.

## Architecture actually required

Concepts that earned concrete representation:

| Concept | Representation |
| --- | --- |
| Objective | `kernel.Objective` data |
| Run | `Execute` + workdir + projection status |
| Artifact | content-addressed files under `artifacts/` |
| Verdict | event + `kernel.Verdict` from verifier |
| Provenance | append-only JSONL hash chain |

## Concepts merged / deferred / eliminated

| EXP-001A concept | Disposition |
| --- | --- |
| Policy | **Merged** into `Bounds` (MaxAttempts, MaxBytes) and pinned criteria JSON |
| Step | **Merged** into `Plan.Steps` string data (not a workflow engine) |
| Worker | **Eliminated** as interface; **deferred** as type. Functions: `ProposeFunc`, `TransformFunc`, `VerifyFunc` |
| Four authorities as types | **Deferred**; enforced as identity strings on events |
| Human approval pause | **Deferred** |
| Skip-by-input-hash | **Deferred** |
| Crash resume of in-flight work | **Deferred** (log replay of completed events works; no partial-output resume) |

Implemented core concept count for this slice: **5** (Objective, Run, Artifact, Verdict, Provenance), with Plan as data and Bounds as pinned run limits.

## Determinism boundary

- `Fold` is pure over events.
- `media.Transform` is a pure function of candidate bytes.
- Generative fixture is frozen bytes for a mode; replay does not call a model.
- Artifact identity is SHA-256 of canonical bytes.

## Verification proof

- Valid fixture → independent `verifier/media-package` pass → `accepted`
- Invalid / claim_pass fixtures → `rejected` despite `ClaimValid`
- Same-identity pass verdict does not move projection to `accepted`
- Criteria hash pinned at start; disk tamper during propose does not change the pin

## Replay proof

Two `Execute` calls with `FixtureValid` and identical criteria produce the same `ArtifactHash`, `PlanHash`, and `CriteriaHash`.

## Failure-path proof

Covered by tests: invalid reject, claim ignored, same-identity fold ignore, hash-chain tamper detection, bounded retries (2 attempts then reject).

## Complexity measurements

| Metric | Value | EXP-001A ceiling |
| --- | ---: | ---: |
| Third-party modules | 0 | 0 |
| Production packages (`internal/*`) | 2 | ≤6 (phase target ≤3) |
| Executables | 1 | 1 |
| Interfaces | 0 | ≤5 (phase target ≤2) |
| Config files | 0 | ≤1 |
| Production Go LOC (non-blank, non-comment) | ~799 | — |
| Test Go LOC | ~313 | — |
| Exported symbols (types/funcs/consts across cmd+internal) | ~35 | keep small; see concern |

Over-design note: several kernel types and event constants are exported for the CLI and tests. A later simplify pass could unexport what only tests need.

## Deviations from EXP-001A / NEXT_EXPERIMENT

- Crash-resume of an in-flight unrecorded worker output was not implemented (deferred). Completed log re-fold works; chain tamper fails open.
- Acceptance criteria live under `experiments/exp-001b-kernel-slice/acceptance/` as specified.
- No HTTP, DB, plugins, FFmpeg, or model SDK.

## AI-built development record

- AI authored all production Go, tests, acceptance JSON, and this experiment record.
- First full `go test ./...` / `go vet ./...` after the initial implementation: **PASS** (exit 0).
- During authoring, a buggy `WriteDenied` path helper was replaced with `UnderDir` before the suite ran; no failing suite iteration was required after that fix.
- Deterministic tests that evaluated the result: valid accept, invalid reject, claim ignore, criteria pin, replay hash equality, content-address stability, transform determinism, provenance completeness, chain tamper, same-identity fold, media leakage scan, bounded retries, zero modules.

## Lessons

1. Function types beat a Worker interface for one slice.
2. Policy as `Bounds` + criteria JSON is enough; a Policy package would have been vacant.
3. Keeping media types out of `internal/kernel` is enforceable with a simple source scan test.
4. Export surface grew faster than package count; count exports next time as a first-class budget check in CI.

## Recommendation for next experiment

`TAMVORI-EXP-001C` (when authorized): one real generative provider behind the same `ProposeFunc` contract, still emitting a package or plan, still no renderer. Keep `go.mod` requirements at zero until a provider truly cannot be a subprocess with stdlib `os/exec`.

Do not start EXP-001C from this document alone.
