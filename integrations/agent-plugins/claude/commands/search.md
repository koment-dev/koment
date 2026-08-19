---
description: Find recorded rationale by topic across the repository
argument-hint: "<topic>"
---

Search every annotation body for:

$ARGUMENTS

Call `koment_search`. The match is case-insensitive full text over bodies, so
phrase it as the words someone would have written when explaining the decision
— "rate limit", not "why does this retry".

Do this before any non-obvious structural decision. Another file has often
already settled the question, and the record explains what was rejected.

When reporting:

- Group by file, and give the annotation's kind and status with each hit.
- A hit whose status is not `ok` is history. Quote it as such.
- If nothing matches, say so and proceed from first principles. Do not
  reconstruct a plausible rationale and present it as something koment
  recorded — that is precisely the failure koment exists to prevent.
- Broaden once before giving up: search terms are matched literally, so a
  single synonym often finds it.
