# Verification model

Hard rule: no worker is the sole authority that its own output is correct.

The governor enforces the rule by identity. The event that records an output names a producer. The event that records a verdict names an evaluator. If those identities are equal, the fold rejects the verdict. A claim artifact written by the producer is stored and does not change the projection to passed.

## What is being verified

A verdict names:

- the output artifact hash
- the acceptance-criteria hash pinned on the run
- the evaluator identity
- a result: pass or fail
- the evidence artifact hashes the evaluator used
- the check set that ran

A verdict that cites a different criteria hash than the run pinned is a failed event, not a pass. That is how a worker, or a confused verifier, is prevented from substituting an easier rubric.

## Evidence strength

Use the strongest class that can answer the question. A higher class outranks a lower class. A pass at a lower class does not override a fail at a higher class.

| Rank | Class | Example | Who may produce it |
| --- | --- | --- | --- |
| 1 | Identity and schema | Hash matches bytes. JSON matches the pinned schema. Required fields present. | Deterministic evaluator |
| 2 | Deterministic predicate | Acceptance assertions: counts, allowed enums, hash of a referenced artifact, equality with a golden fixture. | Deterministic evaluator, code or data pinned before the run |
| 3 | External observation | Process exit code, file exists, byte length, and later ffprobe or test-runner output, captured by a process that is not the producer. | Deterministic evaluator |
| 4 | Baseline comparison | Metric on this artifact versus the pinned baseline artifact. Regression is a fail when policy says the metric is protected. | Deterministic evaluator |
| 5 | Independent critic | A second model, with a different identity, asked a closed question, its answer stored as evidence. | Generative evaluator, never the producer |
| 6 | Human approval | Authorizer event naming the same artifact and criteria hashes. | Human authorizer |

Policy on a step declares the minimum rank required. The default minimum is rank 2. Rank 5 alone never satisfies a step whose question can be answered at rank 1–4. Rank 6 is added when policy says residual risk or an external side effect requires a person. Rank 6 does not erase a rank 2 failure.

### Prefer deterministic evidence

If a predicate can be written as a function of bytes, it is written that way and evaluated at rank 2. Model judgment is for residue: taste, ambiguity, and summaries a schema cannot capture. Residue judgments are labeled rank 5 and cannot flip a failed deterministic check to pass.

Examples for the first media package, all rank 1–2, all in the adapter's criteria rather than the core:

- Package parses as the production-package schema.
- Every asset path named in the package has an artifact hash that is present in the store.
- The criteria hash inside the run matches the file the human wrote.
- The producer field is not the verifier field.

No pixels are required.

## Claims

Producers may attach a claim: a self-check, a log, a critic sub-agent transcript. Claims are rank 0. They are not on the table above. They are kept so the record shows what the producer believed. The fold does not read them when deciding `passed`.

This is the direct response to self-review loops in agent media tools and to agents that mark their own browser tasks complete. Those self-checks can still be stored. They are claims.

## Independence, practically

Until real authentication exists, identity is a name in policy: `worker/package`, `verifier/package`, `human/authorizer`. The first tests use distinct names and a case where the same name is refused. A later authentication system replaces how names are bound. It does not replace the inequality check.

A critic model that shares the producer's prompt, memory, and goal is a weak rank 5. Policy may require a different provider or a prompt hash that differs from the producer's. That requirement is still weaker than rank 2. Do not spend design effort making critics elaborate while a schema check would do.

## Failure

A failed verdict on a required check fails the step. Retry happens only if policy marks that failure retryable and the budget remains. Schema failures of generative output are retryable by default, up to the budget, because the expected recovery is "ask again," recorded as a new attempt with a new hash. Hash mismatches in the store are not retryable generative events. They halt. They mean the log and the bytes disagree.

## What the verifier is not

The verifier is not a workflow engine, a second agent platform, or a plugin. In the first slice it is a function or a subprocess invoked by the governor with read-only access to the artifact store and the pinned criteria. It returns a verdict struct. The governor appends it.
