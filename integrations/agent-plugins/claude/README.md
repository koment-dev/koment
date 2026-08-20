# koment for Claude Code

This plugin gives Claude Code the koment MCP tools, repository policy hooks,
the standing koment skill, and slash commands for creating and maintaining
code annotations.

## Install

Install a released `koment` binary on `PATH`, run `koment bootstrap` in the
repository, then add the marketplace and install the plugin:

```text
/plugin marketplace add koment-dev/koment
/plugin install koment@koment-dev
```

Restart Claude Code or run `/reload-plugins` after installation. Project scope
is the smallest boundary. A user-scoped installation stays silent in
repositories with neither `.koment/policy.yaml` nor annotation records; records
without the policy are treated as incomplete configuration and request
bootstrap.

## Use

The plugin exposes `/koment:add`, `/koment:show`, `/koment:search`,
`/koment:check`, `/koment:convert`, `/koment:reanchor`, `/koment:forget`, and
`/koment:bootstrap`. The full command reference is available in the
[koment documentation](https://github.com/koment-dev/koment/blob/main/docs/reference/slash-commands.md).

The `koment` executable remains the policy authority. The plugin does not
implement a separate comment classifier or annotation store.
