# 0139 — Package a Zed extension, and widen what earns a package

Date: 2026-08-19
Status: Accepted

## Context

ADR 0112 set the bar for a packaged editor: it gets one only when it has a
marketplace **and** offers APIs that LSP cannot reach. VS Code cleared both. The
ADR also said adding another packaged editor is a decision recorded in its own
ADR rather than a directory someone adds. This is that ADR, and it does not
clear the bar as written.

Read in the `zed_extension_api` 0.7.0 source and the WIT definitions rather than
assumed:

- Extensions are Rust compiled to `wasm32-wasip2`. The capabilities are
  languages, language servers, context servers, themes, icon themes, snippets
  and debuggers.
- **There is no decoration or inline-rendering API.** The extension host exposes
  `worktree.which`, `worktree.shell_env`, settings and a `Command` record. So a
  koment extension can start `koment lsp` and register `koment mcp --write`, and
  it cannot render an annotation body beside its line — the one thing ADR 0112
  named as justifying a package.
- A language server **can** attach to languages the extension does not itself
  define, through an explicit list. Verified against `zed-extensions/harper`,
  a published extension whose `extension.toml` names about thirty-five
  languages including Zed built-ins. There is no wildcard.
- Publication is a pull request to `zed-industries/extensions` adding this
  repository as a git submodule and an entry in the top-level `extensions.toml`.
  That entry accepts a `path` field for an extension inside a subdirectory.

Meanwhile `docs/agents/zed.md` documents the current experience: hand-edit
`settings.json` to register the MCP server, and note that "a GUI-launched editor
does not inherit your shell's `PATH`". It does not document the language server
at all, so a Zed user gets no hover, no lenses and no diagnostics unless they
work that out themselves.

Zed is also where the competition is. Zed ships DeltaDB, which records the
agent conversation behind each change — a different data model from koment's
anchored, drift-checked claim, but the same question and the same screen.

## Decision

**Widen ADR 0112's test.** A packaged editor needs a marketplace and *either*
APIs that LSP cannot reach *or* a distribution channel that removes installation
steps the user would otherwise perform by hand. Zed clears the second: the
extension replaces two settings edits and makes the binary discoverable in the
one place a GUI-launched editor will look.

ADR 0112 stays Accepted. Its packaged/configured split, its one-version rule and
its publication ordering are unchanged; only the entry test is amended.

**The extension lives in `editors/zed/`**, published by pointing Zed's
`extensions.toml` at this repository with `path = "editors/zed"`. One repository,
one version, one release — the property ADR 0112 exists to protect.

**Binary resolution** is: the user's `lsp.koment.binary.path` setting, then
`worktree.which("koment")`, then a message naming the `$PATH` problem and the
setting that fixes it. The resolved path is cached and reused for the context
server, which Zed hands a `Project` rather than a `Worktree` and therefore
cannot resolve a binary itself.

**The crate version is `0.0.0` and stays there.** `extension.toml` carries the
repository version through `release-please` markers, exactly like the Helm chart
and the plugin manifests. The crate is never published, so its version has no
consumer, and pinning it flat keeps `Cargo.lock` out of the release diff — cargo
rewrites that file and would strip any marker comment placed in it.

**Nothing is attached to the GitHub release.** Zed's registry builds the
extension from the submodule commit, so an attached `.wasm` would be a second
artifact that nobody consumes and that could differ from what Zed ships. This is
a deliberate exception to ADR 0109's "the release is canonical" for this one
channel, because for Zed the source at a tag *is* the artifact.

**Publication stays manual.** The `zed-industries/extensions` pull request is a
step in `docs/guides/release-koment.md`, taken by a person. AGENTS.md §14
already requires human approval for anything that publishes.

## Consequences

- Zed users install one extension instead of editing settings twice, and get the
  language server they were never told about.
- The repository now contains Rust, a second toolchain to pin, install in CI and
  keep current. `editors/zed/mise.toml` pins it and the `zed` CI job builds the
  wasm on every pull request.
- `zed_extension_api` is a new dependency in a new language. It is first-party to
  the editor it targets, has no transitive footprint in the shipped artifact
  beyond the generated bindings, and is unavoidable — there is no other way to
  write a Zed extension. AGENTS.md §10 is satisfied by this paragraph.
- **The language list is enumerated and will go stale.** koment's anchoring is
  language-agnostic; Zed's `languages` key is not. A language Zed gains, or one
  provided by an extension the user installs, gets nothing from koment until
  someone adds a line and cuts a release. This is the honest cost of the
  channel and there is no configuration that avoids it.
- Every release now has a manual step in another organisation's repository, and
  a release is not fully published until that pull request merges.
- Zed users get less than VS Code users: no inline body beside the line, no
  annotation panel. The extension README says so rather than implying parity.

## Alternatives rejected

**A separate `koment-dev/zed-koment` repository.** Keeps Rust out of a Go
monorepo and gives Zed its own cadence. Rejected because it breaks ADR 0112's
rule that one version spans the binary, the LSP and every editor package: two
repositories means a compatibility matrix between a thin client and the server
it speaks to, which is the thing that rule exists to prevent. Zed's `path` field
makes the monorepo work, so the cost buys nothing.

**Documentation only — add a `zed` adapter to `koment agents install` that
writes `.zed/settings.json`.** Cheapest, and leaves ADR 0112 untouched. Rejected
because it does not solve the problem it would claim to: the binary still has to
be found by a process that did not inherit the user's `$PATH`, and koment stays
absent from the marketplace Zed users actually browse. Worth doing *as well*, and
it is not in this change.

**Bundle the koment binary in the extension**, as the VS Code package does
(ADR 0113). Not possible: Zed extensions are sandboxed WebAssembly with no way to
ship or mark-executable a platform binary for the host. Downloading one from the
GitHub release inside `language_server_command` is possible and was rejected for
now — it would make koment's version depend on when the extension last ran rather
than on what the user installed, and silently diverge from the CLI in the same
repository.

**Attach the built `.wasm` to the GitHub release.** Follows ADR 0109's shape and
gives an install path that does not depend on Zed's registry. Rejected because
Zed builds from source regardless, so the attached file would be an artifact with
no consumer and a second thing that can disagree with what users run.

**Sync the crate version with the release.** What `zed-extensions/harper` does.
Rejected because `Cargo.lock` also carries it, cargo rewrites that file, and
`release-please` needs marker comments that cargo would strip — so the lock would
drift from the manifest on the first build after every release.
