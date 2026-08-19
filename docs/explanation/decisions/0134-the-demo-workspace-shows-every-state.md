# 0134 — The demo workspace shows every state koment can produce

Date: 2026-08-09
Status: Accepted

## Context

`workspace/` is the maintained fixture repository published alongside koment's
own annotations at `why.koment.dev`. For most visitors it is the only running
koment they will ever see before deciding whether to install it.

Today it shows a fraction of the product:

```
kinds:     2 why, 1 gotcha, 1 invariant     (anti-pattern missing)
statuses:  4 ok                             (ambiguous, drifted, orphaned missing)
```

Three of four kinds, one of four statuses. A visitor sees a list of green
badges and learns that koment stores rationale — which is the least
interesting half of the claim. They never see the thing the tool actually
sells: an annotation going **drifted** the moment the code beneath it changes.
The demo is at its least convincing on precisely the point that distinguishes
koment from a wiki.

This is not a content gap that can be fixed once. `workspace/` is maintained
alongside the code, and nothing stops the next change from quietly removing the
only `anti-pattern` or repairing the only drifted anchor. The published site
would silently narrow again, and nobody would notice until a visitor did.

## Decision

**The demo workspace must exercise every annotation kind and every anchor
status koment can produce, and CI asserts it.**

Concretely, `workspace/` carries at least one annotation of each kind — `why`,
`gotcha`, `invariant`, `anti-pattern` — and at least one anchor resolving to
each status — `ok`, `ambiguous`, `drifted`, `orphaned`.

The failing statuses are deliberate fixtures, not neglect. That changes what
`koment check` means in that directory: it is *expected* to exit non-zero, and
the pages workflow can no longer gate on its exit code. Instead CI asserts the
exact tally, so the demo's shape is pinned:

```
koment check   # in workspace/, exits 1
→ "N annotations across M files: A ok, B ambiguous, C drifted, D orphaned"
```

An assertion on that line fails when someone repairs a fixture, deletes the
last annotation of a kind, or lets a real regression change the mix. The demo
cannot narrow without a red build.

koment's own annotations at the repository root keep the ordinary contract:
`koment check` there must exit 0, and a drifted annotation is a defect. The two
directories are different things and this ADR only changes the fixture.

## Consequences

What becomes easier:

- The published site demonstrates drift, ambiguity and orphaning to someone who
  has never run the tool. The screenshot that sells koment is one it can now
  actually produce.
- Every status badge, filter and empty state in the UI has a live example, so
  UI regressions in the rarely-populated states are visible rather than
  theoretical.
- Adding a status or kind to the model forces a demo fixture for it, because
  the tally assertion will not match until one exists.

What becomes harder:

- `koment check` in `workspace/` no longer means "healthy". A contributor who
  runs it there and sees failures has to know they are intentional. The
  workspace README says so, and the CI step is named for what it asserts.
- The tally assertion is a hard-coded expectation that has to be updated
  deliberately whenever the fixture set changes. That is the point — it is a
  gate, not a convenience — but it will occasionally be the thing that fails a
  build for a legitimate change.
- An orphaned fixture needs an annotation pointing at a file that does not
  exist, which looks like a mistake to anyone reading `.koment/` directly. It
  is labelled in its own body.

## Alternatives rejected

- **Leave the demo all-green and describe drift in prose.** No fixture
  maintenance, no CI change. Rejected because the site's job is to show the
  behaviour, and a tool whose central claim is "stale rationale is detected"
  cannot credibly demonstrate that with a screenshot of green badges and a
  paragraph.

- **Keep `koment check` gating and generate the broken states at render time**,
  by mutating fixture files inside the pages workflow before rendering.
  Rejected: the published site would then show states that do not exist in the
  committed repository, so a visitor who cloned it would find something
  different from what they were shown. The demo has to be a real repository or
  it is a mock-up.

- **A second fixture repository holding only the failing states**, left out of
  the gating check. Keeps `workspace/` green. Rejected as more machinery for a
  worse result: two demo repositories to maintain, and the interesting
  comparison — healthy and stale annotations side by side in one tree, which is
  what a real repository looks like — is exactly what gets lost by separating
  them.

- **Assert only that each kind and status appears at least once, without
  pinning counts.** More tolerant of fixture churn. Rejected because the
  looser assertion cannot catch the common regression: a fixture that was
  supposed to be `drifted` silently resolving to `ok` while some other
  annotation drifts by accident. The counts are what make the demo's shape
  intentional.
