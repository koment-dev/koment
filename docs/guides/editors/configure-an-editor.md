# Configure an editor with koment LSP

Install a released `koment` binary on `PATH` and open a repository containing
`.koment/`.

Editors with a dedicated package have their own instructions:

- [Zed](zed.md)
- [VS Code](../../../integrations/editors/vscode/README.md)

For another LSP client, start:

```sh
koment lsp
```

Root the server at `.koment` or `.git` and attach it to every filetype where
you want annotations.

## Helix

Add the server to `languages.toml`:

```toml
[language-server.koment]
command = "koment"
args = ["lsp"]

[[language]]
name = "go"
language-servers = ["gopls", "koment"]
```

Repeat the `language-servers` entry for each configured language.

## Neovim

Neovim 0.11 or newer can use:

```lua
vim.lsp.config['koment'] = {
  cmd = { 'koment', 'lsp' },
  filetypes = { 'go', 'lua', 'python', 'rust', 'typescript' },
  root_markers = { '.koment', '.git' },
}

vim.lsp.enable('koment')
```

Extend `filetypes` freely. Anchoring is excerpt-based and language-agnostic.

## Verify the connection

Open a file carrying an annotation and request hover or code lenses at its
anchor. Drift, orphaning, ambiguity and prohibited comments appear as
diagnostics.

Inline annotation bodies require an editor-specific presentation API and are
not provided by LSP. See
[editor integrations](../../explanation/editor-integrations.md) for the shared
capabilities and package boundary.
