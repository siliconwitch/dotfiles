---
name: fix-context
description: User-invoked repair of AGENTS.md files and skills from observed friction. Use only when the user invokes /fix-context or $fix-context.
disable-model-invocation: true
---

# Fix context

1. Read the complete agent-facing files affected by the request before proposing changes.
2. Identify the exact behavior each instruction must cause.
3. Keep only instructions that specify an observable action, constraint, trigger, or completion condition. Rewrite or omit ambiguous instructions.
4. Keep each instruction in one authoritative location. Report duplication before removing it.
5. Present proposed structural or placement changes to the user and wait for approval.
6. Apply only the approved changes.
7. Validate every changed skill and inspect every changed agent-facing file for placeholders, contradictions, and stale references.

## Maintaining agent-facing files

- Assume small, low-cost models will read the instructions. Use short, direct sentences. Put each trigger beside its required action. State exceptions explicitly. Avoid rules that require combining implications from separate sections.
- Do not copy facts that an agent can recover cheaply from current source, configuration, directory structure, or `--help` output. Record only conventions, reasons, constraints, and traps that the environment does not reveal.
- Make a skill user-invoked when only the user should choose it. Make it model-invoked only when the agent must discover it independently. For a model-invoked skill, state every distinct trigger once in its description.

## Maintaining AGENTS.md

- Keep a rule in `AGENTS.md` only when it must remain available across tasks because it is universal, frequent, or safety-critical. Propose moving task-specific procedures into skills and wait for approval.
- Write every skill pointer as an explicit trigger followed by the action the agent must take.
- Confirm every referenced skill and file exists at the stated path.
- Compare each proposed rule with existing skills and instructions. Report duplicated meaning and identify the authoritative location before removing or rewriting it.
