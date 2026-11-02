# Independent Comparison Measurement Protocol v1

This protocol supports a controlled comparison of FreeCompliance, a manual baseline, and optionally Vanta or Symetri systems the customer is authorized to use. It is not a certification, published vendor benchmark, or accuracy guarantee.

## Pre-registration

Before any scoring, document the workflow, tool versions, control versions, sampling criteria, access to each vendor, named independent reviewer, run timestamps, and whether real production evidence or deliberately synthetic fixtures are used. The same 20+ selected technical claims and at least 5 negative-path fixtures should be evaluated for every system where feasible. Record exclusions. Do not cherry-pick only cases where a product performs well.

## Standard case schema

For the public Go scorer, collect one CSV row per pair of `case_id` and `system`:

| Field | Definition |
|---|---|
| `case_id` | Stable pseudonymous case identifier shared across systems |
| `system` | Tested workflow or product. Any name permitted. |
| `minutes` | Human labor minutes for evidence preparation and review associated with this case; do **not** include machine compute or subscriptions |
| `expected` | `SUPPORTED`, `COUNTEREXAMPLE`, or `UNDETERMINED`, determined by independent adjudication |
| `observed` | `SUPPORTED`, `COUNTEREXAMPLE`, or `UNKNOWN`, mapped from product output in a published, reviewed translation rubric |
| `reviewed` | True only where an independent reviewer reached a supported expected determination |
| `reference` | Identifier of the controlled, access-restricted evidence and adjudication packet |

An unreviewed case must use `expected=UNDETERMINED`. Do not invent a ground-truth determination. A `SUPPORTED` observation must be scoped to one precise technical assertion, not a framework-wide certification. `COUNTEREXAMPLE` requires a concrete modeled witness or source-backed contradiction. Generic product "risk" alerts are not automatically counterexamples.

## Metrics and denominators

- **False-green:** independently adjudicated `COUNTEREXAMPLE` cases reported `SUPPORTED` by a system. Denominator is *all independently adjudicated counterexample cases for that system*, including those reported unknown. An unknown is not false-green, but is separately disclosed.
- **Correctly supported:** independently adjudicated `SUPPORTED` cases reported supported.
- **Counterexample detection:** independently adjudicated `COUNTEREXAMPLE` cases reported counterexample.
- **Abstention:** observed `UNKNOWN` across all cases, with source and reason. Do not suppress abstention to improve apparent accuracy.
- **Human time:** sum of verified minutes and notes, broken down by evidence collection, independent review, remediation and rework in the underlying packet. The compact CSV intentionally aggregates the first two only.
- **Paired comparison:** use exactly matched `case_id` across systems. If `all_systems_paired=false`, or if source scope differs materially, do not report a comparative aggregate.
- **Dollar impact:** measured human hours avoided times approved fully loaded hourly rate, minus implementation and review expense. Never equate savings with reduction in audit fees unless contractually confirmed.

These are operational test metrics, not a scientifically calibrated population-wide false-positive rate. Include confidence intervals and larger randomized samples before making general market performance claims.

## Run the scorer

```bash
go run ./cmd/pilotbench --input examples/pilot/synthetic_benchmark.csv
```

The provided sample is **synthetic and illustrative**, never a Vanta or Symetri assessment. Use `examples/pilot/blank_benchmark.csv` to start a real customer benchmark. The program validates duplicate cases, contradictory independent reviews, malformed times, and missing provenance references. JSON results show independent adjudication coverage and matched-case counts, and leave false-green rates absent if no verified negative cases exist.

## Adjudication independence

The reviewer must inspect the exact underlying evidence and validate the claimed scope rather than accept any vendor's green status at face value. Preserve the raw vendor output in the restricted packet; include the mapping between vendor-specific statuses and the three technical states. Assessor disputes are unresolved, not silently resolved in FreeCompliance's favor.

## Privacy and publication

Use pseudonymous identifiers, read-only evidence, encrypted customer-side storage, and a contractual retention period. The CSV should not include credentials, customer names, source code, or sensitive finding details. Public case studies require separate written approval and must publish the denominator, tested versions, tool limitations, and whether a result was synthetic or production-derived.