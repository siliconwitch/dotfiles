---
name: coding-c
description: C, header, Makefile, linker-script, embedded, and hardware-test rules. Use with the coding skill when writing or reviewing .c, .h, Makefile, .mk, or .ld files, or when configuring C warnings, analysis, dependency generation, or hardware-facing tests.
---

# C

- Check every return value from the kernel, filesystem, flash, crypto, radio
  and hardware libraries. Timeouts, short reads and writes, malformed frames
  and exhausted storage are ordinary states.
- Avoid the preprocessor. `#define` is for hardware pin numbers and physical
  addresses, nothing else. A macro is correct only where a function cannot do
  the job, such as capturing the call site:
  `#define ERROR_CHECK_NON_ZERO(e) error_check_non_zero(e, __FILE__, __LINE__)`

# Memory safety

CERT C at https://cmu-sei.github.io/secure-coding-standards is the reference
for undefined behaviour and memory safety. Cite the rule ID when one applies,
such as `INT30-C` for unsigned wraparound.

- Bounded memory. Prefer static storage and fixed capacity queues. Dynamic
  allocation needs a demonstrated bound and a failure path.
- Fixed width integer types: `uint32_t`, never `unsigned long`.
- Ownership, capacity and encoded length are three separate facts, and all
  three stay visible at every boundary. Never infer one from another. A buffer
  crosses as a pointer plus its capacity, and the call reports how many bytes
  it actually wrote.
- Bounded calls at every boundary: `strnlen` over `strlen`, a checked length
  on every copy, no unbounded read or write.
- No undefined behaviour: no uninitialised reads, out of bounds access, signed
  overflow, invalid shifts, unchecked narrowing, dangling pointers or races.

# Build tooling

`make` is the entry point, even where it wraps an SDK's own build system. For
first-party bare-metal code the full warning set, `-Werror` and `-fanalyzer`
live in the compile rule, so every build is the analysis run. Set those up when
standing up a new project, and never loosen what an existing project already
checks. Keep vendored code quiet with `-w` and `-isystem` so it cannot dirty
our diagnostics.

Formatting is the editor's job on save, and clangd reads a compile database the
build already produces. Never add a make target for formatting, linting or a
compile database.

Traps proven the hard way, so do not relearn them:

- `gcc -fanalyzer -fsyntax-only` analyses nothing and exits 0. `-fanalyzer`
  needs a real `-c` compile.
- `-fanalyzer` cannot see through a naked function that returns in r0, so
  branching on its return raises a false
  `-Wanalyzer-use-of-uninitialized-value`. Suppress it for that object only,
  never globally.
- `clangd --check` prints errors only. It reports zero errors on a file that a
  live LSP session floods with warnings, so validate diagnostics with a real
  didOpen session.
- clangd does not run path-sensitive `clang-analyzer-*` even when
  `.clang-tidy` lists it. The editor cannot catch what `-fanalyzer` catches.
- A rule above `all:` steals the default goal. Keep `$(OBJECTS): Makefile` and
  `-include *.d` at the bottom of the Makefile. Target-specific variable
  assignments are fine anywhere.
- Per-object compiling needs `-MMD -MP` and `-include $(OBJECTS:.o=.d)`, or a
  header edit yields a stale binary that `make flash` happily programs.
- Keep the check list in `.clang-tidy` alone. `.clangd` does CompileFlags
  only, and its `ClangTidy: Add:` re-enables globs that `.clang-tidy` negates.
- `bugprone-reserved-identifier` fires on linker symbols such as `__bss_end`
  and `__STACK_BASE`. Negate it, because the linker script and the vendor MDK
  dictate those names.

# Testing on hardware

For hardware-touching code, write tests that drive the real device or the real
output rather than asserting on intermediate bytes. A printed label proves more
than the PDF bytes behind it.

The build, flash and debug recipe is a project fact, and so is the SoC. Read
the project's own AGENTS.md or CLAUDE.md and its Makefile. Never carry a trap across from
another project: RRAM, SoftDevice programming and RTT capture all differ.
