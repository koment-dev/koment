---
description: Run every koment gate and report what actually passed
argument-hint: ""
---

Run all three gates. They check different things and any one of them can fail
alone:

```sh
koment check
koment comments check
koment agents check
```

- `koment check` resolves every anchor and exits non-zero when one is
  `ambiguous`, `drifted` or `orphaned`.
- `koment comments check` classifies every comment in the repository and rejects
  the ordinary ones.
- `koment agents check` reports drift in the generated agent adapters.

Report each result separately, and **quote the real output**. "Gates pass" is
not a report; copy the final line each command actually printed.

If one fails:

- **Do not report success.** A partial pass is a failure.
- **Do not delete an annotation to make `koment check` green.** A `drifted`
  record means the code moved out from under rationale someone recorded on
  purpose. Fix the anchor with `/koment:reanchor`, or fix the code. Deleting it
  destroys the only thing that knew why the code was that way.
- A `comments check` failure names the file and line. Try the four alternatives
  in order — rename, extract a function, introduce a named type or constant,
  restructure — before recording the comment as an annotation with
  `/koment:convert`.
- An `agents check` failure usually means `koment agents install` needs
  re-running after a contract change. Read the drift it reports before
  regenerating; the diff is the interesting part.
