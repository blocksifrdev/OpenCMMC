# Why Proof-Driven Compliance is a distinct buying decision

**BlockSiFr OpenCMMC** is an Apache-2.0 proof-driven compliance utility. It is **not** a replacement for all governance, audit, risk and compliance management workflows. The open-source wedge is narrower and more falsifiable: evaluate whether admissible source evidence supports a precisely scoped technical claim, find modeled counterexamples, and disclose uncertainty.

## What the other vendors actually offer

- **Vanta** publicly documents automated evidence collection, continuous controls monitoring, custom controls and tests, AI-assisted workflows, audit services and risk management. Sources: https://www.vanta.com/products/automated-compliance and https://www.vanta.com/trust-management-platform
- **Symetri** publicly documents multi-cloud automated evidence collection, continuous control testing, policy-as-code, remediation, OSCAL-based SSPs and compliance workflow management. Sources: https://symetri.io/product/ and https://symetri.io/symetri-core/

Do **not** say these products only collect documents or cannot validate controls. Do **not** claim that either has no proof/logic capability without a controlled evaluation. Public product descriptions do not establish absence of features.

## What OpenCMMC can challenge

| Buyer question | Evaluation method | Current maturity |
|---|---|---|
| Does an evidence artifact exist? | Local scanner with SHA-256 and provenance | Implemented |
| Does a precisely scoped modeled claim follow from admissible evidence? | Bounded finite-world theorem evaluator | Experimental v0.3 |
| Do active GitHub rule declarations support a narrow claim? | Read-only, source-bound ruleset evaluator | v0.4 source-bound feature |
| Can a passing configuration hide a bypass? | Controlled negative-path fixture and source review | Model-relative; production proof not established |
| Can we compare value against an incumbent fairly? | Adjudicated paired-case time and accuracy protocol | Pilot measurement kit |
| Can consequential actions be stopped at runtime? | Separate BlockSiFr Control authority integration | Not in OpenCMMC |

## Buying wedge

Give us one **consequential action that must never execute incorrectly**. Choose one associated control claim, enumerate the actual evidence, and determine whether the claim can be supported over the declared model. If it cannot, show why. An honest `UNKNOWN` is a useful outcome. Do not claim the model is complete when its coverage is unknown.

## Three competitive plays

1. **Coexist first.** Export Vanta or Symetri evidence through approved customer workflows, or use the underlying GitHub/cloud sources directly. OpenCMMC is a second-look control-claim verifier, not necessarily a rip-and-replace purchase.
2. **Prove independent value.** Run the 30-day benchmark with a customer adjudicator. Show exact claims, counterexamples, stale-source invalidation, missing evidence and human time. Publish no winning percentages until real measurements exist.
3. **Convert only for runtime control.** OpenCMMC remains free. Enterprise BlockSiFr Control is optional for protected actions, runtime authorization, policy decisions and execution receipts. Do not label local hash digests as signed receipts.

## Positioning

**Category:** Proof-Driven Compliance. **Message:** "Don't tell me a control is passing. Show me what the evidence can prove." **Commercial extension:** "When a consequential action is about to happen, BlockSiFr checks its authority before execution."

## Proof burden before public comparative claims

Validate target competitors' exact SKUs and versions; use written customer authorization; evaluate on matched, independently adjudicated cases; preserve negative and unknown results; disclose all denominators and testing conditions; obtain counsel review for claims about named competitors. No comparisons in this document constitute measured superiority.