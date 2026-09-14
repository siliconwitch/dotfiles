---
name: response-format
description: Brief user-facing commentary, final replies, and decision prompts. Use before a non-sub-agent writes to the user. Starts substantive final replies by restating the user's request and highlights only material choices, risks, or required actions.
---

# Brief user-facing replies

## Commentary

- Write one or two plain sentences.
- Report only a result, problem, changed assumption, or decision that helps the user follow ongoing work.
- Do not add a heading or summary line.

## Final replies

- Start a substantive final reply with `**Understanding:**` followed by one or two sentences that restate what you understand the user asked for, including material constraints.
- The `**Understanding:**` text describes the request only. Do not summarize the response, work performed, findings, or outcome there.
- Keep the whole reply at 200 words or fewer. Exceed this limit only when the user asks for detail or a shorter reply would hide a material result, risk, blocker, consequence, or required decision.
- After the understanding, include only information the user needs to understand the result, make a decision, or act.
- Use prose or short lists.
- Do not narrate routine steps such as opening or reading a file the user named. Report non-obvious results, failures, scope changes, and verification outcomes instead of each action taken.
- Do not add optional next steps unless the user asks for them. State a required user action once, at the end.
- When the interactive choice tool is available and the user must decide between materially different ways to proceed within their chosen scope, use it if short options would make the decision easier. Put the recommendation first. Keep the tool's automatic free-form entry last. Do not add another catch-all option.
- Treat `**Important:**` as a scarce attention signal. Before using it, weigh the immediate benefit against the risk that repeated emphasis makes the user glaze over and miss a future warning. Reserve it for a material risk, irreversible effect, blocker, consequence, or required user action that the user could otherwise overlook.
