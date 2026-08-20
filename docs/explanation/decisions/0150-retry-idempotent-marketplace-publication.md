# 0150 — Retry idempotent marketplace publication

Date: 2026-08-20
Status: Accepted

## Context

The 3.1.4 editor job attached all seven signed VSIX files and began publishing
them sequentially. VS Marketplace accepted the universal package and the
darwin-arm64 package, then the pinned `vsce` client timed out after three
minutes while publishing darwin-x64. The service may accept a package before a
client observes the response, so the timeout could not say whether that third
immutable platform version existed. Open VSX never ran because the step failed.

Neither hand-publishing the remaining packages nor rerunning the old job is
safe. Marketplace versions are permanent, while a job rerun would also try to
attach duplicate release assets before it reached the marketplace steps. The
written release rule therefore leaves 3.1.4 partial and requires recovery in
3.1.5.

The pinned publisher clients already expose the primitive needed inside one
job. `@vscode/vsce` 3.9.2 and `ovsx` 1.1.0 both document
`--skip-duplicate`: a retry succeeds silently when the same version and target
already reached the marketplace.

## Decision

Publish each VSIX with the publisher's `--skip-duplicate` flag. Wrap each
individual publish command in three attempts separated by ten seconds. A
timeout after the service accepts a package therefore retries the same bytes;
the duplicate is accepted as success and the loop advances to the next target.

Keep signing, GitHub attachment and the two marketplaces in their existing
order. This retry is local to one marketplace command in one release run. It
does not make the whole release workflow replayable and does not permit a
failed historical run to be rerun after another registry version is public.
If all three attempts fail, the job fails and recovery remains the next patch.

## Consequences

- A transient response timeout no longer strands the remaining platform
  packages after the marketplace already accepted earlier ones.
- Duplicate tolerance is explicit and restricted to identical immutable
  marketplace versions; GitHub release assets retain their reject-duplicate
  behavior from ADR 0147.
- Permanent authentication, validation and publisher errors take up to two
  extra attempts and twenty seconds before failing.
- 3.1.4 remains an honest partial release. 3.1.5 is the first version protected
  by this retry boundary.

## Alternatives rejected

**Rerun the failed 3.1.4 editor job.** The run predates this change and would
fail while attaching duplicate GitHub assets. Even if it reached the publisher,
it would treat already accepted marketplace versions as errors.

**Publish only the missing packages by hand.** This bypasses the signed release
workflow, and a timeout cannot reliably identify which package the service
accepted. The release procedure explicitly forbids this recovery path.

**Publish every platform concurrently.** This reduces elapsed time but makes a
partial failure harder to order and does not make an ambiguous timeout safe.

**Cut another patch after every transient failure.** A next patch is the safe
fallback, but using it for a response timeout the publisher can recognize as a
duplicate turns a recoverable network condition into permanent registry
clutter.
