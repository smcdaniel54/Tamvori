# EXP-001B2 — Simplification

Phase: `TAMVORI-EXP-001B2-SIMPLIFICATION`

Frozen baseline: `dfe3f0243f9dde62cf87ec50eae5b191c4cd4a76` (EXP-001B).

## Hypothesis

Given a working governed kernel, AI can remove accidental API surface, dead code, and representational duplication without weakening frozen EXP-001B behavior.

## Complexity before (EXP-001B freeze)

| Metric | Value |
| --- | ---: |
| Production Go LOC | 798 |
| Test Go LOC | 313 |
| Packages | 2 |
| Interfaces | 0 |
| Third-party modules | 0 |
| Config files | 0 |
| Exported symbols | 40 |
| Core concepts | 5 |

## Audit (exports)

| Symbol | Judgment |
| --- | --- |
| `CriteriaHashOf` | TEST_ONLY alias of `HashBytes` → remove |
| `UnderDir` | TEST_ONLY → move into test helper |
| `Store` / `Get` | `Get` unused; `Store` only needed inside `Execute` → unexport store, drop Get |
| `Plan` | internal plan data → inline JSON map |
| `ProposeFunc`/`TransformFunc`/`VerifyFunc` | named aliases → inline function fields |
| Event kind consts | ACCIDENTALLY_EXPORTED → private; tests use string literals |
| `Event.Extra` | one-off → fold into `Note` |
| `eventBody` duplicate | merge via omitempty `EventHash` body hash |
| `Projection.Last*` | unused → remove |
| `SchemaID`, `Criteria`, `Scene`, `Asset`, `ProductionPackage`, `FixtureMode` | adapter-internal → unexport |
| `Log.Append` | only used inside kernel → unexport |
| kernel vs media packages | REQUIRED boundary → keep |

## Simplifications accepted

- Unexported store; removed unused `Get`
- Removed `CriteriaHashOf`, `UnderDir`, `Plan`, named func types, media public schema types
- Private event-kind constants; removed `Event.Extra` and duplicate `eventBody`
- Trimmed unused projection fields
- Deduplicated media required-field checks; dropped redundant JSON Compact / unused unmarshal
- Tests preserve semantic strength (string kind literals; stronger stored-artifact hash check)

## Simplifications rejected

- Merging `internal/kernel` and `internal/media` (would contaminate core)
- Removing Objective / Run / Artifact / Verdict / Provenance as named responsibilities
- Collapsing producer and verifier into one function
- Shrinking provenance event kinds below reconstructability
- Code-golfing `Execute` into fewer lines at readability cost

## Complexity after

| Metric | Value |
| --- | ---: |
| Production Go LOC | 689 |
| Test Go LOC | 326 |
| Packages | 2 |
| Interfaces | 0 |
| Third-party modules | 0 |
| Config files | 0 |
| Exported symbols | 21 |
| Core concepts | 5 (Objective, Run, Artifact, Verdict, Provenance) |

Production LOC −13.7%. Exported symbols −47.5%.

## Behavioral proof

`go test ./...` PASS · `go vet ./...` PASS

Frozen properties rechecked via the existing suite (valid/invalid/claim/pin/replay/hash/provenance/media separation/bounds).

## Lessons

1. Most of the “40 exports” were accidental publicity for types/consts only tests or one package needed.
2. Package count was already right; export count was the real overshoot.
3. Five core concepts remain justified after contact with code.

## Recommendation

Review and commit B2 as a simplification baseline. Next capability phase remains human-authorized provider integration (EXP-001C), not more architecture.
