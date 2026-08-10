# 0138 — Documentation has four sections, and every page belongs to one

Date: 2026-08-10
Status: Accepted

## Context

`docs/` is a flat pile of ten Markdown files plus three directories, ordered by
nothing. `quickstart.md`, `cli.md`, `publishing.md`, `releasing.md`,
`development.md` and `languages.md` sit side by side although they serve
entirely different readers at entirely different moments.

The cost is not tidiness. It is that a writer has no rule telling them where a
new page goes, so each new page lands wherever the previous one did, and the
pile grows without acquiring an order. It is also why the pages themselves blur:
`bootstrap.md` currently opens with an explanation of the data model, walks a
reader through first use, and then documents flags — three jobs in one file,
each done in the other two's way.

A reader arrives in one of four moods, and mixing them is what makes
documentation exhausting: *get me running*, *help me do this specific thing*,
*tell me exactly what this flag does*, *explain why it works like this*.

## Decision

koment adopts the four-section split popularised as
[Diátaxis](https://diataxis.fr/), and **every page under `docs/` belongs to
exactly one section**:

```
docs/
  README.md                map of the four sections

  start/        learning-oriented. Read once, in order, to get running.
                Complete, sequential, no alternatives offered.
  guides/       task-oriented. "How do I publish to Pages?" Assumes you are
                already running. May offer choices. One page per task.
    agents/       per-client setup
    editors/      per-editor setup
  reference/    information-oriented. Flags, commands, statuses, schema,
                supported languages. Dry, complete, no narrative, no opinion.
  explanation/  understanding-oriented. Why the design is this shape.
    decisions/    the ADRs
```

The rules that make it hold:

1. **One page, one section, one job.** A page that teaches, instructs *and*
   lists flags is three pages. Split it rather than filing it under whichever
   job it does most.
2. **The section decides the voice.** `start/` is imperative and sequential.
   `guides/` assumes competence and states a goal in its title. `reference/` is
   exhaustive and neutral — no "you probably want". `explanation/` argues, and
   may disagree with a guide about what is convenient.
3. **A new page names its section before it is written.** If the section is not
   obvious, the page is doing more than one job.
4. **Reference is the only section allowed to be generated**, and where it can
   be generated from the code it should be, because reference is the section
   that rots fastest and the only one a generator can produce faithfully.

`README.md` at the repository root is not part of this structure. It is the
advertisement, and it links the four sections rather than duplicating them.

## Consequences

What becomes easier:

- A writer asks one question — which of the four moods is this reader in — and
  the location follows. That question is answerable by an agent.
- The blurred pages get split, so a reader after one flag no longer reads a
  tutorial to find it.
- The docs site (ADR 0136) gets a navigation tree for free, because the
  directory structure is the tree.

What becomes harder, and one thing deliberately not done here:

- **This ADR does not perform the migration.** Thirty-one links point into
  `docs/` from the README, the workflows, the ADRs and the pages themselves,
  and moving files without repointing all of them would trade an untidy manual
  for a broken one. The structure binds new pages immediately; the existing
  files are moved in a follow-up change that updates every inbound link and
  adds a test asserting placement. Saying this plainly is the requirement of
  ADR 0137 — a structure described as done while the tree still disagrees is
  exactly the drift this project refuses to ship.
- Some pages genuinely span sections and will have to be split, which is more
  files and more cross-linking than a single page.
- Four sections is more ceremony than ten files strictly need. It is chosen for
  where the manual is going, not where it is.

## Alternatives rejected

- **Leave it flat and rely on `README.md` to organise it.** Zero migration, and
  ten files really can be listed. Rejected because the index is a second thing
  to maintain that describes the structure instead of being it, and because the
  flat pile gives a writer no rule at all — which is the actual failure.

- **Group by feature** (`publishing/`, `agents/`, `editors/`, `cli/`).
  Intuitive, and matches how the codebase is organised. Rejected because it
  answers "which subsystem" rather than "what does the reader need right now",
  so each feature directory ends up containing its own tutorial, its own guide
  and its own reference — recreating the blur inside every folder.

- **Group by audience** (`users/`, `contributors/`, `operators/`). Also
  plausible. Rejected because the same person is all three within an hour, and
  because it forces a judgement about who someone is instead of what they are
  doing.

- **Invent a bespoke split.** Rejected in favour of a documented framework with
  an established vocabulary: "this belongs in reference" is a review comment
  that needs no explanation, and an agent given the four definitions places
  pages correctly without further instruction.
