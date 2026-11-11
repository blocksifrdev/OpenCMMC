# Optional BlockSiFr services from OpenCMMC

OpenCMMC is free and offline-first. Its local CMMC readiness workflow, evidence review, and JSON export **do not require** any paid product, account or subscription.

After a user creates a local assessment, the workbench offers **three separate, optional next steps**:

| Need | Service | What the link does |
| --- | --- | --- |
| Remediation in an AI coding session, with change-impact evidence | **CortexTrace** through BlockSiFr's software-delivery governance | Opens the public [software-delivery governance overview](https://exchange.blocksifr.com/solutions/software-delivery-governance) |
| Runtime limits and approvals before agents change systems | **BlockSiFr Control**, including runtime authority and execution receipts | Opens the public [Control overview](https://www.blocksifr.com/control) |
| Deeper vulnerability discovery when potential failures remain | **BLAKE**, consequence-focused security assurance | Opens the public [BlockSiFr overview](https://www.blocksifr.com/) |

## Explicit consent and honest product boundaries

- The optional cards appear **only after** the user creates a local assessment. They never block assessment creation, evidence attachment, local review or report export.
- Each card is a plain outbound HTTPS link. **No data is sent by OpenCMMC**. Only a deliberate browser click opens the separate site. Links carry no scope identifiers, repository names, report summaries, filenames, hashes, CUI, credentials, tokens or session details.
- The user's browser will make an ordinary website request only if they click. This is not a silent API call or a claim that a paid service was provisioned.
- CUI or sensitive evidence must **not** be shared with any connected service until the customer's own authorization and separate consent workflow is complete.
- There is **no embedded payment flow, trial clock, usage cap, registration wall, telemetry, automatic repository access, remote code execution or scan submission**.
- CortexTrace and runtime control are capabilities of the BlockSiFr commercial control stack. A link to an overview does **not** automatically activate either product.
- BLAKE offers deeper security discovery outside CMMC's local readiness review. No vulnerability result is implied and no live testing begins from these links.

## Prepare a private request

Each service card also opens an optional local preparation screen. After explicit confirmation, OpenCMMC downloads a JSON planning request with no assessment data. The file is not submitted and no paid service is activated. See [the local handoff guide](SERVICE_HANDOFF.md).

## Future integration contract

When a service-specific authenticated integration exists, it must use a separate, revocable **opt-in** action with a displayed data inventory, precise scope and recipient, evidence minimization, user approval and a local receipt. Default is **no transfer**. Treat URLs as navigation only until such an authorization flow is implemented and tested.

Review the [local UI](LOCAL_UI.md), [threat model](THREAT_MODEL.md) and [security policy](../SECURITY.md) before extending this workflow.
