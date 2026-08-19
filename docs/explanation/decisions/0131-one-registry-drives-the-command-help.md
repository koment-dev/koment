# 0131 — One registry drives the command help

Date: 2026-08-09
Status: Accepted

## Context

`koment` with no arguments printed a hand-maintained string listing seventeen
invocations. It had drifted in the ways a hand-maintained string always does:

- Three of the seventeen commands carried a description. The other fourteen
  offered a bare synopsis, so a newcomer could not tell what `reanchor` or
  `serve` were for without running them.
- The three descriptions were aligned by hand to two different columns —
  `site` at one, `lsp` at another.
- `koment comments acknowledge ...` ran to 101 characters and wrapped on a
  standard terminal, which is where the listing looked worst.
- Every command was crammed with its flags, so the list read as syntax rather
  than as a menu.

Per-command help was worse in a way that mattered more. `flag.FlagSet` prints
`Usage of add:`, which names neither `koment` nor the positional arguments the
command needs. `koment check --help` printed the header and nothing else,
because `check` declares no flags — a reader learned less than from the
listing they had just left. And every `--help` exited **2**: asking a command
how it works was reported as misuse, which breaks the ordinary shell idiom of
gating on the exit status.

## Decision

The command table is data. One `sections` registry holds every top-level
command with its group, its argument synopsis and a one-line summary; a
`subcommands` registry holds the nested forms (`comments convert`,
`agents install`). Both the top-level listing and each command's `--help`
render from it, so the two surfaces cannot disagree about what exists.

Column alignment is computed from the widest name rather than typed, so it
cannot rot. Groups (`Getting started`, `Annotations`, `Policy`,
`Read and share`, `Integrations`) replace one flat list of seventeen. Flags
move out of the listing and into per-command help, which the epilogue points
at.

`flagSet` — already the single constructor every command uses, and already
passed the canonical name including subcommand names — sets `flags.Usage` from
the registry. A command gets a proper header, a `koment <name> <args>`
synopsis and its flags for free, and cannot opt out.

Parsing routes through one `parse` helper that separates the two outcomes
`flag` conflates: `flag.ErrHelp` exits **0**, any other parse error exits
**2**.

Three tests hold this in place: every listed command must dispatch, every
subcommand must resolve to its own usage, and every summary must start in the
same column.

## Consequences

What becomes easier:

- Adding a command means adding a row. Forgetting to document it fails
  `TestEveryListedCommandIsDispatchable` rather than shipping.
- `koment <anything> --help` is scriptable, because it exits 0.
- The listing answers "what is this for" for all seventeen commands instead of
  three.

What becomes harder, and one thing left undone:

- **The five injected servers still print their own help.** `ui`, `site`,
  `serve`, `mcp` and `lsp` parse flags inside their own packages, so
  `koment ui --help` renders in a different style — prose, then an indented
  synopsis — from `koment add --help`. The registry lists and describes them,
  so the top-level listing is uniform, but the second level is not. Unifying it
  means either exporting the renderer across four packages or moving flag
  parsing out of them, and ADR 0007 deliberately keeps `internal/mcp` and
  `internal/ui` from being imported here. That is a larger change than this
  one and is not attempted; it is recorded so the next reader knows the
  asymmetry is known rather than overlooked.
- The registry is a second place to edit when a command's flags change. The
  synopsis string can go stale against the real flags — the tests check that a
  command exists and is described, not that its synopsis is accurate.

## Alternatives rejected

- **Keep the string and just tidy it.** An hour's work and no new
  abstraction. Rejected because it fixes today's misalignment without
  removing the cause: the next command added by hand re-ranks the columns, and
  nothing fails when it does.

- **Adopt a CLI framework (`cobra`, `urfave/cli`, `kong`).** All of them solve
  this and much more. Rejected under ADR 0010's dependency bar: the standard
  library already parses flags here, the whole surface is one file of data plus
  sixty lines of rendering, and a framework would rewrite every command's
  argument handling to buy formatting.

- **Generate the listing by reflecting over the dispatch map.** Removes the
  duplicate list of names entirely. Rejected because the map holds functions,
  not descriptions — summaries, groups and ordering have to be written
  somewhere regardless, and a map has no order. The test that every listed
  command dispatches recovers most of the safety without the machinery.

- **Print per-command help to stdout instead of stderr.** More correct for an
  explicit `--help`. Rejected for now because `flag.FlagSet` writes usage and
  parse errors to one configured output; splitting them means intercepting
  `ErrHelp` before `flag` prints, which is a larger change to every command
  than the exit code alone. The exit code was the part that broke scripts.
