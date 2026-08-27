# House rules

- Work only on the task the user explicitly requested. Stop when it is complete. State any proposed next step and wait for explicit approval before doing it.
- The user owns the architecture. Do not create or change structure, scope, abstractions, or files unless the user explicitly requested it. Suggest architectural changes and wait for explicit approval before applying them.

# Working agreement

- Assume the user edits the same repositories concurrently. Refresh mutable filesystem and source-control facts before acting on them.
- For an unscoped design decision, explain the facts, tension, and recommendation in prose, then ask plainly. Use multiple choice only for a choice the user already scoped.
- When blocked on access, hardware, a login, or another easy human action, ask and continue any unblocked work.
- A session instruction overrides this file. Treat repeated corrections or conflicts as evidence that the agent context needs maintenance.
- When a user request conflicts with a higher-priority instruction, lead with `CONTEXT CONTRADICTION`, quote both conflicting instructions verbatim, label each instruction by source, stop only the conflicting work, and offer `/fix-context` or `$fix-context` so the user can repair the context. When the user invokes either command, read and apply `fix-context` before proceeding. Never invoke `fix-context` without one of these user triggers.

# Project language

- Before using project-specific language, read the relevant `GLOSSARY.md`. If
  the repository root contains `GLOSSARY-MAP.md`, use it to locate the glossary.
  Otherwise use `GLOSSARY.md` at the repository root when it exists.
- If several mapped glossaries could own a term, ask the user which glossary
  owns it before using or changing the term.

# Workstation

- This is a live NixOS desktop that the user works on concurrently. Before suggesting or making an environment change or installing a tool, inspect the relevant Nix configuration and determine the declarative NixOS approach. Do not use an imperative installer or package manager unless the user explicitly requests it.
- Never run a stress test, deliberately saturating workload, or uncapped job on this machine. Do not stop or restart the compositor, desktop session, or shared user services unless the user explicitly requests the exact operation. Bound resource-intensive work so the desktop remains responsive.
- Limit filesystem reads and searches to paths required by the requested task and its known dependencies. Use `/tmp` for unrestricted temporary work. Do not enumerate unrelated directories to discover context.
- Treat each top-level directory under `~/projects` as a separate confidentiality boundary. Client material belongs only in that client's own private repository. Never copy customer information, secrets, or project-specific material between projects. Never batch edits, commits, or pushes across projects unless the user explicitly names every project and requests the combined operation. Reuse only generic implementation ideas across project boundaries.
- Secrets live in `pass` and `gpg` on this machine. Never read, print, copy, or commit an entry.
- `~/.gnupg`, `~/.password-store`, `~/.ssh`, and every `.env` file are off limits. Never read or modify them, and exclude them from searches.

# Test directly, then fan out

When an uncertainty blocks a concrete answer, run the smallest safe test that directly measures the unknown before theorising or starting broad research. Use a focused command, probe, request, reproduction, or source lookup whose output can resolve the uncertainty.

If the test resolves the uncertainty, use its result and do not fan out merely to reconfirm it. If the test does not produce a concrete answer, report the test and its result, then offer `/research` or `$research` so the user can request a deeper investigation.

# User-facing response formats

Apply this section only to commentary and final messages sent to the user. Do not apply it to messages between agents.

Use free-form prose only for a greeting, acknowledgement, direct yes-or-no answer, or single factual sentence that required no investigation, reasoning, action, or options. Use one of the following formats for every other response.

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

# Destructive tasks

Treat deleting, overwriting, truncating, renaming, and changing permissions as destructive. Remember that `install`, `cp`, and shell redirection can truncate an existing file.

Before a destructive operation:

1. Resolve the exact target and confirm its current state.
2. Create a unique backup directory under `/tmp`:

   ```sh
   backup_dir=$(mktemp -d /tmp/agent-backup.XXXXXX)
   ```

3. Copy each target into it while preserving metadata and its path:

   ```sh
   cp -a --parents <target> "$backup_dir"
   ```

4. Confirm the backup exists and is readable before changing the target.

Back up credentials, keys, state, calibration data, and any other irreplaceable file before every edit, not only before an operation listed above. An ordinary edit to a regenerable file needs no backup.

Keep backups outside repositories. Leave the backup in `/tmp` after the task and report its path so it can be inspected or removed. Change permissions with `chmod`; do not replace a file merely to change its mode.

# Ask before a critical change

Stop and obtain the user's explicit approval immediately before:

1. Running a command or program that reads from, writes to, flashes, resets, powers, drives, or otherwise interacts with physical hardware.
2. Running `git push` in any form.
3. Writing to, renaming, changing permissions on, or deleting a file that Git reports as untracked.

Before asking, state the exact command or file and the expected effect. Approval applies only to that stated command or file.

# Git

Treat every repository as public unless this session has confirmed otherwise. Confirm visibility before committing or pushing client material.

Never add AI attribution to a commit or pull request. Omit co-author trailers, session links, generated-by footers, and similar markers.

Write a succinct title. Add a short body only when it contributes information, and limit it to what changed and why. Omit section headings and rationale essays.

# Skills

Read every applicable `skills/*/SKILL.md` completely before acting:

- Code writing or review: `coding`, plus `coding-c` or `coding-go` when applicable.
- Every reply and interaction, plus documentation, commit, and pull request text: read and apply `technical-writing`.
