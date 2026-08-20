# @koment/opencode-koment

The OpenCode plugin keeps one writable koment MCP process alive for the session,
checks proposed edits through that process and runs all repository gates when
the session closes.

## Install

Add the npm package to `opencode.json`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "plugin": ["@koment/opencode-koment"]
}
```

[OpenCode installs configured npm packages at startup](https://dev.opencode.ai/docs/plugins/).
Restart it after changing the configuration.

## Requirements

- A released `koment` binary on `PATH`
- `.koment/policy.yaml` in every repository where the plugin should enforce
  koment policy

Bootstrap a repository with:

```sh
koment bootstrap --agents opencode --non-interactive
```

## What it does

1. At session load, it starts `koment mcp --write` over stdio.
2. Before an edit or write, it calls `koment_pre_tool` and denies ordinary
   explanatory comment intent.
3. At disposal, it runs `koment check`, `koment comments check` and
   `koment agents check`.

The package does not duplicate comment detection or annotation resolution.
Without a policy or annotation records, its policy calls are silent no-ops.
Annotation records without the policy are treated as incomplete configuration
and fail with a bootstrap instruction.

## Verify

```sh
koment check
koment comments check
koment agents check
```

All three must pass before a session can complete.

## Generated adapter

`koment bootstrap` also generates `.opencode/plugins/koment.js` for a
repository. The generated adapter and npm package share the same policy
protocol; they differ only in distribution. Do not load both in the same
OpenCode configuration.
