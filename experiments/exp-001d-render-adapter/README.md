# EXP-001D — Render adapter

Phase: `TAMVORI-EXP-001D-RENDER-ADAPTER`

Frozen C3 baseline: `d9e72ca695f4e0f7d586b7771ed481d28e1ac7d4`

## What this experiment is

**Governed deterministic text-card rendering proof.**

It proves that an accepted media package can authorize bounded FFmpeg execution and produce a content-addressed MP4 with provenance, without contaminating `internal/kernel`.

## What this experiment is not

- Useful visual video production for customers
- Cross-environment byte-identical rendering claims
- A product demo, brand film, or polished media pipeline

Scenes are black frames with white `drawtext` captions. That is intentional for infrastructure proof. Visual product imagery is out of scope for 001D (see EXP-001E).

## Hypothesis

An accepted media package can drive bounded FFmpeg execution and produce a verified, content-addressed MP4 without contaminating `internal/kernel`.

## FFmpeg discovery

| Item | Value |
| --- | --- |
| FFMPEG_AVAILABLE | YES |
| FFMPEG_PATH | `C:\ffmpeg\bin\ffmpeg.exe` |
| FFMPEG_VERSION | N-118858-g740d400965-20250318 |
| FFPROBE_AVAILABLE | YES |
| FFPROBE_PATH | `C:\ffmpeg\bin\ffprobe.exe` |

## Defaults

1280×720, 30 fps, 2s per card, MP4/H.264, black cards with white drawtext. Font discovered from system (`arial.ttf`), copied only into the run workdir (not committed).

## Authority

`AuthorizeAndRender` transforms + verifies with hardened criteria first. Rejected packages never invoke FFmpeg.

## Safety

- `exec.CommandContext` with explicit argv (no shell)
- model text written to `textfile=` cards, not shell strings
- outputs constrained to workdir
- timeout 120s, retries 0

## Evidence

| Item | Value |
| --- | --- |
| SOURCE_MEDIA_PACKAGE_HASH | b16800608182234c09ede3678ac3f094f41824e4c295098bdc452ef2883038bb |
| PLAN_HASH | 26d97b379ad0b1dd0b068bb3be2cc4cf6952ab007cb060327054a1d9d6de6fed |
| RENDER_1_SHA256 | 56c41a99b474e1d18de773c73317403ec0606761656b871dc85391d41f734fad |
| RENDER_2_SHA256 | 56c41a99b474e1d18de773c73317403ec0606761656b871dc85391d41f734fad |
| SAME_ENV_BYTE_DETERMINISM | PASS (same machine/toolchain only) |
| PLAYABLE_PATH | `experiments/exp-001d-render-adapter/evidence/artifacts/governed-render.mp4` |

`LIVE_OPENAI_CALLS_DURING_001D = 0` (froze C3 package).

## Git recommendation for MP4

`LOCAL_ONLY` — ignored via `.gitignore` (binary experiment output). Keep locally for visual inspection.

## Next

EXP-001E: imagery-primary visual demonstration using the same verify-before-render boundary.
