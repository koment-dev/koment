# 0151 — Package a Codex agent plugin

Date: 2026-08-25
Status: Accepted

## Context

Codex now supports installable plugins with a `.codex-plugin/plugin.json`
manifest, skills, MCP configuration and lifecycle hooks. The existing Codex
integration is only a generated repository adapter. Claude and OpenCode also
have independently packaged plugins, so Codex users cannot install the same
policy surface without changing every repository.

The packaging decision must preserve earlier failure lessons. ADR 0118 keeps
comment classification in one Go implementation. ADR 0144 requires a missing
binary to fail visibly instead of leaving a plugin that appears active. ADR
0149 makes `.koment/policy.yaml` the automatic activation boundary, so a global
installation must stay silent in unrelated repositories. Release packaging
must include every file the manifest or hooks reference.

The [OpenAI plugin package specification](https://developers.openai.com/plugins/build/plugins)
requires a `.codex-plugin/plugin.json` manifest. Codex discovers plugin hooks
from `hooks/hooks.json`, and requires users to trust their exact definitions
before they run.

## Decision

Add a self-contained Codex marketplace at
`integrations/agent-plugins/codex/`. Its installable plugin is at
`plugins/koment/`, which matches its `koment` manifest name and the Codex
marketplace path contract. The `codex/` host directory is another instance
inside the agent-plugin area defined by ADR 0143. This does not create a new
repository boundary.

The plugin contains:

- `.codex-plugin/plugin.json` with the repository release version;
- `.mcp.json` starting `koment mcp --write` from the active workspace;
- `skills/koment/SKILL.md` with the standing repository procedure;
- `hooks/hooks.json` with `SessionStart`, `PreToolUse` and `Stop` hooks;
- `scripts/session-start.sh`, which emits guidance only for an active
  repository; and
- a README that states the binary, policy and trust requirements.

The pre-tool and Stop hooks call `koment agents hook pre-tool` and `koment
agents hook stop`. They do not classify comments or resolve annotations in the
plugin. The session-start script uses `koment agents check`, whose empty output
is the inactive response. Therefore Claude, OpenCode and Codex all use the same
activation decision. A repository with neither policy nor annotation records
stays silent. Annotation records without a policy remain an incomplete
workspace and request `koment bootstrap`, as ADR 0149 requires.

The plugin requires a released `koment` binary on `PATH`. It does not bundle a
second binary. Release Please updates its manifest with the repository version.
The plugin release job builds a signed archive that expands into a Codex local
marketplace with the plugin at `plugins/koment`. The archive supplies the
marketplace metadata that `codex plugin marketplace add` needs without adding
a second discovery tree to the repository root.

The generated `.codex/config.toml` and `.codex/hooks.json` adapter remains
available for repository-scoped installation. CI validates both forms,
packages every referenced plugin file and exercises the inactive repository
contract. The source tree is also a valid local marketplace, so it can be
installed before its first release archive exists. Public universal-directory
publication is not claimed until the external submission is accepted.

## Consequences

- Claude, OpenCode and Codex each have a first-class installable plugin.
- A global Codex plugin stays inert outside repositories that adopted koment.
- The plugin archive can be installed through Codex without copying files into
  a repository.
- The plugin manifest adds one more release-version file that must stay in
  lockstep.
- Hook trust remains a user decision. The plugin cannot claim enforcement when
  a user declines the Codex trust prompt.
- The released binary remains a separate prerequisite, so installing the
  plugin alone is insufficient.
- Universal-directory acceptance remains external work and cannot be verified
  by repository CI.

## Alternatives rejected

**Keep only the generated Codex adapter.** This preserves project scope, but it
does not provide the plugin coverage requested for globally reusable clients.
Every repository would still need generated client configuration.

**Reuse the Claude plugin directory for Codex.** Claude commands and manifest
metadata are host-specific. One mixed directory would recreate the incomplete
artifact risk that ADR 0143 removed.

**Implement activation and comment checks in plugin scripts.** A second policy
implementation would drift from Claude, OpenCode and CI. The binary already
owns the complete activation and policy rules.

**Bundle the koment binary in the plugin.** This would duplicate the canonical
release artifact and create a plugin-to-binary compatibility matrix. ADR 0109
requires consumers to use the authenticated platform archive.

**Add a repository-root Codex marketplace.** Codex requires a marketplace
discovery path and a separate `plugins/koment` source path. Duplicating the
self-contained integration at the repository root would create two owners for
one artifact. Building that layout only inside the release archive preserves
the closed repository contract.

**Claim universal-directory availability with the first archive.** Packaging
is controlled by this repository. Directory review and acceptance are not, so
documentation must keep that boundary visible.
