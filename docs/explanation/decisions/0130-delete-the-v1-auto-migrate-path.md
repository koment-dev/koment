# 0130 — Delete the v1 auto-migrate path and refuse `version: 1`

Date: 2026-08-09
Status: Accepted

## Context

[ADR 0119](0119-make-the-annotation-a-kubernetes-shaped-resource.md) reshaped the
annotation into a Kubernetes-shaped resource and shipped a loader that upgraded
a pre-v1alpha `version: 1` record in place on read.
[ADR 0121](0121-every-committed-koment-file-is-a-resource-with-a-pinned-schema.md)
did the same for `.koment/policy.yaml`. Both were explicitly temporary. ADR 0119
set the deadline in its own consequences:

> Auto-migrate must be deleted in 1.1.x.

The code says the same thing to anyone who opens it:

```go
// Deprecated: delete legacyRecord and upgradeLegacy in the release after
// 1.0.0. By then every repository that a 1.0.x binary has read is already in
// the current shape, and a record still carrying `version: 1` should be
// refused rather than silently upgraded.
```

The deadline passed. koment is at 2.2.0 and both shims are still in the tree,
so the project is running past a cutoff it wrote down and published. That is
the precise failure mode koment exists to prevent — a recorded decision that
the code quietly stopped honouring — and leaving it costs more than the shims
save.

The shims also impose a structural cost that is easy to miss. Because a v1
record is rewritten during `Store.Load`, **reading a repository mutates it**.
That single fact forced a special case: a read-only mount would fail its write,
so `persistUpgrade` swallows the error to a `slog.Warn` and the ordering in
`decodeAnnotation` had to be pinned by an invariant annotation so that a future
editor did not reverse it. An entire class of care exists only to keep a
temporary upgrade from breaking read-only deployments.

## Decision

Delete the v1 auto-migrate path from both loaders. `legacyRecord`,
`legacyGit`, `upgradeLegacy`, `upgradeLegacyGit`, `legacyPolicy`,
`legacyCommentsPolicy` and `persistUpgrade` go, along with the `upgraded bool`
that `decodeAnnotation` and the policy `decode` threaded back to their callers.

koment keeps **recognising** `version: 1` and refuses it with an error that
names the way out:

```
incompatible <name>: this record is in the pre-v1alpha `version: 1` shape,
which koment no longer reads (ADR 0130). Read this repository once with
koment 2.x, which rewrites every record in the v1alpha shape, then retry.
```

Recognition is deliberately retained. `LegacyRecordVersion` and
`LegacyVersion` survive as constants whose only job is to make this error
reachable. Dropping detection along with migration would send a v1 record down
the generic "no apiVersion" path, which tells the reader their file is
malformed when it is merely old.

`Store.Load` becomes side-effect-free: reading a repository no longer writes to
it under any circumstance.

This ships as `feat!:` and takes koment to 3.0.0. Per the back-compatibility
rule in `AGENTS.md`, a change is breaking unless the binary performs a
migration or an ADR names the version the old shape was cut off at. The binary
no longer migrates, so this ADR is that naming: **`version: 1` records and
policies are readable by koment 0.x through 2.x, and by no later version.**

## Consequences

What becomes easier:

- Reading is a read. The read-only-mount special case, the swallowed write
  error and the ordering invariant that protected it all disappear rather than
  being maintained.
- Two decode paths collapse to one. `decodeShape` and the policy `decode`
  return a value and an error instead of a value, a flag and an error, and no
  caller has to decide what to do about the flag.
- The published record shape is what the code reads, with no second shape kept
  alive behind it. A reader of `internal/store` sees one generation.

What becomes harder:

- A repository last touched by a 0.x binary needs one run of koment 2.x before
  a 3.x binary will read it. This is a real migration step for anyone who
  skipped the 1.x and 2.x lines entirely. The error names the step, and the
  2.x binaries stay downloadable from the GitHub releases; they are not
  withdrawn.
- The v1 shape is no longer exercised by any test, so the shape itself is now
  documented only by this ADR and by 0119. That is the intended end state — the
  shape is history — but it does mean the record of what v1 looked like lives
  in decisions rather than in code.

Two annotations were bound to machinery this change deletes and were retired
with it: the `why` on `persistUpgrade` (read-only mounts tolerate a failed
rewrite) and the `invariant` on `decodeAnnotation`'s ordering (validate before
touching the filesystem). Both described a hazard that no longer exists, and
keeping them would have left the repository asserting a constraint about code
it no longer contains. Retiring them required deleting the record files by
hand, because koment has no `forget` command — the gap tracked in issue #89.
The rationale is preserved here instead, which is where a decision about a
removed structure belongs.

## Alternatives rejected

- **Keep auto-migrate indefinitely.** Cheapest today and breaks nobody.
  Rejected because it makes the published deadline a lie, and because the cost
  is not the code — it is that reading keeps mutating the repository, which is
  a surprising property to carry forever for the benefit of a shape no
  supported binary has written since 1.0.0.

- **Delete detection along with migration**, letting a v1 record fall through
  to the existing "no apiVersion" error. One less constant and one less branch.
  Rejected because it degrades a good error into a misleading one: the user is
  told their record is not a koment record, when in fact koment wrote it. One
  struct field is a small price for an error that names the fix.

- **Ship a `koment migrate records` one-shot command as the replacement.**
  Gives users a forward path without an old binary. Rejected on the same
  grounds ADR 0119 rejected it for the original migration — it is a step that
  has to be remembered — and more decisively because it means writing new code
  to support a shape being removed. The 2.x binaries already do this and
  remain published.

- **Do it in a minor release.** Rejected outright. A 3.x binary refuses a
  repository a 2.x binary reads; that is a break by any definition, and
  `AGENTS.md` requires it be labelled as one.

- **Keep the policy shim but drop the record shim** (or the reverse). The two
  were introduced by different ADRs, so splitting them is defensible.
  Rejected because they share one deadline and one rationale; retiring them
  separately would mean two breaking releases where one suffices, and would
  leave a reader wondering why one generation of file is still upgraded and
  the other refused.
