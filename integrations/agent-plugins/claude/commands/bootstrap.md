---
description: Set koment up in a repository that does not have it yet
argument-hint: "[adapter names, e.g. claude,cursor,opencode]"
---

Bootstrap this repository:

$ARGUMENTS

Check first. If `.koment/` already exists, this repository is set up — run
`/koment:check` instead and report what it says.

Confirm a released binary is installed (`koment version`), then:

```sh
koment bootstrap
```

Pass `--agents <names>` to choose adapters and `--non-interactive` when running
unattended. Available adapters are `claude`, `copilot`, `cursor`, `codex`,
`opencode` and `agents`.

It writes:

- `.koment/policy.yaml` — the comment policy this repository enforces
- `AGENTS.md` — the managed contract every client reads
- per-adapter files: `.mcp.json` and `CLAUDE.md`, `.cursor/rules/koment.mdc` and
  `.cursor/mcp.json`, `.codex/hooks.json` and `.codex/config.toml`,
  `.opencode/plugins/koment.js` and `opencode.json`,
  `.github/copilot-instructions.md` and `.vscode/mcp.json`

Then:

1. Run `koment agents check` and confirm it reports `agent policy: ok`.
2. Run `koment comments check`. On an existing codebase this usually fails, and
   the count is the honest starting point — report it rather than hiding it.
3. **Do not mass-convert those comments.** Each one needs the rename / extract /
   named-type / restructure ladder tried first. Offer to work through them, or
   suggest raising the policy's strictness gradually.

Leave everything uncommitted and tell the user what changed. Enabling a gate
across someone's repository is their decision to commit, not yours.
