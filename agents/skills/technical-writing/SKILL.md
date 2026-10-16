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

For comments, write one only in these cases:

- A heading over a section. It says what the section is for, never how it works.
- A short note beside a value or name whose purpose the code does not show, such as
  `i2c_write(0x12, 0x50); // LED_CFG0: 1A drive, active discharge enabled` or
  `herdr # terminal multiplexer`.
- One sentence on a constraint the code cannot express: a settling time, an ordering
  requirement, or another program's quirk.
- A block explaining unreadable or highly optimised code, prefixed `IMPORTANT:`. Never
  write one unless asked, and never remove or reword one unless asked.

Never say what the code does, narrate the edit, or justify its history. In a file you are
already editing, delete every other comment and replace it with self describing code.
Leave vendored code and untouched files alone.
