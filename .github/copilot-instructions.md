<!-- koment:managed-start -->
# KOMENT PROCEDURE — MANDATORY

This file is a managed contract. The repository enforces it with
client hooks and a required CI status. You MUST follow every rule.
Partial compliance is a bug.

## Before any edit or write to an existing file

- You MUST invoke `koment_get`("<repository-relative path>") via MCP
  and treat the returned annotations as the authoritative context for
  the file. An annotation whose status is `ambiguous`, `drifted`
  or `orphaned` is history, not current fact; surface it to the human
  before continuing.
- You MUST invoke `koment_search`("<topic>") before any non-obvious
  structural decision that another file may already explain.

## Adding or changing a comment is FORBIDDEN

- You MUST NOT write an ordinary inline comment in source. The
  repository classifies every comment group and rejects ordinary ones
  on the protected branch. The only exceptions are the intrinsic
  classes enabled in `.koment/policy.yaml` (toolchain directives,
  generated markers, upstream links, `// Deprecated:`, godoc on
  exported identifiers) and any additional pattern declared under
  `spec.comments.allowedAnnotations`.
- Before keeping a comment, you MUST attempt in order: rename the
  thing, extract a function whose name is the comment you were about
  to write, introduce a named type or constant, restructure so the
  invariant is obvious from control flow. If the rationale still needs
  saying, call `koment_add` bound to the code with `--excerpt`
  and record yourself honestly as an agent.
- If a comment already exists in source, you MUST call
  `koment_convert_comment` first to record it as an annotation,
  then remove the comment from source.
- Keeping an inline comment requires `koment_acknowledge_comment`
  with `acknowledge_inline_comment: true` and a human-readable body.
  The acknowledgement is auditable.

## Anchoring an annotation

- `excerpt` is the anchor. It must match the file byte for byte,
  including indentation, and it has NO line limit. If an excerpt is
  rejected as matching several places, extend the excerpt itself with
  adjacent lines until it is unique.
- `before` and `after` are context hints only, capped at three
  lines each. They do NOT disambiguate a repeated excerpt, so widening
  them is never the fix for an ambiguous anchor.
- If an excerpt is reported missing but you believe it is present, the
  difference is whitespace: indentation, a trailing space, or CRLF
  endings. koment says so when it can detect it.

## Before you stop

- You MUST run `koment check`, `koment comments check` and
  `koment agents check`. You MUST NOT report success while any
  fails.

A back-compatibility claim needs evidence: a migration path the binary performs, or an ADR naming the version the old shape was cut off at. Without either, the change is breaking and its commit subject says so with `feat!:`.
<!-- koment:managed-end -->
