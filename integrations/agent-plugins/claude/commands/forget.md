---
description: Delete an annotation, with the record of who removed it
argument-hint: "<annotation id, or a description of the annotation>"
---

Delete the annotation identified by:

$ARGUMENTS

`koment forget` is destructive. Treat it that way:

1. **Find it first.** If the argument is an id, read it. If it is a description,
   use `koment_search` or `koment_get` and identify the candidates. If several
   are plausible, list them and ask — do not guess.
2. **Show the full body and get explicit confirmation before deleting.** Never
   delete more than one annotation from a single confirmation.
3. Only then run `koment forget <id>`.

**Deleting to make a gate pass is prohibited.** If `koment check` reports the
annotation as `drifted`, `ambiguous` or `orphaned`, that is a broken anchor, not
a worthless record — use `/koment:reanchor`. The whole value of this repository
is that rationale outlives the code it described; an agent that deletes
inconvenient history has inverted the tool.

**Prefer correcting over deleting.** `koment edit` rewrites a headline or body
in place and keeps the id, the author and the creation date. Reach for it when
the rationale is merely wrong or outdated rather than unwanted, and offer that
path before this one.

Git keeps who removed it and when, so the deletion is auditable. Say that
plainly — it is not a way to make something disappear.
