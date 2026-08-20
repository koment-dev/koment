# 0148 — Reject references to retired repository paths

Date: 2026-08-20
Status: Accepted

## Context

ADR 0143 made the repository tree a closed contract and required every
repository-controlled reference to move with a path migration. Its executable
check rejects files at retired locations, but it did not inspect file contents.
The documentation migration therefore removed the old release guide while four
older ADRs continued to name its retired path. Every current entry point used
the new guide, so CI stayed green while authoritative explanation pages made a
false path claim.

A repository path written in prose, a workflow or a script is part of the same
ownership contract as the file itself. Review and search found this instance,
but neither prevents the next migration from leaving another reference behind.
The project treats wrong documentation and workspace clutter as technical debt,
so a completed migration needs a durable content gate rather than a remembered
cleanup step.

Annotation records are different. Their Git provenance records the path at the
time an annotation was written. Rewriting that historical field after a move
would replace evidence with the current layout and make the record less true.

## Decision

The `internal/projectlayout` checker maintains explicit retired-to-current
reference migrations. It scans every tracked and non-ignored repository file
and fails when one contains a retired path, naming the required replacement.
Read failures stop the check rather than silently omitting a file.

Files under `.koment/` are excluded from the content scan because their stored
Git paths are historical provenance. Their active target paths and anchors
remain covered by `koment check`; the exception does not permit live source or
documentation to retain an obsolete reference.

The generated repository layout reference states that retired references are
part of the contract. A regression test proves both rejection and the narrow
historical-provenance exception. Adding a compatibility mapping for a future
migration is part of that migration's implementation and documentation work.

## Consequences

- A moved file cannot leave a tracked prose, workflow or script reference
  behind while the layout gate remains green.
- The failure names both the obsolete text and its replacement, so the repair
  is deterministic.
- Retired strings must not appear contiguously in the checker's own source; the
  migration table composes them from named path fragments so the gate does not
  exempt itself.
- Scanning repository files adds bounded local I/O to a gate that already asks
  Git for the same finite path set.
- Historical annotation provenance remains immutable and auditable.

## Alternatives rejected

- **Rely on the migration checklist and review.** ADR 0143 already required
  atomic reference updates, and the stale release-guide references still
  survived. Repeating the same human control is not a fix.
- **Add a third-party Markdown link checker.** It would catch broken Markdown
  links but not path claims in code spans, workflow inputs or scripts. It also
  introduces a dependency for a check the standard library can perform.
- **Keep redirect files at retired locations.** Compatibility stubs recreate
  two plausible homes and violate the closed tree that the migration was meant
  to establish.
- **Reject the retired text inside `.koment/` too.** That would require changing
  immutable Git provenance to make a current-layout check pass. The result
  would be cleaner-looking but historically false.
- **Exclude all ADRs as historical documents.** ADRs also describe active
  constraints and point readers to current implementation surfaces. Excluding
  them is what allowed these false references to survive.
