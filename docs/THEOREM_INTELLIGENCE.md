# Theorem Intelligence: Proof-Driven Compliance (experimental v0.3)

**Product principle:** Compliance is a set of scoped propositions. Evidence must support each proposition; uncertainty is not a pass. Open source readiness is free. An auditor or a regulator, not OpenCMMC, determines certification and official findings.

## Current working capability

`freecompliance prove` accepts a caller-supplied finite model of discrete variables, timestamped observations, conjunctions of technical obligations, and hypothetical candidate evidence. A deterministic Go evaluator enumerates every state within the declared domain (hard cap 4,096 worlds), excludes inadmissible evidence, and checks whether **every evidence-consistent modeled state** satisfies each obligation. It can show a counterexample assignment and choose a minimum-cost candidate evidence set through exhaustive search of **at most 12** candidate inputs.

A claimed `VERIFIED` evidence record is admitted only when it contains a source reference and an explicit validity window containing the assessment time. **The engine does not authenticate that source**, validate a real GitHub ruleset, or independently confirm that the finite model covers real-world behavior. No scans, deployments, privileged API calls or automation are triggered by the proof command. Model hashes are not signatures.

## Demo

```sh
go test ./...
go run ./cmd/freecompliance prove --input examples/theorem/branch-controls.json \
  --json theorem.json --html theorem.html
```

The synthetic configuration claims review and checks are required. It does not establish that an authorized principal cannot bypass both. The engine shows a modeled world with `bypass_possible=true` and identifies the smallest specified additional evidence candidate that, **if independently confirmed**, would eliminate the counterexample from the current model.

For explicit CI policy gating:

```sh
go run ./cmd/freecompliance prove --input examples/theorem/branch-controls.json --fail-on-counterexample
```

This intentionally exits nonzero. A modeled counterexample is not a confirmed policy violation or exploit.

## Logical semantics

For a finite state universe `Ω`, and admitted observations `E`, define the consistent worlds as:

`S(E) = {s in Ω : every admitted e in E holds in s}`

For a technical obligation `q` represented as a conjunction of equalities, return:

- `MODEL_SUPPORTED` if `S(E)` is nonempty and `q` holds in **every** state in `S(E)`.
- `MODEL_COUNTEREXAMPLE` if `S(E)` is nonempty and at least one state in `S(E)` violates `q`. The witness is a hypothetical countermodel, not an observed breach.
- `UNKNOWN` if no consistent state exists (contradictory admitted observations). Invalid schemas fail closed with an error.

This is a bounded **model-relative entailment test**, not proof of organizational compliance. The result can only be as comprehensive as the declared variables, values and assumptions.

**Evidence-plan semantics:** Evaluate all subsets of supplied candidate observations up to the hard cap, exclude subsets that leave no consistent state, and return the minimum summed cost among subsets that make every remaining state satisfy the obligation. The search is exact **only within the caller's finite candidate pool** and does not claim the actions are possible, valid, cost-effective in the real world, or sufficient outside this model. A contradictory existing verified observation cannot be silently overridden.

## What comes next

1. Independent source validators (GitHub REST effective rulesets, AWS/Azure/GCP permissions, real CI execution histories), evidence provenance and tamper-evident signed receipts.
2. Curated, reviewed, version-locked framework crosswalks to narrowly scoped and testable predicates. Distinguish free informational mappings from auditor-vetted obligations and licensed standards texts.
3. Multi-period operating evidence with coverage checks and reviewer identity validation.
4. Property-based, fuzz, contradictory-evidence and adversarial tests; published performance and soundness assumptions.
5. Optional boundary to BlockSiFr Control. **Never** turn a `MODEL_SUPPORTED` finding into a runtime authorization grant.

## Open-source boundary

This implementation is an independently authored public Go finite-model evaluator. It does **not** copy or publish the existing private BLAKE/CFAV engine, internal framework catalogs, production policy or BlockSiFr Control enforcement code. A production connector may use a stable public JSON contract without releasing those proprietary components.
