# EXP-001C — Provider slice

Phase: `TAMVORI-EXP-001C-PROVIDER-SLICE`

Frozen B2 baseline: `674742a1738bdaf3ef40638caf0e84b1a34b105e`

## Hypothesis

A real generative model can propose a media-package candidate while deterministic software remains authoritative over criteria, transform, verification, acceptance, identity, and provenance.

## Provider

| Choice | Value |
| --- | --- |
| Provider | OpenAI (one only) |
| Model | gpt-4o-mini |
| Why | `OPENAI_API_KEY` already present in the development environment; HTTP JSON API works with Go stdlib |
| Integration | `internal/openai` → `media.ProposeLive` |
| Modules | zero third-party |

No provider registry, interface hierarchy, or SDK.

## Secret handling

- Key from `OPENAI_API_KEY` only
- Never logged, never in provenance notes beyond `provider=openai model=... response_hash=...`
- Raw evidence is the HTTP response body only

## Nondeterminism boundary

- Live assistant content is nondeterministic across calls
- Frozen assistant JSON + same criteria + same transform → same artifact hash (offline `ProposeFrozen` / evidence files)

## Implementation

- `Candidate.Evidence` / `EvidenceNote` + provenance kind `provider_response`
- Propose network errors become governed rejection (no fabricated accept)
- Offline tests use `httptest`; live path is `tamvori run --live`
- Fixture path unchanged for EXP-001B behavior

## Live proof

See [evidence/LIVE_RUN.md](evidence/LIVE_RUN.md). One attempt; accepted.

## Complexity

| Metric | B2 | After 001C |
| --- | ---: | ---: |
| Production Go LOC | 689 | ~947 |
| Exported symbols | 21 | 29 |
| Production packages | 2 | 3 (`openai` added) |
| Third-party modules | 0 | 0 |
| Interfaces | 0 | 0 |

## Lessons

1. Stdlib HTTP was enough; an SDK would have been pure convenience.
2. Capturing raw response bytes (not headers) is enough for replay evidence.
3. Empty asset names from the model still need schema discipline later; verification passed via derived hashes — watch this.

## Next

Optional EXP-001D: deterministic render worker (FFmpeg) as adapter, still no provider framework.
