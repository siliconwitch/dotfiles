---
name: coding
description: Shared coding and review rules for C, Go, JavaScript, Lua, and similar imperative code. Use before writing or reviewing source code, tests, dependencies, module boundaries, names, layout, or comments. Apply coding-c or coding-go as an additional layer when relevant.
---

# Structure

- **Functional.** Values in and values out, immutable inputs, no hidden global
  state. Where hardware or a library demands state, follow its style and keep
  that state private to the unit that owns it.
- **Procedural.** A long function that reads top to bottom is the goal, not a
  compromise. Never shorten or split a function while it stays procedural.
- **Guard clauses.** Handle errors and edge cases first and return early, so
  the happy path stays unindented.
- **Inline single use.** A constant used once is a literal at the point of
  use. Logic used once is written where it runs, never behind a function
  called from one place.

# Architecture

Three layers. Dependencies point downward and may skip a layer: L1 may call L2
or L3, and L2 may call L3. Calls never point upward or sideways within a layer.

1. **Setup, surface, and run.** The entry point: `main.c`, `main.go`, or the
   top-level script. Initialises the system, starts and stops its work, exposes
   every feature, and directs control to the module that owns it. May contain
   the routing, composition, and lifecycle logic needed to hold the system
   together. Contains no business or application logic. A reader must be able
   to enumerate the system's features and find each L2 entry point from L1.
2. **Application logic.** One module per structural part of the system. Every
   decision that explains why a feature behaves as it does lives here. Never
   split a module or abstract over it.
3. **Workhorse.** The low-level mechanisms through which L2 acts, including
   drivers, network protocols, storage protocols, and hardware behaviour.
   Decisions here explain how a mechanism works, not why a feature behaves as
   it does.

Classify code by the question it answers:

- Where does a feature enter the system, and where is it sent? L1.
- Why does a feature behave this way? L2.
- How does the underlying mechanism work? L3.

A unit called only from L1 is not L3 merely because it looks infrastructural.
Place it in L2 when it implements application behaviour. Place it in L3 only
when it implements a low-level mechanism through which application logic acts.

A module is the unit the language encapsulates, which is not always a file: a
package in Go, a file and its header in C, a module in Lua. Several files in
one Go package are still one module, because nothing stops one reaching into
another. A layer split needs a boundary the compiler can refuse a call across.

Deleting a module and its one line of wiring at the setup layer removes that
feature entirely.

Helpers sit outside the layers: logging, error checking, and small
self-contained utilities, callable from anywhere.

A module reaching for another module in the same layer means the callee belongs
in L3. Move it down instead of calling sideways.

# Names and layout

- Meaningful names, never abbreviated: `plotWidth`, not `pw`. Short names only
  for loop indices, receivers, `err`, `ok`, and established mathematical
  notation.
- Group related statements into paragraphs with a blank line between them.
  Always a blank line before an error check.
- Never combine a call with its error test or its return.

# Comments

Write one only in these cases:

- A short single line heading breaking a long procedural function into
  sections. Says what the section is for, never how it works.
- A short inline note on a magic number:
  `i2c_write(0x12, 0x50); // LED_CFG0: 1A drive, active discharge enabled`
- One sentence on behaviour required by hardware or an external contract that
  cannot be seen in the code: a settling time, an erratum, a protocol ordering
  constraint.
- A block explaining unreadable or highly optimised code, prefixed
  `// IMPORTANT:`. Never write one unless asked, and never remove or reword
  one unless asked.

In code you are already editing, delete every other comment and replace it
with self describing code. Leave vendored code and untouched files alone.

# Tests

- Table driven: input to expected case, iterated in a loop. Never repeated
  near-identical assertions.
- Test the failure path, not only the completed operation.
- Never split a function or create single use code to make it testable. Test
  the logic as a block.
- Invented neutral identities in fixtures. Never a real name, address,
  credential or customer record.

# Dependencies

- No dependency for simple or single use logic.
- Never hand roll security, encryption, or a parser of untrusted input. A
  robust, well tested library belongs there.
- A dependency is right for a moving standard its maintainer tracks better
  than we can.
- Nothing a distribution's packager would balk at. Evaluate a candidate by
  running it, never from memory of how it behaves.
