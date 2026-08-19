---
description: Turn an existing source comment into an annotation and remove it
argument-hint: "<file> [line or the comment text]"
---

Migrate the comment in:

$ARGUMENTS

Order matters. **Record it before you delete it**, or the rationale is gone with
no way to recover it:

1. Read the file's existing annotations first with `koment_get`.
2. Call `koment_convert_comment` for the comment. It writes the annotation
   anchored to the code the comment described.
3. Only then remove the comment from source.
4. Re-run `koment comments check`.

Anchor to the **code**, not to the comment text. An annotation anchored to a
comment you are about to delete orphans immediately.

Detection reads the file's syntax, so this works in every filetype koment scans
— Go is parsed with `go/ast`, everything else with the line and block markers
its syntax declares. Only prose and data formats (`.md`, `.json`, `.svg`, …)
report `no comment detector for <file>`; there is no comment syntax there to
find. See the [language reference](https://github.com/koment-dev/koment/blob/main/docs/reference/languages.md).

**Not every comment should become an annotation.** Leave these where they are:

- toolchain directives (`//go:generate`, `# shellcheck`, `# type:`)
- generated-file markers
- `Deprecated:` markers
- links to an upstream issue or spec that explains external behaviour
- godoc on an exported identifier — that is API documentation, not commentary
- commented-out code — converting it turns disabled config into rationale

If the comment is genuinely worth keeping in source, use
`koment_acknowledge_comment` with `acknowledge_inline_comment: true` and a body
saying why all four code-clarity alternatives failed. That acknowledgement is
auditable, so make it honest.
