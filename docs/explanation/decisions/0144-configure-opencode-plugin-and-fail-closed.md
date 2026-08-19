# 0144 — Configure the OpenCode plugin and fail closed

Date: 2026-08-19
Status: Accepted
Supersedes: 0129

## Context

ADR 0129 chose a first-class OpenCode package, a direct JSON-RPC client for the
narrow MCP surface and a generated repository adapter for users who do not want
the package. Those boundaries remain useful, but two operating assumptions in
the decision are false.

OpenCode has no `opencode plugin install` command. Its documented stable
surface is a package name in the `plugin` array of `opencode.json`; OpenCode
installs configured npm packages with Bun at startup. A postinstall warning is
not a dependable part of that path. The package must describe configuration,
not a command that does not exist.

The plugin also cannot enforce pre-tool policy without a koment binary. Its
first action is to spawn `koment mcp --write`, and its disposal hook invokes the
three CLI gates. The old package metadata said the plugin would load and its
pre-tool MCP gate would run without the binary, which contradicts the code.

The hand-written JSON-RPC client silently ignored malformed server output and
carried a hard-coded `0.1.0` client version. A malformed response could leave a
request pending forever, and every release advertised the wrong client version.
That conflicts with the project's fail-loud rule and one-version release model.

## Decision

Keep both OpenCode integration forms:

- `koment agents install` generates the project-local adapter at
  `.opencode/plugins/koment.js` and registers it in `opencode.json`.
- The published package remains `@koment/opencode-koment`. A user installs it by
  adding that package name to the `plugin` array in `opencode.json`; OpenCode
  performs installation at startup.

The two forms must not be loaded together. Both require a released `koment`
binary on the OpenCode process's `PATH`. The npm plugin starts
`koment mcp --write` for the pre-tool decision and runs `koment check`,
`koment comments check` and `koment agents check` at disposal. Failure to spawn
the binary prevents plugin initialization with an actionable error.

The npm package has no postinstall script. Requirements belong in the package
README and the plugin's startup error because those surfaces run regardless of
how OpenCode obtained the package.

Keep the dependency-free JSON-RPC client because it uses only `initialize`,
`notifications/initialized` and `tools/call`. It must terminate the connection
and reject every pending request on malformed JSON, an unexpected response id,
a stream error or child exit. Its `clientInfo.version` is read from the same
`package.json` that release-please updates. Node's built-in test runner covers
fragmented responses and every failure path; the required plugins CI job runs
those tests and packages each self-contained integration.

The release job publishes plugin archives and npm only after both canonical
binary and image jobs succeed, as ADRs 0109 and 0129 intended.

## Consequences

- Installation instructions match the actual OpenCode configuration contract.
- A missing binary fails at startup rather than creating a plugin that appears
  active while enforcing nothing.
- Protocol corruption produces an error instead of a hanging edit.
- The package adds no runtime dependency and its reported version cannot drift
  from the release version.
- The historical ADR remains available, including the assumptions corrected
  here; current documentation links to this replacement.

## Alternatives rejected

**Keep the nonexistent command as convenient shorthand.** A command in package
metadata is a user-facing claim. Shorthand that fails when copied is stale
documentation, not an abstraction.

**Keep a postinstall warning for direct npm users.** OpenCode owns package
installation and lifecycle-script execution is not a reliable integration
surface. Startup already has enough context to fail with the exact missing
binary and repository path.

**Let the plugin load without koment and enforce only what JavaScript can.**
This would either disable the policy silently or duplicate koment's classifier
in a second language. Both outcomes violate the reason the plugin exists.

**Add an MCP SDK to obtain protocol error handling.** The package uses three
methods and no resources, prompts or server-initiated requests. A small tested
state machine is less supply-chain surface. Reconsider an SDK when the plugin
needs a wider MCP capability.

**Continue after malformed or unsolicited server output.** There is no safe
partial result for a policy decision. Terminating the connection makes the
failure visible before an edit proceeds.
