# Maintained workspace

This session package is part of koment's verification surface, and it is also
the demo published at [why.koment.dev](https://why.koment.dev/). It is small
enough to read in one sitting and useful enough to exercise real invariants.

**`koment check` exits non-zero here, on purpose.** Per
[ADR 0134](../docs/decisions/0134-the-demo-workspace-shows-every-state.md) this
workspace carries one annotation of every kind and one anchor in every state —
including `ambiguous`, `drifted` and `orphaned` — so that a visitor can see
what stale rationale looks like rather than only reading that koment detects
it. Each deliberately broken fixture says so in its own body.

CI asserts the exact tally, so the demo cannot quietly narrow. Repairing a
fixture is a red build, not a fix.

Run it from the repository root:

```sh
go test ./workspace/...
cd workspace && go run ../cmd/koment check   # expected: 5 ok, 1 ambiguous, 1 drifted, 1 orphaned
```

The published koment site exposes this workspace through the normal repository
switcher. It uses the same source, records, resolver and rendering path as any
other repository.
