# Why editor integrations share one language server

Annotation semantics live in one process. `koment lsp` speaks the Language
Server Protocol over stdio and uses the same snapshot and mutation services as
the CLI, UI and MCP server. An editor never reimplements record validation,
anchoring or comment classification.

That boundary gives most editors a complete koment experience with a few lines
of LSP configuration. A dedicated package is justified only when a marketplace
removes manual installation or an editor API can present something LSP cannot.

## Shared capabilities

The server reports these capabilities during initialization:

| Capability | What it provides |
|---|---|
| `hoverProvider` | annotation text under the cursor |
| `codeLensProvider` | annotations above their anchor line |
| `codeActionProvider` | comment conversion, acknowledgement and reanchoring |
| `executeCommandProvider` | `koment.add`, `koment.reanchor`, `koment.convertComment` and `koment.acknowledgeComment` |
| `publishDiagnostics` | drift, orphaning, ambiguity and prohibited comments |

Text is synchronized in full on change, and saves include the text. Diagnostics
only carry conditions that fail `koment check`; a resolving annotation is not
marked regardless of where its excerpt moved.

Rendering a body beside its source line without adding it to the buffer is not
an LSP capability. Editor-specific presentation APIs decide whether that extra
surface is possible.

## Packaged editors

### VS Code and Open VSX clients

The VS Code extension starts `koment lsp` and presents rationale as an inline
signal and a panel containing the complete bodies for the file. Its six
platform packages carry the canonical released binary; the universal package
falls back to `PATH`.

Publishing the same package through Open VSX covers Cursor, Windsurf and
VSCodium. See [the extension README](../../integrations/editors/vscode/README.md).

### Zed

The Zed extension starts `koment lsp` and registers `koment mcp --write` as a
context server, so the editor and Agent Panel use the same annotations. Zed
extensions are sandboxed WebAssembly and do not carry the koment executable.

Zed exposes no decoration API, so annotation bodies appear through hover and
code lenses rather than beside the source line. Its manifest enumerates
languages because Zed has no wildcard language-server registration.

See [the Zed guide](../guides/editors/zed.md) and
[the extension manifest](../../integrations/editors/zed/extension.toml).

## Configured editors

Helix, Neovim, Emacs, Sublime, Kate and any client that can start a stdio
language server receive hover, lenses, actions and diagnostics without
koment-specific code. They cannot receive inline bodies through LSP.

Follow [configure an editor](../guides/editors/configure-an-editor.md) for the
setup. [Language support](../reference/languages.md) enumerates the comment
detectors, and [agent setup](../guides/agents/README.md) covers MCP rather than
human editor behavior.

ADR 0110 owns the shared LSP boundary, ADR 0112 owns the package threshold and
ADR 0139 extends that threshold to Zed's distribution and context-server
surface.
