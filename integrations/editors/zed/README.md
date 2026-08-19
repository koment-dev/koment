# koment for Zed

Rationale that lives beside the code instead of inside it. This extension starts
`koment lsp` for your worktree and registers `koment mcp --write` as a context
server, so both you and Zed's agent read the same annotations.

## Install

Install the released `koment` binary, then find **koment** in Zed's
**Extensions** view. Until the first registry pull request merges, install this
directory as a dev extension instead.

```sh
brew install koment-dev/tap/koment
```

The extension does not carry a binary. Zed extensions run as sandboxed
WebAssembly and cannot ship an executable for your platform the way the VS Code
package does (ADR 0113); it finds the one you installed instead.

Only this extension directory is licensed under
[GPL-3.0-or-later](LICENSE), as recorded in ADR 0145. The koment binary and the
rest of the repository remain AGPL-3.0-or-later.

## What you get

| | |
|---|---|
| hover | the annotation body under the cursor |
| code lens | annotations above the line they anchor to |
| code actions | convert a comment, acknowledge it, reanchor a stale record |
| diagnostics | drift, orphaning, ambiguity and prohibited comments |
| context server | `koment_get`, `koment_search` and the write tools, in the Agent Panel |

Diagnostics carry only what fails `koment check`. An annotation that resolves is
never marked, wherever its excerpt has moved to.

**Annotation bodies are not rendered beside their line.** That needs an editor
decoration API, and Zed extensions have none — it is the one thing the VS Code
package does that this cannot (ADR 0139). Hover and the code lens carry the same
text.

## If it does not start

The extension looks for `koment` on `$PATH`. A Zed launched from the Finder or
the Dock does not inherit the `$PATH` your shell sets, so point it at the binary:

```json
{
  "lsp": {
    "koment": {
      "binary": {
        "path": "/opt/homebrew/bin/koment"
      }
    }
  }
}
```

The context server takes the same override under `context_servers`:

```json
{
  "context_servers": {
    "koment": {
      "command": {
        "path": "/opt/homebrew/bin/koment"
      }
    }
  }
}
```

The repository needs a `.koment/` directory. Run `koment bootstrap` in one that
does not have it.

## Languages

`extension.toml` names the languages the server attaches to. Zed has no wildcard,
so the list is explicit even though koment's anchoring is language-agnostic and
works in any file. A language that is missing needs a line in that list and a
release — open an issue.

## Develop

```sh
mise run build
mise run check
```

Then **zed: install dev extension** from the command palette and choose this
directory.
