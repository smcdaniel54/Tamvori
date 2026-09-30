# Complexity budget

Complexity can fail a build. The goal is the minimum complexity required to deliver the capability, not the lowest number that still compiles.

## What is counted

| Metric | Why it is a signal |
| --- | --- |
| Third-party module requirements | Supply chain and API surface the team does not control. |
| Packages in the core module | Navigation cost. |
| Exported symbols in the core module | Public API the next change must preserve. |
| Config files and config keys | Behavior that is neither code nor criteria. |
| Core concepts | The eight in [ARCHITECTURE.md](ARCHITECTURE.md). A ninth is an architecture change. |
| Interfaces in the core | Indirection. |
| Executables and long-running services | Operational surface. |
| Required external services | Anything that must be up before a run works. |
| Layers between objective and artifact | Each layer is a place for hidden state. |
| Production lines | A trailing indicator. Do not optimize it with clever compression. Use it to notice unexplained growth. |

Cyclomatic complexity is a per-function lint once code exists. It is a poor repository-level budget and is not a gate in the first slice.

## Budgets for the first implementation (EXP-001B)

These ceilings are proposals for the phase that writes code. They are not enforced in EXP-001A because EXP-001A contains no Go module.

| Metric | Ceiling |
| --- | --- |
| `go.mod` requirements | 0 |
| Core packages | 6 |
| Exported interfaces in core | 5 |
| Executables | 1 (`tamvori`) |
| Long-running services | 1 (that same process, and it may be a CLI that exits) |
| Required external services | 0 |
| Databases, brokers, containers | 0 |
| Config files | 1, and only if flags are insufficient |
| Config keys | 8 |
| Core concepts | 8 |
| HTTP routes | 0 |
| Layers from objective pin to stored artifact | 3: run, step, worker |
| Adapter concepts linked into the core state machine | 0 |

The media package schema does not count against core concepts. It counts as an adapter. If the schema's types appear in the governor's switch statements, that is a budget failure.

## When a ceiling may rise

A human authorizer records a one-line reason: the capability, why it cannot fit, and which ceiling moves by how much. The change is a criteria version. An AI proposer may suggest the rise. It may not apply it by editing the budget file inside a candidate that is judged by the old tests only. Budget files are acceptance criteria.

A ceiling should not rise to make room for a dependency "we will need later." Later has its own phase.

## Detecting creep

Once EXP-001B exists, the test command should fail when:

- `go list -m all` shows any requirement beyond the module itself
- the number of directories under the core tree exceeds the package ceiling
- an HTTP listen is introduced
- a second `func main` appears
- the architecture concept heading count in `ARCHITECTURE.md` exceeds 8 without a versioned budget change

A later, still small, check can parse Go files for exported identifiers. Hand-maintained counts will drift. Prefer a test that measures the repo.

Review cue for humans and for AI proposers: a pull request that adds a concept must name which existing concept was insufficient, in the words of the "why it exists" table.

## Removing complexity

The `simplify` proposal in [CONTINUOUS_IMPROVEMENT.md](CONTINUOUS_IMPROVEMENT.md) is the mechanism. An agent may propose deletion of:

- a dependency
- an exported symbol
- a config key
- a package
- a concept, if it can show the concept's job is already done by one of the eight
- an adapter that leaked into core

The proposal passes when the pinned behavioral suite passes and the measured metric is lower. Deletion without that evidence is churn, and the verdict fails it.

## Known pressure, watched on purpose

Research showed catalogs (integrations, nodes, skills, formats, model adapters) as the usual growth path. The budget's job is to make that path visible the moment someone adds the second media noun to the core or the first third-party module "for JSON." The standard library already parses JSON.
