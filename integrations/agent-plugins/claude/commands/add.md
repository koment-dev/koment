---
description: Record why the code is the way it is, bound to the code
argument-hint: "<file> — and what needs explaining>"
---

Record rationale for:

$ARGUMENTS

**First, try to make the annotation unnecessary.** In this order:

1. Rename the thing. Most explanations exist because a name is bad.
2. Extract a function whose name is the sentence you were about to write.
3. Introduce a named type or constant instead of a bare value.
4. Restructure so the invariant is obvious from control flow.

Only when all four fail has the rationale earned a record. Say which ones you
tried and why they did not dissolve it.

Then call `koment_add`:

- `file` — repository-relative path.
- `excerpt` — the anchor, matched **byte for byte including indentation**. It
  has no line limit. If it is rejected as matching several places, extend the
  excerpt itself with adjacent lines; widening `before`/`after` will not help,
  because they are context hints and do not disambiguate.
- `kind` — `why`, `gotcha`, `invariant` or `anti-pattern`.
- `title` — one line, at most 72 characters.
- `body` — explains **why**, never **what**. If it narrates the code, delete it
  and go back to step 1.

Anchor to the **code**, never to a comment you are about to remove — the
annotation orphans the moment the comment goes.

If an excerpt is reported missing but you can see it in the file, the difference
is whitespace: indentation, a trailing space, or CRLF endings.

Authorship is recorded automatically as an agent. Do not work around that.

If the rationale is project-wide, or about a structure rather than a place, it
belongs in an ADR under `docs/explanation/decisions/` instead. Say so rather than forcing it
onto a line.
