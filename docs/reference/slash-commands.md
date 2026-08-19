# Slash commands

The [Claude Code plugin](../guides/agents/claude-code.md) installs eight commands. They
are the human-facing surface: you type one when you want koment to do a specific
thing now. The bundled skill is the agent-facing surface — a standing procedure
the model follows without being asked (ADR 0141).

Install the plugin, then `/reload-plugins` or restart:

```text
/plugin marketplace add koment-dev/koment
/plugin install koment@koment-dev
```

| command | argument | what it does |
|---|---|---|
| `/koment:check` | — | runs `koment check`, `comments check` and `agents check`, and reports each separately |
| `/koment:show` | `<path>` | the annotations on one file, with unresolved ones called out as history |
| `/koment:search` | `<topic>` | full-text search across annotation bodies |
| `/koment:add` | `<file> — what needs explaining>` | records rationale, after trying the four alternatives that dissolve it |
| `/koment:convert` | `<file> [line]` | records an existing comment as an annotation, then removes it from source |
| `/koment:reanchor` | `<id, file, or nothing>` | repairs a `drifted` or `ambiguous` anchor, keeping the id |
| `/koment:forget` | `<id or description>` | deletes an annotation, with confirmation and an audit trail |
| `/koment:bootstrap` | `[adapters]` | sets koment up in a repository that does not have it |

## What they refuse to do

Each command carries the constraint that makes it safe, so an agent running one
cannot take the shortcut that destroys the record:

- `/koment:check` will not report success while any gate fails.
- `/koment:forget` will not delete an annotation to make `koment check` pass —
  it points at `/koment:reanchor` instead, and offers `koment edit` before
  deletion.
- `/koment:add` will not record rationale until rename, extract, named type and
  restructure have each been tried and named.
- `/koment:convert` records before it deletes, and lists the comment classes
  that should stay in source.
- `/koment:bootstrap` leaves everything uncommitted and will not mass-convert an
  existing codebase's comments.

## Other runtimes

OpenCode, Codex and Cursor get the same behaviour through the generated
adapters and the MCP tools rather than through slash commands; see
[agent setup](../guides/agents/README.md). The commands directory is Claude Code's
format and is not portable.
