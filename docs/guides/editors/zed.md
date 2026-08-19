# Annotate code in Zed

Zed reads koment through two channels, and the extension turns both on at once:
the language server gives *you* hover, code lenses, quick fixes and drift
diagnostics, and the context server gives Zed's **agent** the same annotations
through MCP.

## Install

1. Install a released `koment` binary.

   ```sh
   brew install koment-dev/tap/koment
   ```

2. Find **koment** in Zed's **Extensions** view and install it. Until the first
   registry pull request merges, run
   `mise --cd integrations/editors/zed run build`, choose **zed: install dev
   extension** from the command palette, and select
   `integrations/editors/zed/` instead.

3. Open a repository containing `.koment/`. Run `koment bootstrap` first if it
   has none.

The extension does not carry a binary. Zed extensions are sandboxed WebAssembly
and cannot ship an executable, so unlike the VS Code package (ADR 0113) this one
finds the `koment` you installed.

The extension code alone is
[GPL-3.0-or-later](../../../integrations/editors/zed/LICENSE), which Zed accepts
for registry distribution. The binary it starts and the rest of koment remain
AGPL-3.0-or-later (ADR 0145).

## Check it started

Open an annotated file. Hovering a line that carries rationale shows the
annotation body, and a code lens sits above it.

For the agent side, open the Agent Panel — `koment` is listed under its context
servers with the read and write tools. Ask it something concrete and compare:

```sh
koment show internal/store/ulid.go
```

## When it does not start

The extension looks for `koment` on `$PATH`. **A Zed launched from the Finder or
the Dock does not inherit the `$PATH` your shell sets**, which is the usual
cause. Give it an absolute path in your settings:

```json
{
  "lsp": {
    "koment": {
      "binary": {
        "path": "/opt/homebrew/bin/koment"
      }
    }
  },
  "context_servers": {
    "koment": {
      "command": {
        "path": "/opt/homebrew/bin/koment"
      }
    }
  }
}
```

Open the settings file with `zed: open settings file` from the command palette.

## What Zed does not get

**Annotation bodies are not rendered beside their line.** That needs an editor
decoration API and Zed extensions have none — it is the one thing the VS Code
package does that this cannot (ADR 0139). Hover and the code lens carry the same
text, so nothing is unreachable; it is one interaction away instead of always on
screen.

## Languages

The extension names the languages it attaches to explicitly, because Zed has no
wildcard for this. koment's anchoring is language-agnostic and works in any file,
so the list is a limitation of the channel rather than of koment.

If your language is missing, koment still works from the CLI, `koment ui` and the
agent's MCP tools in that repository — only the in-editor surface is absent.
[Open an issue](https://github.com/koment-dev/koment/issues) and it gets a line
in `integrations/editors/zed/extension.toml`.

## Without the extension

Any LSP client can run `koment lsp` directly, and Zed can register the MCP server
by hand. See [the Zed agent guide](../agents/zed.md) for the settings, and
[editor integrations](../../explanation/editor-integrations.md) for what the server provides.
