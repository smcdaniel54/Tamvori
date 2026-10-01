# EXP-001C3 — Live regression

Phase: `TAMVORI-EXP-001C3-LIVE-REGRESSION`

Frozen C2 baseline: `206fb79a6a46b38be6080fa3b470eb5b570c5a2b`

## Hypothesis

After C2 hardened nonempty asset/scene rules, one real OpenAI call can produce an accepted package without weakening criteria or adding architecture.

## Historical-criteria resolution (from C2 freeze)

| Item | Value |
| --- | --- |
| Classification | HISTORICAL |
| EXP-001B criteria path | `experiments/exp-001b-kernel-slice/acceptance/criteria.json` (restored B evidence) |
| Current hardened criteria | `acceptance/media-package.json` |
| History rewritten | NO |

## Prompt difference

Communicated already-pinned requirements: nonempty trimmed title/audience/message/narration; nonempty scene id/outline; nonempty asset names. Prompt does not grant acceptance authority.

## Live result

One attempt. Accepted. See [evidence/LIVE_RUN.md](evidence/LIVE_RUN.md).

## Contrast

| Response | Under hardened C2 criteria |
| --- | --- |
| Old EXP-001C frozen | REJECT (`empty asset name`) |
| New C3 frozen | ACCEPT |

## Replay

Frozen C3 assistant + same criteria → same artifact hash offline.

## Complexity

Prompt string change only in production; evidence + focused regression test. Provider count remains 1; zero third-party modules.

## Next

EXP-001D render adapter (optional, human-authorized). Verification/provider path is adequate to freeze before rendering.
