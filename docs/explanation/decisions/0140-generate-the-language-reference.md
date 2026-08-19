# 0140 — Generate the language reference from the syntax table

Date: 2026-08-19
Status: Accepted

## Context

ADR 0132 replaced a Go-only comment gate with a marker scan over a per-extension
syntax table, so `koment comments check`, `comments convert` and
`comments acknowledge` reach every filetype bar a short prose and data
exclusion list. That shipped in 3.0.0 on 2026-08-10.

Nine days later the manual still described the world ADR 0132 removed. Read in
the source rather than assumed:

- `docs/languages.md` carried a two-row table reading `Go: yes` /
  `everything else: no`, the sentence "Outside one, those commands report
  `no comment detector for <file>`", and "**Caveat.** Detection covers `*.go`
  only."
- `docs/editors/README.md` said the editor offers conversion "only in a language
  koment can parse, which today is Go alone".
- The `Detector` godoc in `internal/commentpolicy/comments.go` said "koment
  ships one". It ships two.

Nothing failed. `koment comments check` cannot read prose, godoc on an exported
identifier is an allowed comment class, and no test compares a sentence to a
map. The repository that exists because descriptions rot silently had shipped
three descriptions that rotted silently, and ADR 0137 had already declared that
unacceptable eight ADRs earlier.

Correcting the three sentences would restore the manual for exactly as long as
the next extension takes to land.

## Decision

The language reference is generated from the table that decides it.

`internal/commentpolicy` exports the tables as data — `DetectedFiletypes`,
`FallbackMarkers`, `UndetectedExtensions`, `ScriptFilenames` and `DetectorName`
— and `internal/commentpolicy/gen` renders `docs/reference/languages.md` from
them. A `//go:generate` directive above `syntaxByExtension` binds the two, so
the page is produced by the same edit that changes the behaviour.

The existing `generate-check` task already runs `go generate ./...` and rejects
a dirty tree, so CI gates this with no new job. Adding an extension to
`syntaxByExtension` without regenerating is a red build.

The page moves to `docs/reference/` and `docs/languages.md` is deleted. ADR 0138
says reference is the only section that may be generated and should be where the
code can produce it faithfully; an exhaustive per-extension table of markers and
passed-through directives is exactly that.

Prose the code cannot produce — why `--` is absent from the fallback, why the Go
parser is tried first — stays in the generator's template rather than moving to
an unchecked page, so one file produces the whole page.

## Consequences

- A filetype cannot enter koment without entering the manual in the same commit.
- `internal/commentpolicy` gains five exported identifiers that exist for the
  documentation rather than for callers. `support_test.go` holds them to the
  tables so an accessor cannot quietly stop covering one.
- The narrative sections of the page now live in Go string literals, which is a
  worse place to edit prose than a Markdown file. That cost is accepted: split
  across two files, the generated half would be true and the hand-written half
  would rot, which is the situation this ADR exists to end.
- The repository gains its first production `//go:generate`. It is a toolchain
  pragma, so it is an allowed comment class and needs no acknowledgement.
- Anyone linking `docs/languages.md` gets a 404. Both in-repository links were
  repointed; the ADR 0138 mention is left as history.

## Alternatives rejected

**Fix the three sentences.** The cheapest change and the one that fails again.
The table has grown twice since ADR 0114 introduced the `Detector` interface,
and prose lost the race both times. A repository selling drift detection cannot
answer its own drift with a manual edit.

**Add a test asserting the page mentions every extension.** Keeps the page
hand-written and catches omissions, but not staleness: a row saying `.rs` is
undetected passes a mention test. It also makes the failure "go edit the docs"
rather than producing them.

**Generate only the table and `{{include}}` it into a hand-written page.**
Markdown has no include, so this needs a second generator with a marker-block
protocol — the shape `koment agents install` already uses for managed blocks. It
was rejected because the boundary is arbitrary: the prose around the table makes
claims about the table ("only these report `no comment detector`"), so the half
that could rot is the half that describes what the generated half means.

**Name each filetype ("YAML", "Rust") rather than listing extensions.** Reads
better, but the table is keyed by extension and carries no names, so the
generator would need a second hand-maintained map — a new thing to rot, to fix
rot. A reader asking "does koment detect `.tf`?" is served better by the key
they already have.
