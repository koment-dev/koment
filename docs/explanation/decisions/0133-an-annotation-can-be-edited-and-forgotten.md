# 0133 — An annotation can be edited and forgotten

Date: 2026-08-09
Status: Accepted

## Context

[Issue #89](https://github.com/koment-dev/koment/issues/89) reported the
annotation lifecycle as write-once. Two of its findings are the same shape:
once a record exists, koment offers no way to change it or retire it.

- **Nothing removes an annotation.** The CLI and the MCP surface both stop at
  `reanchor`. Retiring a record meant `rm`-ing the YAML by hand, or reanchoring
  it somewhere irrelevant so the rationale quietly became wrong. This is not
  hypothetical: removing the v1 auto-migrate path ([ADR 0130](0130-delete-the-v1-auto-migrate-path.md))
  orphaned two annotations describing deleted machinery, and the only way to
  retire them was `git rm`.
- **A title is frozen at creation.** `koment add --title` rejects anything over
  72 characters, `koment comments convert` has no `--title` flag at all, and
  nothing can set one afterwards. A record written without a title keeps the
  first sentence of its body as its headline forever unless someone hand-edits
  the YAML. Every one of the nineteen conversions done under ADR 0132 hit this.

## Decision

Two commands, both narrow.

**`koment forget <id>` deletes the record file.** No tombstone, no `status:
removed`, no `--reason`. Git already holds every fact a tombstone would carry:
who removed it, when, in which commit, alongside what other change, and with
the full prior content recoverable. koment's founding decision is that
[git is the record](0100-one-git-record-per-annotation.md); duplicating git's
audit trail inside the data would be the tool distrusting its own premise. The
command prints the headline it removed and the exact `git checkout --` needed
to bring it back, so the escape hatch is on screen at the moment of deletion.

**`koment edit <id> [--title <text>] [--body <text|->]`** rewrites the two
prose fields in place. Identity, authorship, creation time, kind and anchor are
not editable here — `reanchor` owns the anchor, and the rest are provenance.
Editing with neither flag is a usage error rather than a silent no-op.

The 72-character title limit **stays a hard rejection**. Relaxing it was the
preferred option when this work started, but `maxLength: 72` is written into
`schema/v1alpha/annotation.schema.json`, which [ADR 0121](0121-every-committed-koment-file-is-a-resource-with-a-pinned-schema.md)
froze at 1.0.0. Warning in Go while the published schema still refuses the
record would make koment write files that fail its own schema — including the
editor validation wired up through `.vscode/settings.json`. Relaxing it
properly means a `v1beta` directory and a generation migration, which is a
larger decision than a usability fix should make. What changes instead is that
the rejection is no longer permanent: a title can now be shortened at write
time and improved later with `edit`. The error message says so.

## Consequences

What becomes easier:

- Retiring an annotation is a command with an audit trail rather than a manual
  `rm` that leaves no trace of intent.
- A title can be polished after the body is written, which is the order people
  actually write in.
- ADR 0132's conversions, and any future bulk conversion, can title records
  afterwards instead of accepting whatever the first sentence happens to be.

What becomes harder:

- `forget` is destructive and takes no confirmation. In a repository with
  uncommitted work the record is gone from disk, and only committed records are
  recoverable from git. The printed `git checkout --` is the mitigation, and it
  is honest about that: it restores what git has, not what was never committed.
- `edit` can rewrite a body to say something the anchored code does not support.
  Nothing validates that rationale still matches the code, because nothing can.
  `koment check` still verifies the anchor resolves, which is the only claim the
  tool ever made.
- The title limit remains a first-write rejection, so the reported friction is
  reduced rather than removed. The ceiling moves when the API version does.

## Alternatives rejected

- **Tombstone on forget: keep the id, set `status: removed`, require
  `--reason`.** What the issue asked for, and the usual "supersede, don't
  erase" pattern. Rejected because git already stores every field a tombstone
  would, with better fidelity, and because tombstones accumulate: every
  retired annotation would stay in `list`, `search` and the published site
  forever, and every reader would have to learn to filter them. A record that
  no longer describes anything is not history worth serving — the commit that
  removed it is.

- **Tombstone plus a `koment prune` to collect them later.** Keeps both
  options open. Rejected as the worst of both: it builds the accumulation
  problem and then builds a second command to undo it, and it forces every
  reader of the store to handle two kinds of record.

- **Make `forget` require `--yes` or an interactive confirmation.** Safer for
  a destructive command. Rejected because it is an id-addressed operation —
  the caller has already had to find the exact ULID, which is not something
  done by accident — and because an interactive prompt is hostile to the
  agent and hook callers that are koment's primary users.

- **Let `edit` change `kind` and the anchor too.** One command for every
  mutation. Rejected: `reanchor` already owns the anchor and does it with
  validation `edit` would have to duplicate, and a kind change is a
  reclassification that deserves its own thought rather than riding along with
  a typo fix.

- **Soft-warn over-length titles now and fix the schema later.** The original
  plan. Rejected once `maxLength: 72` turned out to be in the frozen v1alpha
  schema: koment would emit records that fail its own published validation,
  which is a worse defect than the one being fixed.
