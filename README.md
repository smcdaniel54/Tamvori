# Tamvori

Tamvori is a Go-native platform for turning an objective into a verified outcome.

Generative AI proposes and creates. Deterministic software governs. Specialized workers do bounded work. An independent verifier decides whether that work met acceptance criteria the worker cannot change. Provenance records what happened. Improvements must beat the current baseline, including improvements that only remove complexity.

Continuous media production is the first workload intended to prove this architecture. Tamvori is a governed production system. Media is an adapter, not the core.

## Economic hypothesis

The valuable work is expensive coordination: mistakes cost money, work crosses systems, evidence matters, and people spend their time shuttling tasks, checking outputs, and recovering from failure. Generative models are useful where the work is ambiguous. Deterministic governance is useful where a wrong release, a silent retry, or an unauditable decision is expensive. The commercial surface is governed, domain-specific autonomous workflows. This repository does not choose a final vertical.

Trusted document-to-action workflows are an important later direction: a document becomes a canonical artifact, a worker proposes an action, and a separate authority verifies the action against criteria before anything is released.

## Development model

Humans set objectives, business constraints, and acceptance criteria, and they approve major architectural decisions.

An AI development system may research, propose, design, write code and tests, run tests, and propose fixes, simplifications, and improvements.

Deterministic governance decides whether acceptance criteria passed. Workers do not edit those criteria in order to pass.

This repository is at the research freeze. It contains no production code.

## Current phase

[EXP-001A research and architecture freeze](experiments/exp-001a-research-architecture/README.md)

That experiment studied eight open-source systems as inputs, then defined an original architecture from Tamvori's requirements. Those systems are not implementation sources. Tamvori is not a port of any of them.
