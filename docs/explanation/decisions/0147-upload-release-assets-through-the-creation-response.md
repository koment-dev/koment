# 0147 — Upload release assets through the creation response

Date: 2026-08-20
Status: Accepted

## Context

ADR 0146 made every asset publisher wait until `gh release view <tag>` could
rediscover a release created by the preceding release-please job. Version
3.1.1 passed that gate, but version 3.1.2 did not: release-please published the
release at 08:19:19 UTC, while the binary job received `release not found` from
all six probes between 08:21:40 and 08:24:12 UTC. The release was public and
had a stable numeric identifier throughout that interval.

The retry treated tag lookup as evidence of release creation even though
release-please already returns the create-release response. That response
contains `upload_url`, the release-specific hypermedia endpoint GitHub defines
for raw asset uploads. Rediscovering the same object by tag in another job
added a consistency boundary without adding information.

The dependency graph introduced by ADR 0146 still prevented the partial
3.1.2 release from publishing Helm, MCP, editor, plugin or Homebrew versions.
Only its canonical container image escaped before the binary failure. The tag,
release and image therefore remain published, and recovery must use 3.1.3.

## Decision

Pass release-please's `upload_url` output to every asset-publishing job. The
upload helper validates that the URL targets the current repository, removes
the URI-template suffix and posts each asset as raw binary data to that exact
endpoint. Asset names are restricted to the portable characters already used
by every release artifact.

Downstream jobs consume the exact release version from this workflow through
public asset URLs. Homebrew and editor packaging no longer ask GitHub CLI to
rediscover the release by tag, and the release verification matrix passes the
exact version to the setup action instead of resolving `latest`. Those bounded
downloads retry transient HTTP failures but never choose another release.

The helper performs no tag lookup, visibility retry or replacement upload. A
create-release response is the authoritative identifier for the release it
created. A failed or partial write stops the workflow; published versions are
never repaired by replacing an asset or rerunning an immutable distribution
job. Recovery remains a code fix followed by the next patch release.

ADR 0146's canonical dependency graph remains in force. This decision
supersedes only its tag-visibility gate and replacement-upload mechanism.

## Consequences

- Asset upload no longer depends on cross-job tag lookup becoming consistent.
- The workflow uses the release identifier returned by the write it is
  continuing, rather than searching for the object it just created.
- Downstream packaging and verification cannot drift to a different `latest`
  release or repeat the authenticated tag lookup that failed publication.
- A malformed or cross-repository upload URL fails before any network write.
- Duplicate names fail instead of silently replacing bytes in a published
  release.
- Each file is a separate HTTP request, so a later failure can leave an honest
  partial release. The mandatory next-patch rule handles that state without
  rewriting history.
- The deterministic helper test verifies repository binding, URI-template
  removal and one raw upload request per asset.

## Alternatives rejected

- **Increase the visibility timeout.** Version 3.1.2 proved that elapsed time
  is not evidence, and the create response already contains the needed
  identifier.
- **Keep `gh release upload --clobber` after a successful lookup.** It preserves
  the unnecessary lookup and makes replacement of published bytes appear to be
  a normal recovery operation.
- **Upload assets in the release-please job.** Binary, chart, plugin and editor
  builds require different toolchains and permissions. Combining them would
  erase the existing least-privilege and parallel-build boundaries.
- **Use a third-party release action.** GitHub documents the asset endpoint
  directly, the repository already has curl, and a dependency would add more
  code and authority than this handoff requires.
- **Repair 3.1.2 by hand or rerun its jobs.** Its container image is already
  public. Either path would create an out-of-band or mutable release record.
