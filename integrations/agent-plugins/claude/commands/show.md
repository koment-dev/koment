---
description: Show the recorded rationale for a file
argument-hint: "<repository-relative path>"
---

Read the annotations on this file before touching it:

$ARGUMENTS

Call `koment_get` with the repository-relative path. Fall back to
`koment show <file>` only if MCP is unavailable.

When you report back:

- Lead with anything whose status is **not** `ok`. An `ambiguous`, `drifted` or
  `orphaned` annotation describes code that cannot be resolved reliably — it is
  **history, not current fact**. Say so explicitly rather than presenting it
  alongside resolved rationale as if both were true.
- Summarize the bodies; do not dump the YAML.
- Name the author kind. An annotation written by an agent carries different
  weight from one a person wrote, and the record says which.
- If the file has no annotations, say that plainly. Empty means nothing was
  recorded — do not infer rationale from the code and present it as recorded.

If the path is wrong, `koment_get` says so rather than returning nothing. A
result of zero annotations for a path that exists is a real answer.
