# FreeCompliance 30-Day Enterprise Proof Challenge

**Category:** Proof-Driven Compliance. **Sponsor:** BlockSiFr. **Offer:** Free Apache-2.0 source and a separately scoped, optional professional-services proof engagement. **Status:** Pilot protocol, not independently validated product ROI.

## A decision, not a demo

**Can a compliance status be supported by source-bound evidence over the exact technical scope being claimed?** Compare FreeCompliance with an organization's current workflow and, only where the customer has legitimate access, its Vanta or Symetri deployment. No unlicensed competitive access, scraping, inference about competitor internals, or fabricated outputs.

**Customer outputs:** a versioned control-claim register, evidence provenance, theorem-supported and counterexample/unknown results, independently reviewed false-assurance findings, a prioritized evidence acquisition plan, and an actual time-on-task scorecard. All exports remain customer-controlled.

## Who participates

- Executive sponsor: CISO, Chief Compliance Officer, or Head of Internal Audit.
- Technical owner: GitHub and CI/CD administrator with least-privilege read-only access.
- Independent adjudicator: customer Internal Audit, assessor, or delegated reviewer who did not configure either tested tool.
- Security/privacy owner: approves scope, access, retention, and data export.
- Product operator: runs the same cases through FreeCompliance and existing systems, without changing production controls.

## 30-day execution

| Days | Work | Deliverables / exit gate |
|---|---|---|
| 1-3 | Written scope; 1 high-consequence workflow; approved connector privileges; agree control claims and authoritative standards versions | Signed scope and non-production verification plan |
| 4-7 | Record existing-tool baseline, existing manual work, and time logs; freeze at least 20 scoped technical cases | Adjudication sheet and baseline measurement |
| 8-14 | Ingest read-only GitHub evidence; create scoped theorem obligations; document source restrictions, invalidations and unknowns | Source-linked JSON plus bounded theorem results |
| 15-20 | Reproduce and review at least 5 intentionally controlled negative-path fixtures; review live-source exceptions only when authorized | Independently reviewed counterexamples or unknowns |
| 21-25 | Re-run same cases for each accessible platform; record human verification and remediation effort separately | Paired operator-measured CSV, evidence packets |
| 26-30 | Independent adjudication; calculate false-green rate, abstention, time reduction, and investigation workload | Signed customer acceptance, measured ROI narrative or explicit no-go |

20 cases and 5 fixtures are target sample sizes, not claims of statistical confidence. Any protocol deviation must be logged before interpreting results.

## Safety and evidence integrity

- No production configuration modifications, incident testing, exploit execution or write tokens in the core challenge.
- No vendor account access unless granted by the customer under its contract.
- Hash and timestamp the underlying inputs, but do not treat SHA-256 as source authentication or proof of operational effectiveness.
- A GitHub ruleset declaration proves the API reported a particular rule at the retrieval time. It does not prove every possible execution path, hidden bypass actor, actual deployment, or organization-wide compliance.
- Keep source-specific uncertainty. `UNKNOWN` is a safety-preserving result, not an automatic error or a pass.
- No SOC 2 opinion, FedRAMP authorization, CMMC assessment, or ISO certification is conferred.
- Do not publish customer data or name customer/competitor without written permission.

## Acceptance criteria

1. All included cases have documented scope, source references, reviewer, timestamps, and criterion version.
2. At least 95% of included records have a usable source reference; missing evidence is not silently counted as successful.
3. 100% of intentionally unsupported or unobservable claims in the negative fixture set produce either a specific modeled counterexample or `UNKNOWN`, never an unsupported `SUPPORTED`.
4. Every observed `SUPPORTED` assertion is independently checked against its exact claim and scope, with deviations flagged.
5. A paired subset of cases and fully disclosed reviewer minutes supports any comparison claim. If not paired, report only separate measurements.
6. Any estimated savings are calculated from actual time logs, burden rates approved by the customer, and implementation expenses.

Acceptance is technical completion, not a warranty of legal or regulatory compliance.

## Commercial structure (proposed, not published pricing)

- Open-source evaluation: $0 software license, customer-operated.
- Optional assisted 30-day proof engagement: **$25,000 fixed fee**, limited to one workflow, one primary evidence source, and a mutually agreed case set. Professional-service scope must be signed before kickoff.
- Post-pilot: customer can keep FreeCompliance indefinitely. Optional BlockSiFr Control proposal covers additional integrations, runtime authority and protected-action enforcement separately. No coercive license conversion.

## Decision at day 30

**Expand:** independent reviewers reproduce a material control blind spot or measurable assurance-efficiency gain without unacceptable false assurance. **Refine:** benefits exist but important coverage is missing. **Stop:** no independently supported advantage, excessive integration effort, or unacceptable operational risk.