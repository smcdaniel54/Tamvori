# EXP-001C2 — Verification hardening

Phase: `TAMVORI-EXP-001C2-VERIFICATION-HARDENING`

Frozen C baseline: `d0da49fcee711d7a4a0cf4a7ef3cee9c935881ef`

## Hypothesis

Deterministic verification must reject candidates that are structurally present but semantically empty (empty/whitespace strings, empty scene content, empty asset names), without subjective quality judgment or an LLM judge.

## Known defect

EXP-001C live package used `"name":""` assets and was accepted because criteria only required `len(assets) > 0` plus derived hashes.

## Adversarial matrix

| CASE | OLD (001C) | CORRECT | REASON |
| --- | --- | --- | --- |
| empty title | reject | reject | already TrimSpace |
| whitespace title | reject | reject | already TrimSpace |
| empty asset name | **accept** | reject | count/hash only |
| whitespace asset name | **accept** (after trim→empty still hashed) | reject | same |
| empty scene outline | **accept** | reject | count only |
| empty/whitespace narration | reject | reject | already TrimSpace |
| partial object | reject | reject | required fields |
| model self-claim | ignore | ignore | claims non-authoritative |
| extra unknown fields | accept if otherwise valid | accept | no schema framework |
| malformed JSON | reject | reject | transform/parse fail |
| valid control | accept | accept | regression |

## Changes

Pinned operational criteria (current) gained:

- `require_nonempty_asset_names`
- `require_nonempty_scene_ids`
- `require_nonempty_scene_outlines`

Transform trims string fields before hashing/sorting. Verify enforces the new pinned flags.

## Historical criteria integrity

`HISTORICAL_CRITERIA_CLASSIFICATION = HISTORICAL`

`experiments/exp-001b-kernel-slice/acceptance/criteria.json` is EXP-001B evidence of the criteria used then. C2 initially overwrote it; that was corrected before freeze:

- restored B criteria exactly from `dfe3f02`
- placed hardened criteria at `acceptance/media-package.json` (current operational baseline)
- runtime/tests now load the operational path

`HISTORY_REWRITTEN = NO` after correction. EXP-001B criteria remain reconstructable at their original path.

## Frozen 001C live response

`FROZEN_001C_RESPONSE_STILL_ACCEPTED = NO`

Replay rejects with `empty asset name`. Historical false acceptance confirmed; evidence files left unchanged.

## Export review (001C additions)

| Symbol | Class |
| --- | --- |
| ChatJSON | REQUIRED_CROSS_PACKAGE (media→openai) |
| WithEndpoint | REQUIRED_CROSS_PACKAGE (tests) |
| DefaultTimeout | REQUIRED_CROSS_PACKAGE |
| LiveProducerID | REQUIRED_CROSS_PACKAGE |
| DefaultLiveModel | REQUIRED_CROSS_PACKAGE |
| LiveConfig | REQUIRED_CROSS_PACKAGE |
| ProposeLive | REQUIRED_CROSS_PACKAGE |
| ProposeFrozen | REQUIRED_CROSS_PACKAGE |

No privatizations in C2 (secondary goal).

## Rejected over-complexity

JSON-schema dependency, LLM-as-judge, subjective quality scoring, schema engine.

## LIVE_CALLS_C2

0

## Next

EXP-001D render adapter remains optional; first fix prompt/criteria so live packages cannot omit asset names.
