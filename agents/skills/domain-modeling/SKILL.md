---
name: domain-modeling
description: User-invoked creation or revision of a project's GLOSSARY.md domain glossary. Use only when the user invokes /domain-modeling or $domain-modeling.
disable-model-invocation: true
---

# Domain modeling

Create or revise the project's domain glossary. Keep `GLOSSARY.md` a glossary,
not a specification, design document, scratch pad, or record of implementation
and architecture decisions.

## Locate the glossary

1. Follow the project language rules in `AGENTS.md` to locate the relevant
   `GLOSSARY.md`.
2. Do not create `GLOSSARY-MAP.md` or split a repository into multiple
   glossaries unless the user explicitly requests that structure.

## Resolve the language

1. Read the relevant existing glossary, code, and documentation.
2. Collect only domain concepts specific to the project. Exclude general
   programming concepts, libraries, utilities, and implementation mechanisms.
3. Identify vague, overloaded, conflicting, and synonymous terms. Propose one
   precise canonical term for each concept.
4. Surface every conflict between the proposed meaning, existing glossary,
   documentation, and code. Treat code as a consistency check, not proof of
   intended meaning.
5. Test unclear concepts with concrete scenarios that expose edge cases and
   boundaries.
6. Ask the user to resolve every ambiguity that would change a term, meaning,
   owner, or boundary. Do not choose an unscoped domain definition for them.

## Write the glossary

Write only after the user asks for the file change or approves the proposed
entries. Create `GLOSSARY.md` only when there is at least one resolved term to
record. When editing an existing glossary, change only the approved entries.

Use this format:

```md
# {Context name}

{One or two sentences describing what this context is and why it exists.}

## Language

**{Canonical term}**:
{One or two sentences defining what the term is, not what it does.}
_Avoid_: {Discouraged synonym}, {discouraged synonym}
```

- Use one canonical term for each concept.
- Keep every definition to one or two sentences.
- Omit `_Avoid_` when there are no discouraged synonyms.
- Add subheadings only when terms form clear natural groups. Otherwise keep one
  flat language list.
