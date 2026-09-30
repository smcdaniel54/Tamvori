# Continuous improvement

Designed, not implemented.

## Loop

```text
observe
  -> name a deficiency or a simplification
  -> write a proposal artifact
  -> run the candidate under pinned criteria
  -> evaluate with an independent verdict
  -> compare to the baseline run
  -> accept or reject
  -> keep the comparison in provenance
```

Observe means read measurement artifacts and failure verdicts already in the log. A generative worker may propose the change. It may not start the candidate run by itself unless policy allows auto-start for a low-risk class. Acceptance is never automatic from the proposer's claim.

## The rule that keeps the loop honest

The candidate run pins the same acceptance-criteria hash as the baseline. The proposal worker's write set excludes the criteria store. A diff that touches criteria fails the proposal at rank 1, before any outcome is considered.

A change to criteria is a different operation:

- An authorizer creates a new criteria version.
- The new version has a new hash.
- Old runs keep their pinned hash.
- The authorizer records a reason.
- No in-flight run is re-pinned.

Re-grading history under new criteria is a new run that names the old artifacts. It does not edit the old log.

This is the ownership model for acceptance criteria. Workers do not own them. Verifiers do not own them. Proposers do not own them. Authorizers own new versions. The governor owns enforcement.

## Comparison

A proposal declares one of two intended benefits:

1. Outcome benefit. A named metric on a measurement artifact improves against the baseline by a threshold written in the pinned criteria, and no protected metric regresses.
2. Simplification. A named complexity metric drops, every pinned acceptance check still passes, and no protected outcome metric regresses.

A change that only adds surface area, with no measured outcome benefit and no complexity drop, is rejected. Ties lose. The baseline remains current.

Protected metrics and thresholds live in the criteria, so the candidate cannot pick a friendlier metric after seeing the result.

## Regression

Regression is a failed rank-4 verdict: a protected metric on the candidate is worse than the baseline beyond the tolerance in the criteria. The candidate run is then `rejected`. The current accepted artifact, policy, and worker versions stay in place.

Flaky generative metrics need a rule in the criteria before they are allowed to gate acceptance: for example, a deterministic fixture only, or a minimum sample count. Until that rule exists, generative quality scores are reported and do not gate. Deterministic checks do.

## Rollback

Rollback is a pointer move, not a history rewrite.

- The accepted version is an event: `accepted_pointer` names an artifact hash or a policy hash.
- Rollback appends a new event naming the previous hash, with an authorizer identity.
- The log still contains the failed candidate.
- External side effects that already escaped (a sent message, a published file) are not undone by the pointer. Adapters that have external effects must make release a separate authorized event, and rollback of those effects is an adapter compensation step with its own run. The core does not pretend a pointer move recalls an email.

## Simplification as a first-class improvement

A proposal schema `simplify` must name:

- the metric (dependency count, package count, exported symbols, config keys, concept count, service count)
- the baseline measurement
- the deletion or merge

The candidate is the tree or the configuration after the deletion. The verifier re-runs the pinned suite. The complexity check re-measures the metric. Pass requires the suite green and the metric lower. A simplification that needs a criteria edit to pass is rejected by the ownership rule.

An AI developer is allowed, and later expected, to file `simplify` proposals. The governor, not the developer, accepts them.

## Version pins on a run

Every run records:

- criteria hash
- policy hash
- worker version for each step
- input artifact hashes
- output artifact hashes
- verdict hashes

Comparing two runs means comparing those pins and the measurement artifacts, not comparing chat transcripts.

## What EXP-001A does not build

No proposal runner, no metric store, no automatic deletion of code. The next experiments can add a complexity measurement to the test command before they add a generative proposer. Measurement first, proposer second.
