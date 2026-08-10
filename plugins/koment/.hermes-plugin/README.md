# koment for Hermes Agent

Keeps rationale out of source comments and stale annotations out of the tree,
in the agent that is writing the code.

## Install

```sh
hermes plugins install koment-dev/koment-hermes
```

Then put the plugin in your `~/.hermes/config.yaml`:

```yaml
plugins:
  enabled:
    - koment
```

The `koment` binary must be on the `PATH` of the Hermes process:

```sh
brew install koment-dev/tap/koment
```

## What it does

**`pre_tool_call`** — before Hermes writes or patches a file, the content goes
to `koment agents hook pre-tool`. If the edit adds an ordinary explanatory
comment, the write is refused and the agent is told to record the reasoning as
an annotation instead:

> koment policy blocked ordinary comment intent (retry.go: // retry 3 times
> because the API is flaky). Record the rationale against nearby code with
> `koment_add` instead.

**`pre_verify`** — before a turn that edited code is allowed to finish,
`koment agents hook stop` runs `koment check`, `koment comments check` and
`koment agents check`. A failure returns `{"decision": "block"}`, which keeps
the agent working rather than letting it stop with the repository in a state
its own policy rejects. Hermes bounds this with `agent.max_verify_nudges`.

Both decisions are made by the koment binary, not here. This plugin translates
Hermes' tool shapes into the one koment already reads, so it cannot disagree
with the Claude Code hooks, the OpenCode plugin or CI about what the policy is.

## Reading annotations

The gate is half the story. To let Hermes *read* the reasoning before it edits,
add the MCP server as well — see
[docs/agents/hermes.md](https://github.com/koment-dev/koment/blob/main/docs/agents/hermes.md):

```yaml
mcp_servers:
  koment:
    command: "koment"
    args: ["mcp", "--write"]
```

## Turning it off

```sh
KOMENT_PLUGIN_DISABLED=1
```

Both hooks become no-ops without uninstalling. The plugin also stays silent
when `koment` is not on the `PATH`, so a machine without it installed is not
blocked from working.

## Licence

AGPL-3.0-or-later, as the rest of koment. Commercial licences available.
