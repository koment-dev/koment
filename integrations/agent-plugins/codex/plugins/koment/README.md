# koment for Codex

This plugin gives Codex the writable koment MCP tools, repository policy hooks
and the standing koment skill.

## Install

Install a released `koment` binary on `PATH`. Add this plugin's containing
marketplace, then run:

```sh
codex plugin marketplace add /path/to/koment-codex-marketplace
codex plugin add koment@koment-dev
```

A release that contains this plugin provides the same marketplace in a signed
archive.

Codex asks you to trust the plugin hooks before they run. Review and accept the
exact hook definitions. Start a new thread after installation.

## Activate a repository

Run this command in each repository where the plugin must enforce koment:

```sh
koment bootstrap --agents agents,codex --non-interactive
```

The plugin stays silent in a repository with neither `.koment/policy.yaml` nor
annotation records. Annotation records without the policy are an incomplete
configuration and request `koment bootstrap`.

## What it does

1. It starts `koment mcp --write` from the active workspace.
2. It loads the standing koment procedure at session start and after compaction.
3. It checks `apply_patch` calls before Codex writes ordinary comment intent.
4. It runs the annotation, comment and adapter gates before Codex completes.

The `koment` executable owns every policy decision. The plugin does not
implement another comment classifier or annotation store.

## Generated adapter

`koment agents install` also generates `.codex/config.toml` and
`.codex/hooks.json` for one repository. Use either the installable plugin or the
generated hooks. Do not load both hook configurations.
