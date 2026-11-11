# Veterans Day 2026 | Launch kit

**Public early access announced:** Monday, November 2, 2026  
**Official Veterans Day launch:** Wednesday, November 11, 2026  
**Available to use now:** https://github.com/blocksifrdev/OpenCMMC

## Campaign identity

**Name:** OpenCMMC Veterans Day Service Initiative

**Headline:** Built in gratitude. Given in service.

**Product statement:** Stop guessing about CMMC. Start proving your controls.

**Supporting statement:** Know what's required. See what's documented. Find what's missing.

**Call to action:** Start free on GitHub: https://github.com/blocksifrdev/OpenCMMC

**Audience:** Defense Industrial Base contractors, assessors, small business assistance programs, and government technical teams.

## Early-access announcement (November 2, 2026)

**OpenCMMC is available now. Our official Veterans Day launch is November 11.**

As the son of a U.S. Army lieutenant colonel, I wanted BlockSiFr to honor military service with something lasting and useful to the defense community.

Today, we are opening OpenCMMC for anyone to use, test, and improve. It is free, open source, and designed to work offline: an initial CMMC evidence workbench to help contractors identify requirements, document what they have, and see what still needs review.

No subscription. No login. No mandatory cloud service.

This is a readiness utility, not CMMC certification or a government-endorsed program. Full objective-by-objective verification and official scoring remain engineering milestones.

**Access it today:** https://github.com/blocksifrdev/OpenCMMC

**Official launch:** Veterans Day, November 11, 2026.

## Veterans Day founder announcement (scheduled for November 11, not yet posted)

This Veterans Day, I am honoring my father's service as a lieutenant colonel in the U.S. Army in a way that reflects what I have spent my career building.

We have already made OpenCMMC publicly available as a permanent, open-source contribution to the defense community. Today we formally launch the Veterans Day service initiative.

Small defense contractors should not need a large compliance budget just to organize CMMC evidence, understand what remains undocumented, or prepare for a review. And organizations working in disconnected environments should not have to upload sensitive evidence to a commercial cloud to get started.

OpenCMMC provides a free offline CMMC evidence workbench, local evidence-integrity checks, and open-source assessment reporting. It is not a government certification, a substitute for official assessments, or a government-endorsed product.

Our commitment is simple: **Compliance should be proven, not presumed. The tools to build that proof should be free.**

To everyone who has served, and every family that supports them: thank you.

Start here: https://github.com/blocksifrdev/OpenCMMC

## 60-second demo script

1. Show an entirely disconnected laptop or isolated virtual machine.
2. Run `opencmmc cmmc init --level 2 --scope "Demo CUI Enclave" --out workspace.json`.
3. Show 110 Level 2 identifiers, all **UNASSESSED**.
4. Associate a fictional evidence file with a fictional reviewer claim. Show integrity validation only, not a compliance pass.
5. Run `opencmmc cmmc assess --input workspace.json --evidence-dir evidence --json readiness.json`.
6. Open the JSON report. Explain DOCUMENTED versus UNASSESSED, GAP, and UNVERIFIABLE.
7. State clearly: no credential, cloud upload, subscription, government submission, or false certification claim.
8. Close with the roadmap for complete NIST 800-171A objective coverage and source-bound proof tests.

> **Release accuracy:** The source and historical preview downloads are public. Do not call a full v1.0.0 release verified or published until all release gates have passed and GitHub publication has occurred.

## Launch checklist

- [ ] Run Go tests, race detection, vet, and offline smoke tests on the release candidate.
- [ ] Review CMMC program guidance and qualification caveats as of launch day.
- [ ] Ensure a versioned, immutable source release with SHA-256 hashes and build instructions exists.
- [ ] Validate documentation against implemented CLI commands and output statuses.
- [ ] Publish synthetic proof demo without CUI, real secrets, or customer names.
- [ ] Ask at least three independent DIB practitioners to trial the offline walkthrough.
- [ ] State no government endorsement, authorization, accredited assessment, SPRS submission, or verified enterprise adoption.
- [ ] Publish the official Veterans Day founder message on November 11, after completing the release checks. Public access is already open.

## Performance indicators

Track signed-up community contributors, independently reproducible installs, offline smoke-test success, objective-mapping completion, reviewer findings, time saved in evidence preparation, and qualified government or prime-contractor technical evaluations. Do not represent GitHub stars as proof of CMMC compliance or accepted government deployment.

## Licensing and long-term economics

Core OpenCMMC remains Apache-2.0 open source. Enterprise BlockSiFr Control is optional and separate. No time-limited free tier, paid unlock, or mandatory identity registration is added for the Veterans Day campaign.
