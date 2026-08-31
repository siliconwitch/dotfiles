---
name: response-format
description: Required response formats for user-facing replies. Use for every commentary and final response, except a greeting, acknowledgement, direct yes-or-no answer, or one or two factual sentences. Sub-agents do not need to apply this skill.
---

# User-facing response formats

Apply this section only to commentary and final messages sent to the user.

Use free-form prose only for a greeting, acknowledgement, direct yes-or-no answer, or one or two factual sentences. Use one of the following formats for every other response.

Use **Reporting work** when reporting actions, investigation, analysis, review, or changes. Use **Reporting options** when presenting ideas or choices. When a response both reports work and presents options, render the complete **Reporting work** contract followed by the complete **Reporting options** contract.

Each contract starts its `S`, `A`, `C`, `O`, and `N` counters at 1. Within a contract, each counter increments independently so the user can refer to one item directly.

## Reporting work

```markdown
**Short noun-phrase headline**

S1. What you understood the user to want, in one or two lines.

**Actions done**

A1. What you did and what verified it. Name the test, build, probe, source, or other check. Write "unverified" when no check ran. Include failures.

C1. Write each consequence directly beneath the action that caused it. Include side effects, changed behavior elsewhere, dependencies, restarts, and excluded scope. Omit consequences when none exist.

A2. Continue action numbering sequentially.

C2. Continue consequence numbering sequentially across all actions.

C3. Give one consequence per item.

**Next steps for you**

N1. Give one required user action per item. Give each command its own copyable code block. Write "None" when empty.
```

## Reporting options

```markdown
**Short noun-phrase headline**

S1. What decision, idea, or choice you understood the user to want.

**Options**

O1. State one option. Mark the recommended option.

C1. Write each consequence or tradeoff directly beneath the option that causes it. Omit consequences when none exist.

O2. Continue option numbering sequentially.

C2. Continue consequence numbering sequentially across all options.

C3. Give one consequence per item.

O3. An option may have no consequence items.
```
