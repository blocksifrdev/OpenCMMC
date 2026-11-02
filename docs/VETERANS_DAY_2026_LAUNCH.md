# Veterans Day 2026 | Launch kit

**Planned public announcement:** Wednesday, November 11, 2026

## Campaign identity

**Name:** FreeCompliance Veterans Day Service Initiative

**Headline:** Built in gratitude. Given in service.

**Product statement:** Stop guessing about CMMC. Start proving your controls.

**Supporting statement:** Know what's required. See what's documented. Find what's missing.

**Call to action:** Start free on GitHub: https://github.com/blocksifrdev/freecompliance

**Audience:** Defense Industrial Base contractors, assessors, small business assistance programs, and government technical teams.

## Founder announcement (ready to adapt and post on November 11)

This Veterans Day, I am honoring my father's service as a lieutenant colonel in the U.S. Army in a way that reflects what I have spent my career building.

We are making FreeCompliance available as a permanent, open-source contribution to the defense community.

Small defense contractors should not need a large compliance budget just to organize CMMC evidence, understand what remains undocumented, or prepare for a review. And organizations working in disconnected environments should not have to upload sensitive evidence to a commercial cloud to get started.

FreeCompliance provides a free offline CMMC evidence workbench, local evidence-integrity checks, and open-source assessment reporting. It is not a government certification, a substitute for official assessments, or a government-endorsed product.

Our commitment is simple: **Compliance should be proven, not presumed. The tools to build that proof should be free.**

To everyone who has served, and every family that supports them: thank you.

Start here: https://github.com/blocksifrdev/freecompliance

## 60-second demo script

1. Show an entirely disconnected laptop or isolated virtual machine.
2. Run `freecompliance cmmc init --level 2 --scope "Demo CUI Enclave" --out workspace.json`.
3. Show 110 Level 2 identifiers, all **UNASSESSED**.
4. Associate a fictional evidence file with a fictional reviewer claim. Show integrity validation only, not a compliance pass.
5. Run `freecompliance cmmc assess --input workspace.json --evidence-dir evidence --json readiness.json`.
6. Open the JSON report. Explain DOCUMENTED versus UNASSESSED, GAP, and UNVERIFIABLE.
7. State clearly: no credential, cloud upload, subscription, government submission, or false certification claim.
8. Close with the roadmap for complete NIST 800-171A objective coverage and source-bound proof tests.

## Launch checklist

- [ ] Run Go tests, race detection, vet, and offline smoke tests on the release candidate.
- [ ] Review CMMC program guidance and qualification caveats as of launch day.
- [ ] Ensure a versioned, immutable source release with SHA-256 hashes and build instructions exists.
- [ ] Validate documentation against implemented CLI commands and output statuses.
- [ ] Publish synthetic proof demo without CUI, real secrets, or customer names.
- [ ] Ask at least three independent DIB practitioners to trial the offline walkthrough.
- [ ] State no government endorsement, authorization, accredited assessment, SPRS submission, or verified enterprise adoption.
- [ ] Publish the founder dedication on November 11 after completing the release checks.

## Performance indicators

Track signed-up community contributors, independently reproducible installs, offline smoke-test success, objective-mapping completion, reviewer findings, time saved in evidence preparation, and qualified government or prime-contractor technical evaluations. Do not represent GitHub stars as proof of CMMC compliance or accepted government deployment.

## Licensing and long-term economics

Core FreeCompliance remains Apache-2.0 open source. Enterprise BlockSiFr Control is optional and separate. No time-limited free tier, paid unlock, or mandatory identity registration is added for the Veterans Day campaign.
