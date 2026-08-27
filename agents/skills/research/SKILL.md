---
name: research
description: User-invoked multi-model research, adversarial fact verification, and evidence arbitration. Use only when the user invokes /research or $research.
disable-model-invocation: true
---

# Research

Run six phases in order. Do not merge phases or let an agent verify, judge, or arbitrate its own work.

Keep every agent non-mutating: do not edit project files or external state. Put temporary test artifacts under `/tmp`. Launch agents within a phase concurrently. Wait for the phase to finish before starting the next one.

Use different model families within each phase when the harness exposes model choice. Otherwise use distinct agents on the available model and disclose the missing model diversity.

## 1. Frame

1. State the question, its boundaries, and what evidence would answer it.
2. Define a material claim as one whose falsity would change the final answer, recommendation, risk assessment, or required action.
3. Split the question into two to four independent research lanes. Assign one investigator to each lane. If the question has only one lane, assign two investigators with different evidence strategies.
4. Give every investigator its lane, the complete user question, and the same evidence rules. Do not give it another investigator's conclusions.

## 2. Investigate

Each investigator must:

1. Look locally first: project documentation, vendored references, submodule headers, source, and generated help.
2. Follow claims to the source that owns them: official documentation, upstream source, standards, datasheets, or first-party APIs.
3. Search for counterevidence and incompatible interpretations.
4. Return each material claim with its source, supporting evidence, limits, counterevidence, and verification status.
5. Work out derived claims explicitly.
6. Label unsupported claims `unverified`.

For facts in a PDF table or diagram, preserve visual structure. Confirm a pin number, register field, limit, or timing value by rendering the page or using another method in addition to raw text extraction.

Prefer current sources when behavior can change. Record access or publication dates when they affect interpretation.

## 3. Verify

After every investigator returns:

1. Build one ledger of all material claims. Preserve disagreements instead of resolving them by vote.
2. Launch at least two verifier agents. Assign every material claim to two verifiers that did not originate it.
3. Require each verifier to open the cited source, confirm that it supports the claim, check whether it is current and authoritative, search for contrary evidence, and reproduce any safe test or derivation.
4. Mark each checked claim `confirmed`, `disputed`, or `unverified`. Attach the evidence and criticism that produced the verdict.

## 4. Arena

Run an arena when any condition holds:

1. The user explicitly requests an arena, competing answers, or adversarial research.
2. The answer guides an action where an error could cause physical harm, a security or privacy breach, legal or medical harm, difficult-to-reverse financial loss, unrecoverable data loss, hardware damage, credential compromise, or a production outage.
3. A material claim remains disputed because primary sources conflict, independent agents reach incompatible conclusions, or no verifier can confirm it.

Do not run an arena merely because the question is broad, complex, or has many sources.

For each arena question:

1. Write one prompt containing the exact question, scope, evidence standard, and three to six gradeable criteria.
2. Launch at least three isolated candidates on the same prompt. Use different model families when available. Do not show candidates earlier conclusions or each other's work.
3. Require every candidate to return its answer, primary evidence, counterevidence, rejected interpretations, uncertainty, and rationale.
4. Treat convergence as a lead, not proof. Treat divergence as a reason to inspect the framing and evidence, not to average the answers.

## 5. Arbitrate

After all arena candidates return:

1. Launch an evidence arbiter that did not investigate, verify, or compete. Give it the raw candidate outputs, claim ledger, criticisms, and source locations. Require it to reopen decisive sources, run safe tests, test each criticism, and issue an evidence verdict for every disputed claim.
2. After the evidence arbiter returns, launch a synthesis arbiter on a different model family when available. Give it all raw outputs and the evidence verdict. Require it to reopen decisive sources and formulate one set of final discoveries without resolving uncertainty by vote.
3. If the arbiters disagree, return the disputed claim to a new evidence arbiter. Do not present it as settled until evidence resolves it.

## 6. Report

Before answering, open and spot-check every source supporting a decisive claim.

Present:

1. The direct answer and final discoveries.
2. Citations beside the claims they support.
3. Confirmed, disputed, and unverified claims under distinct labels.
4. Material counterevidence and rejected interpretations.
5. Any agent dropout, unavailable model family, untested criticism, or other coverage gap.

Do not include agent transcripts or majority votes as evidence.

If an agent drops out, retry once with a different agent. Run extra waves sequentially when concurrency is limited. Do not reduce the required two investigators, two verifiers, three arena candidates, evidence arbiter, or synthesis arbiter.
