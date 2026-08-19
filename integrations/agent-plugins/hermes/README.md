# koment for Hermes Agent

This plugin blocks ordinary explanatory comments before Hermes writes them and
prevents a turn from finishing while the repository's koment gates fail.

## Install

Install a released `koment` binary on the Hermes process's `PATH`, then
extract the signed release archive into the user plugin directory:

```sh
version=<version>
plugin_home="${HERMES_HOME:-$HOME/.hermes}/plugins"
mkdir -p "$plugin_home"
curl -fsSL \
  "https://github.com/koment-dev/koment/releases/download/v${version}/koment-plugin-hermes_v${version}.tar.gz" \
  | tar -xz -C "$plugin_home"
hermes plugins enable koment
hermes plugins list
```

The archive expands to `$plugin_home/hermes/`; the manifest identifies the
plugin as `koment`. The release procedure documents checksum and Sigstore
verification before installation.

## What it does

**`pre_tool_call`** sends proposed writes and patches to `koment agents hook
pre-tool`. If the edit adds ordinary explanatory comment intent, koment refuses
it and tells the agent to record the rationale as an annotation.

**`pre_verify`** runs `koment agents hook stop` before an edited turn can
finish. That gate runs `koment check`, `koment comments check` and
`koment agents check`. Hermes bounds repeated verification with
`agent.max_verify_nudges`.

Both decisions come from the koment binary. The Python package only translates
Hermes lifecycle events into koment's shared policy surface.

## Read annotations

Add the writable MCP server to `~/.hermes/config.yaml`:

```yaml
mcp_servers:
  koment:
    command: "koment"
    args: ["mcp", "--write"]
```

See [the Hermes guide](https://github.com/koment-dev/koment/blob/main/docs/guides/agents/hermes.md)
for local, remote and filtering configurations.

## Disable temporarily

Set `KOMENT_PLUGIN_DISABLED=1` before starting Hermes. Both hooks become
no-ops without uninstalling the package.

## License

AGPL-3.0-or-later, as the rest of koment. Commercial licenses are available.
