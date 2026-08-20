# Agent plugins

Each child directory is one self-contained integration shipped independently
from the others.

| Runtime | Installable root | Distribution |
|---|---|---|
| Claude Code | [`claude/`](claude/) | koment Claude marketplace and signed release archive |
| Hermes | [`hermes/`](hermes/) | signed release archive |
| OpenCode | [`opencode/`](opencode/) | npm and signed release archive |

All three delegate policy decisions to koment rather than implementing a second
comment classifier.

## Claude Code

Install a released `koment` binary on `PATH`, bootstrap the repository, then
add this repository's marketplace:

```text
/plugin marketplace add koment-dev/koment
/plugin install koment@koment-dev
```

Project scope is the smallest installation boundary. A user-scoped installation
is also inert in repositories with neither `.koment/policy.yaml` nor annotation
records. Annotation records without the policy remain a visible incomplete
configuration. Restart Claude Code or run `/reload-plugins` after installation.

The package contributes the writable MCP declaration, hooks, the standing
`koment` skill and the slash commands documented in
[the command reference](../../docs/reference/slash-commands.md).

## Hermes

Install the released archive into Hermes's user plugin directory:

```sh
version=<version>
plugin_home="${HERMES_HOME:-$HOME/.hermes}/plugins"
mkdir -p "$plugin_home"
curl -fsSL \
  "https://github.com/koment-dev/koment/releases/download/v${version}/koment-plugin-hermes_v${version}.tar.gz" \
  | tar -xz -C "$plugin_home"
hermes plugins enable koment
```

The archive expands to one directory containing `plugin.yaml`,
`__init__.py` and its README. See [the Hermes package](hermes/README.md).

## OpenCode

Add the published npm package to `opencode.json`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "plugin": ["@koment/opencode-koment"]
}
```

OpenCode installs configured npm plugins at startup. The package keeps one
writable MCP process for the session and applies the same pre-tool and
completion policy as the generated repository adapter.

See [the OpenCode package](opencode/README.md).

## Generated repository adapters

`koment agents install` writes client-owned discovery files such as
`.mcp.json`, `.codex/config.toml`, `.cursor/mcp.json`,
`.opencode/plugins/koment.js` and `.vscode/mcp.json`. Those files belong at
fixed repository-root paths and are separate from the installable plugin
sources in this directory. `koment agents check` rejects adapter drift.
