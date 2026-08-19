---
description: Point a stale annotation back at the code it belongs to
argument-hint: "<annotation id, file, or nothing to fix all reported drift>"
---

Repair the anchor for:

$ARGUMENTS

Run `koment check` first to see what is actually broken. The three failing
statuses mean different things:

| status | what happened | what fixes it |
|---|---|---|
| `drifted` | the excerpt no longer matches | a new excerpt from the current code |
| `ambiguous` | the excerpt matches several places | a **longer** excerpt |
| `orphaned` | the file is gone | a decision, not a reanchor |

Then call `koment_reanchor` with the new excerpt, matched byte for byte
including indentation.

For `ambiguous`, extend the **excerpt** with adjacent lines until it is unique.
Widening `before` or `after` will not work — they are context hints capped at
three lines and they do not disambiguate a repeated excerpt.

The id survives a reanchor. That is the point: the record keeps its history and
its authorship, and only the anchor moves.

**Read the body before you repair the anchor.** If the code changed in a way
that makes the rationale wrong rather than merely relocated, reanchoring it
produces a confident, well-anchored lie. In that case say so and use
`/koment:forget` or rewrite the body with `koment edit` instead.

For `orphaned`, ask whether the reasoning moved with the code or died with the
file. Do not silently reanchor it to whatever looks closest.
