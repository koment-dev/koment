# 0142 — The note column carries its own height, and moves to the line on a phone

Date: 2026-08-19
Status: Accepted

## Context

The reading view puts code in one column and annotations in another, and the
browser nudges each note down to meet its line. That shape exists for a measured
reason recorded in 2026-08: with notes inside the code grid, a row grew to its
tallest cell, and ten annotations turned a five-line `const` block into about
thirty-five rendered lines. Separating the columns is what keeps line heights
uniform however much prose hangs off any one line.

Two consequences of that design had not been faced.

**The column has no height.** Aligned notes are `position: absolute`, so the
gloss contributes nothing to layout and `.reading` — a grid with
`align-items: start` — sizes its row from the code alone. The placement loop
walks a monotonic floor downward with no bound. When the stacked notes are
taller than the file, they hang past the end of the code with an empty column
beside them, absorbed only by a `6rem` bottom padding. On a short file carrying
long rationale that is most of what the reader sees, and it is what prompted
this ADR.

**On a phone the notes are nowhere near their code.** The single-column
breakpoint stacks the gloss after the code, and the gloss is emitted after the
*entire* code element — so a 500-line file puts every annotation 500 lines from
the thing it explains. The annotation on that branch asserted the opposite:
"on a narrow screen the note already sits directly under the line it annotates".
It never did. The anchor resolved cleanly the whole time, so `koment check`
could not catch it, and the claim survived because nothing compares a sentence
to a layout.

A third, smaller fault sat inside the placement loop: `.aligned` was added after
measuring, so every first paint computed offsets against the flow layout and
applied them to the absolute one. It self-corrected on the `document.fonts.ready`
re-run, which turned a wrong layout into an intermittent jump.

## Decision

**The gloss is told its height.** The placement loop already knows the bottom of
the last note, so it sets `gloss.style.height` from that floor and clears it
before measuring. The grid row now grows with the notes and the page ends where
the content ends. The separate-columns property is untouched — the code is still
never moved by a note.

**A long body folds at a paragraph boundary, server-side.** `internal/ui/view.go`
splits a note's paragraphs into a visible lead and a folded tail rendered inside
`<details>`. Both thresholds were measured against this repository's 124 bodies —
median 385 characters, p90 822 — and set so the median note is untouched and the
top sixth fold. A paragraph is never cut, so a single long one renders whole;
that is safe now the column carries its height.

**On a narrow viewport each note is moved to sit after its row.** This reverses
the earlier position that translating notes toward a line "would break that
reading order to chase an alignment there is no room for". The reasoning held for
*translating* a note within a column; it does not hold for relocating it into the
flow, where there is no alignment to chase and no other note to collide with. An
interleaved note is `position: sticky; left: 0` so a horizontal code scroll does
not carry the rationale off-screen, and width-clamped in `vw` so it cannot widen
the `max-content` code block.

Both remain enhancements, which is the constraint the whole file is held to
(ADR 0026): with scripting off the page still serves complete code followed by a
labelled column of notes in line order. The relocation is idempotent and reversed
when the viewport widens.

## Consequences

- A file shorter than its rationale renders as a page rather than as a column
  running off the end of one.
- The alignment is right on first paint instead of on the second pass.
- Some rationale is now behind a click that was not before. `<details>` keeps the
  text in the document, so a reader without scripting, a printer and a search
  engine all still get it — but it is one interaction further away.
- The reading view now has three layouts to keep working — aligned, folded and
  interleaved — where it had one and a degradation. The Go half of that is
  tested; the CSS and the relocation are not, and this repository has no browser
  in CI to test them in.
- `--gloss` became `clamp(16.5rem, 24vw, 23rem)`, so the code column is no longer
  crushed just above the 900px breakpoint.

## Alternatives rejected

**Place notes with CSS grid rows instead of transforms.** The modern shape:
give each note `grid-row: <its line>` and delete the placement code entirely. It
reintroduces the exact defect the two-column layout was built to fix — a note in
a grid row grows that row and pushes the code apart — and the repository has the
measurement to prove it. Spanning rows only moves the threshold.

**Emit notes interleaved and lift them into a column on wide screens.** Mobile
source order for free, and the existing absolute positioning would do the
lifting. Rejected because it inverts which layout survives without scripting: a
desktop reader with JS off would get notes wedged between code lines instead of
the column ADR 0026 promises.

**Render the notes twice and hide one copy per breakpoint.** No DOM moves, pure
CSS. Rejected for doubling the page weight and putting every annotation body in
the document twice, where assistive technology and search both have to be told
which copy is real.

**Clamp long bodies with CSS `max-height` and a fade.** Simpler than splitting in
Go, and needs no template change. Rejected because a scriptless reader would then
have no way to reach the hidden text at all, which is the one thing this file is
not allowed to do.

**Leave mobile alone and add a jump link from each marked line to its note.**
Cheap and safe. Rejected as navigation standing in for layout: it asks the reader
to leave their place in the code to read one sentence, and then find their way
back.
