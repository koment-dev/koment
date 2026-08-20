# 0146 — Gate distribution on visible release assets

Date: 2026-08-19
Status: Superseded by 0147

## Context

The 3.1.0 release created its tag and GitHub Release, then published the
container image, Helm OCI artifact and MCP Registry metadata. Both the binary
and chart jobs built and signed their files, but their `gh release upload`
commands returned `release not found`. The GitHub Release was visible after the
workflow stopped, but its asset list was empty. Jobs downstream of `binaries`
were skipped.

Re-running the failed chart job would package the chart from a new checkout and
push the already-public `3.1.0` OCI tag again. Even if its logical files were
unchanged, the package bytes and digest are not guaranteed to be identical.
That would replace an immutable registry version, which the release procedure
forbids. The safe recovery is a workflow fix followed by 3.1.1.

ADR 0109 requires canonical artifacts before distribution metadata. The
workflow already applied that order to plugins, editors and package-manager
metadata, but the chart and MCP Registry jobs depended only on the image. That
allowed both to publish while the canonical GitHub archives were absent.

## Decision

Every GitHub Release asset upload goes through
`scripts/upload-release-assets.sh`. The helper names the repository explicitly,
waits for the release to become visible through the publishing job's token, and
fails after six bounded attempts. Only after visibility is established does it
perform one upload with replacement enabled for a partial upload from that same
job attempt.

The chart and MCP Registry jobs depend on both canonical publication jobs:
`binaries` and `image`. A binary-asset failure therefore prevents registry
metadata and a Helm version from becoming the only successful downstream
channels. Plugins already had this dependency and remain unchanged.

The workflow uses `github.token` explicitly for GitHub CLI operations and passes
`--repo` to every release upload or download. Repository inference and a token
alias are no longer implicit parts of the release contract.

Version 3.1.0 remains as published. Its tag, release, image, chart and registry
metadata are not deleted, moved, overwritten or retried. Version 3.1.1 is the
first release to use the corrected publication graph.

## Consequences

- A delayed GitHub Release becomes an explicit bounded wait instead of a late
  `release not found` after artifacts have already been built and signed.
- A persistent visibility or permission failure stops before the immutable Helm
  and MCP distribution surfaces publish.
- The image and binary publishers can still run in parallel because both are
  canonical surfaces under ADR 0109.
- Release publication can take up to 150 additional seconds while GitHub makes
  a new release visible.
- The helper has a deterministic shell test with fake `gh` and `sleep`
  executables, so retry count, delays, repository targeting and upload arguments
  are enforced without contacting GitHub.

## Alternatives rejected

- **Re-run the failed 3.1.0 jobs.** This could replace the already-published
  chart digest and would contradict the mandatory patch-release recovery rule.
- **Upload the missing assets by hand.** That bypasses the signed workflow and
  would leave no reproducible publication record.
- **Retry `gh release upload` after any error.** Permission failures and invalid
  arguments are not transient. The helper retries only the read-only visibility
  probe, then performs the write once.
- **Make the chart and MCP jobs wait only for a timer.** Time is not evidence
  that the canonical assets exist. Their dependency on `binaries` is the
  evidence.
- **Serialize every publication job.** It reduces partial concurrency but makes
  independent canonical image and binary builds wait on each other without
  eliminating failures in later external registries.
