# Codex

Codex can install the complete koment plugin or use generated project files.
Both forms require a released `koment` binary on `PATH`.

## Install the plugin

Install directly from a koment checkout:

```sh
codex plugin marketplace add /path/to/koment/integrations/agent-plugins/codex
codex plugin add koment@koment-dev
```

A release that contains the Codex plugin also includes the signed
`koment-plugin-codex_v<version>.tar.gz` archive. Extract it and add the
`koment-codex-marketplace` directory with the same two commands.

Review and trust the exact hook definitions when Codex prompts you. Start a new
thread after installation. The plugin provides the writable MCP server, the
standing koment skill, a pre-tool hook and a Stop hook.

Run this command in each repository where the plugin must enforce policy:

```sh
koment bootstrap --agents agents,codex --non-interactive
```

Without `.koment/policy.yaml` or annotation records, the global plugin stays
silent. Annotation records without the policy remain visible as incomplete
configuration and request bootstrap.

## Generate project configuration

`koment agents install` writes the project MCP entry and supported Codex hooks.

```sh
codex mcp add koment -- koment mcp --write
```

That writes the entry for you. Or do it by hand:

```toml
[mcp_servers.koment]
command = "koment"
args = ["mcp", "--write"]
```

Codex keeps MCP servers in TOML. Project-scoped configuration lives at
`.codex/config.toml` and applies only to trusted projects. Setting `cwd` pins
the repository explicitly:

```toml
[mcp_servers.koment]
command = "koment"
args = ["mcp", "--write"]
cwd = "/path/to/your/repo"
```

## Verify

```sh
codex plugin list
codex mcp list
```

The installed plugin and `koment` MCP server should appear. Then ask for
something you can check against `koment show <file>`.

## Make it use them

Codex reads the managed contract in `AGENTS.md`. Both hook forms deny ordinary
explanatory comment intent in `apply_patch`. Their Stop hook checks annotations,
comments and adapters before the turn can finish. Run `koment agents check` in
CI because hooks remain a workstation guardrail, not the authoritative
boundary.

## Notes

- TOML table names are the server name: `[mcp_servers.koment]` produces a server
  called `koment`. Several repositories means several tables with distinct names
  and their own `cwd`.
- Codex also supports `env` (a nested `[mcp_servers.koment.env]` table) and
  `env_vars` for forwarding existing variables. koment needs neither — it reads
  local files and takes no configuration.
- Do not load the plugin hooks and generated `.codex/hooks.json` together.
