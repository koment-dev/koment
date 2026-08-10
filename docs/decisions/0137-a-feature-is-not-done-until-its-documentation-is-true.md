# 0137 — A feature is not done until its documentation is true

Date: 2026-08-10
Status: Accepted

## Context

koment exists because prose that describes code stops being true and nothing
says so. The project has been running that failure on itself.

Every one of these was found in a single working session, and none of them was
caught by a test, a linter or a reviewer:

- `docs/publishing.md` — the copy-paste workflow the README points newcomers at
  — pinned `koment-dev/koment@v0.2.0`, **three major versions behind**, plus
  four superseded action pins. Anyone who copied it installed a year-old koment.
- `workspace/README.md` stated the demo was "kept green rather than carrying
  intentionally stale rationale", one commit after ADR 0134 made it carry
  exactly that, deliberately.
- Two MCP tool descriptions told agents that comment conversion handled a "Go
  comment group", after ADR 0132 made it handle every language.
- `docs/agents/hermes.md` described Hermes as an MCP client only, after the
  plugin that enforces policy in Hermes had shipped.
- The README led its install section with mise and reached a demonstration of
  the tool failing only at line 109.

The common shape is not carelessness. It is that **changing behaviour and
changing the prose that describes it were separable acts**, so the second one
was reliably deferred and then forgotten. A test suite gates the code. Nothing
gated the sentences.

The same argument the project makes about comments applies here exactly: a
description that has silently stopped being true is worse than no description,
because a reader cannot tell the difference.

## Decision

**Documentation is part of the change, not follow-up work.** Four rules, added
to `AGENTS.md` so they bind humans and agents alike.

1. **Adding a feature includes the prose that describes it, in the same commit.**
   If a user-visible capability has no line in `docs/` or the README, it is not
   finished. A commit that adds a command, a flag, an environment variable or a
   published artifact and touches no documentation is incomplete by definition.

2. **Removing a feature includes removing every description of it.** Deleting
   code and leaving the paragraph is how a manual starts lying. Search for the
   name before declaring the removal done.

3. **A version, command or flag written in an example is a claim, and claims
   are checked.** Every example must work against the current version at the
   time it is written. Where an example must name a version, it names a floating
   alias (ADR 0135) so that it keeps being true without editing.

4. **Reported output is quoted, never paraphrased.** Terminal output shown in
   documentation is copied from a real run. Paraphrasing produces text that
   looks authoritative and does not match what the reader will see — this ADR
   was written immediately after a README draft invented a plausible
   `koment check` output that differed from the real one in wording and in
   pluralisation.

The bar for "did you document it" is the same as the bar the repository already
sets for rationale: if a future reader could reasonably be misled, the change is
not done.

## Consequences

What becomes easier:

- The manual can be trusted, which is the only property that makes a manual
  worth having.
- A reviewer has something concrete to ask for: which file did this change, and
  does it still read correctly.
- Removals stop leaving orphaned paragraphs behind, which are the hardest kind
  of documentation rot to notice because nothing errors.

What becomes harder:

- Every behaviour change costs more. That cost is the point; it was previously
  being deferred onto readers.
- Some of this cannot be mechanically enforced today. Rules 1 and 2 are
  judgement, and judgement is exactly what fails under time pressure. Partial
  enforcement is possible and worth building — the command registry from ADR
  0131 knows every command koment has, so a test can assert that documentation
  never references one that does not exist, and a lint can reject an example
  pinning a version older than the current major. Neither is built by this ADR,
  and saying so is better than implying the rules are automatic.
- A documentation-only fix now sometimes needs its own commit, which slightly
  inflates history. Cheap next to the alternative.

## Alternatives rejected

- **Rely on review.** Free, no process. Rejected on the evidence above: this
  repository has a single maintainer, an agent contract, required CI and a
  strict comment policy, and it still shipped a three-major-stale install
  example on its own front page. Review did not catch any of the five failures.

- **A periodic documentation audit.** Catches drift eventually and does not
  slow ordinary work. Rejected because "eventually" is the failure mode: the
  stale pins survived several releases, and an audit would have found them at
  the same point a reader did. It also concentrates the work into a chore
  nobody volunteers for.

- **Move documentation into code comments or generated reference.** Guarantees
  proximity, and generated CLI reference genuinely cannot drift. Rejected as a
  general answer because it only works for reference material: the parts that
  matter most — why you would adopt this, how the tiers differ, what the
  workflow looks like — are prose that no generator produces. Generated
  reference where it fits is welcome and does not replace the rule.

- **Write the rules but leave `AGENTS.md` alone**, keeping them in this ADR.
  Rejected because agents read the managed contract, not the decision log, and
  agents write most of the documentation here now. A rule they never see is a
  rule that is not in force.
