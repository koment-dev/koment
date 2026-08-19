# 0132 — Comment detection covers every file, with an opinionated catch-all

Date: 2026-08-09
Status: Accepted

## Context

koment's comment gate has only ever run on Go. `detectors` held one entry,
`Detects(file)` returned false for everything else, and `sourceFiles` used that
to decide what to walk:

```go
if !entry.IsDir() && Detects(file) && !configured.Excludes(file) {
```

The consequence is not "YAML is unsupported". It is that **the gate reports
success on a repository it never examined**. Demonstrated on a scratch
repository holding three files, each with a textbook prohibited comment:

```
# We retry three times here because the upstream API is flaky and
# occasionally 500s under load. Do not lower this.
retries: 5
```

plus the Python and Rust equivalents. `koment comments check` printed
`comment policy: ok` and exited 0.

That is the failure koment exists to prevent, committed by koment. A YAML-only
repository — `home-ops` is one — gets a green gate and a managed `AGENTS.md`
telling agents never to write comments, with nothing enforcing it. Issue #89
filed this as "no `comments convert` for YAML"; the convert gap is a symptom of
the check gap underneath it.

It also breaks the product thesis. koment's value for an agent is a fast signal
that it just wrote a comment it should not have. Today that signal fires only
in `.go` files.

ADR 0114 already anticipated this: the `Detector` interface exists "so a second
language is an implementation rather than a rewrite". This ADR is that second
implementation, and it is deliberately not one language.

## Decision

Two detectors, in priority order.

**`goDetector` is unchanged** and still matches first. `go/parser` knows which
comments are godoc on exported identifiers, which no marker scan can infer.
Where a real parser exists and already earns its place, koment keeps it.

**`markerDetector` handles everything else** from a table of comment syntaxes
keyed by file extension — line markers and block delimiters for the priority
languages (YAML, TOML, JS, TS, Rust, C, C++, Java, Python, Lua) and the config
and shell formats that surround them. A file whose extension is not in the
table still gets scanned, using a fallback marker set of `#`, `//` and `--`.

The catch-all is opinionated, and these are the opinions:

- **A comment must own its line, or follow code after whitespace.** A marker
  found mid-token is not a comment. Before treating a trailing marker as one,
  koment counts unescaped quotes earlier on the line; an odd count means the
  marker sits inside a string and is skipped. This is what keeps a URL in a
  YAML value from being read as a `//` comment.
- **Consecutive line comments are one group**, exactly as `go/parser` groups
  them, so a rationale paragraph is one violation and one excerpt rather than
  five.
- **Prose and data formats are excluded outright.** Markdown, reStructuredText,
  plain text and JSON have no comment syntax to find, and `#` in Markdown is a
  heading. Treating them as commentable would make every heading a violation.
- **Binary files are skipped** by looking for a NUL byte in the first 8 KiB
  rather than by maintaining a list of binary extensions.
- **A shebang is a toolchain directive**, not commentary.

Intrinsic classification is lifted out of the Go-specific path. The classes
that are language-neutral — generated markers, upstream links, `Deprecated:`,
and the user's `allowedAnnotations` patterns — apply everywhere. Toolchain
directives become a per-language prefix set (`# noqa`, `# type:`,
`# shellcheck`, `# yaml-language-server:`, `// eslint-`, `// @ts-`,
`#![allow(`, and so on). `public-api` stays Go-only, because it is the one
class that needs a parser to decide.

**This will be wrong sometimes, and that is accepted.** A heuristic that
catches a real comment in ten languages is worth more than a parser that
catches every comment in one. Mis-detections are issues to iterate on, not
reasons to withhold the gate. The escape hatches already exist: a repository
can exclude paths, add `allowedAnnotations` patterns, or acknowledge a specific
comment.

## Consequences

What becomes easier:

- The gate runs on the whole repository. A YAML-only or Python-only project
  gets the same enforcement Go projects have had.
- `koment comments convert` works wherever the detector does, which closes
  issue #89 finding 7 and removes the hand-migration instruction from managed
  `AGENTS.md` blocks.
- Adding a language is a table row, not a type.

What becomes harder:

- **Existing repositories will fail on first upgrade.** Any project with
  comments in non-Go files has been passing a gate that was not looking; after
  this it looks. This is a breaking change to what `koment comments check`
  reports and ships as `feat!:`. It is the correct break — the previous green
  was false.
- False positives are now possible where none existed, because there was no
  detection at all. The quote heuristic is a heuristic; a marker inside a
  multi-line string or a heredoc can still be misread.
- Python docstrings are out of scope. Python has no block comment, and a
  docstring is a string expression that conventionally documents public API.
  Treating them as comments would flag every documented function; treating them
  as `public-api` needs a parser. They are left alone, and this is recorded so
  the next reader knows it was decided rather than missed.

## Alternatives rejected

- **Write a real parser per language.** Most accurate, and what `goDetector`
  does. Rejected as the general strategy: ten parsers is ten dependencies
  against ADR 0010's bar, an enormous surface to maintain, and it still leaves
  every unlisted filetype silently unchecked — which is the actual bug.

- **Use tree-sitter with per-language grammars.** One dependency, real
  grammars, accurate comment nodes. Genuinely tempting and the strongest
  alternative. Rejected for now on ADR 0010 grounds: it adds a cgo or WASM
  runtime and a grammar per language to a tool whose distribution story is a
  single static binary, to improve accuracy on a gate whose failure mode is a
  false positive a user can exclude. Worth revisiting if the heuristic proves
  noisy in practice — the `Detector` interface means it would be an
  implementation, not a rewrite.

- **Generic marker scan with no per-extension table.** Much less code.
  Rejected because it cannot know that `#` starts a comment in YAML but a
  heading in Markdown, that `--` is a comment in Lua and SQL but an operator
  elsewhere, or that `/* */` spans lines. The table is where the accuracy
  lives; the fallback is only for what the table has not reached.

- **Keep skipping unmapped filetypes.** Safe, no false positives. Rejected
  because it preserves the silent no-op for anything not yet listed, and a
  quiet false pass is the one outcome this project treats as worse than a
  crash.

- **Warn on non-Go files instead of failing.** A softer migration. Rejected
  because a warning that never fails is a comment about a comment: agents and
  CI both ignore it, and the fast-feedback signal the tool exists to give would
  not fire.
