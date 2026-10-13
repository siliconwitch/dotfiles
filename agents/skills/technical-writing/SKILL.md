---
name: technical-writing
description: Concise technical prose for replies, documentation, specs, skills, AGENTS.md, CLAUDE.md, READMEs, comments, commit messages and pull requests. Use whenever writing or reviewing prose for engineers.
---

# Technical writing

Write for a tired engineer reading once.

- Cut filler. Use the shortest everyday word that stays precise.
- Use active voice and present tense.
- Be concrete. Name the actual things instead of abstract categories. Write
  "a test or network call that might hang", not "every command and wait that
  can block".
- Name real files, symbols, flags, and commands. Use one name for each thing.
- Keep one thought or instruction per sentence. Put conditions before the action they govern.
- Give an example for any rule that isn't obvious.
- Say why when the reason isn't obvious.
- Write a calculation as a formula, such as `remaining = target - progress`.
- Avoid new terms. When you need one, explain it in plain words where it
  first appears.
- Use commands for procedures. Give every required action its own list item and self-contained command block.
- Use numbered lists for sequences and bullets for unordered sets.
- Make headings carry the point. Use sentence case.
- Prefer periods to semicolons. Never use an em dash.
- Avoid hype, filler, hedging, invented metaphors, agent narration, and claims of simplicity.
- Don't copy the style of a dense existing document.
- State what the system does now. When documentation and code disagree, treat the documentation as wrong.
- Do not explain what the code or another document already makes clear.

For plans, designs, and specifications, write only what the user asked for. Add no feature, option or edge case the user did not ask for.

For commits and pull requests, omit AI attribution and generated-by markers. Use a succinct title and a short body that says what changed and why. Add no section headers unless the project requires them.

For comments, describe a contract, constraint, or reason that the code cannot express. Do not narrate the edit or justify its history.
