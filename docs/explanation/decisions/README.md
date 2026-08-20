# Architecture decisions

This directory contains the active decisions for the approved design. They use
the 0100 series to keep their identities distinct from the pre-deployment
prototype decisions.

On 2026-08-03, before any live deployment or durable served database existed,
the project deliberately returned to design stage. The earlier 0001–0026 set
mixed implemented behaviour, future intent, demo choices and dependency notes
until it was no longer an honest description of the product. It was removed
with explicit project-owner approval rather than carried forward as 26 active
constraints.

The old documents remain available in Git:

```text
git ls-tree -r v0.2.0 docs/decisions
git show v0.2.0:docs/decisions/0022-index-annotations-in-a-database.md
```

`DESIGN.md` marks implemented and planned boundaries explicitly. Until a planned
stage lands, the code describes current behaviour and the active ADR describes
the approved destination.

## Active decisions

- [0100 — Keep one authoritative Git record per annotation](0100-one-git-record-per-annotation.md)
- [0101 — Fail deterministic resolution when an anchor is ambiguous](0101-fail-ambiguous-anchor-resolution.md)
- [0102 — Build every read surface from one repository snapshot](0102-one-repository-snapshot-for-every-reader.md)
- [0103 — Give humans and agents explicit capabilities in three tiers](0103-three-tiers-with-human-and-agent-capabilities.md)
- [0104 — Serve assigned repositories as atomic commit snapshots](0104-atomic-multi-repository-snapshots.md)
- [0105 — Authenticate remote access and materialize writes through Git](0105-authenticated-writes-materialize-through-git.md)
- [0106 — Use konflate as the operational baseline](0106-konflate-is-the-operational-baseline.md)
- [0107 — Enforce the comment-free thesis on koment itself](0107-dogfood-the-comment-free-thesis.md)
- [0109 — Distribute authenticated artifacts instead of Go invocations](0109-distribute-authenticated-artifacts-instead-of-go-invocations.md)
- [0110 — Keep editor semantics behind LSP](0110-keep-editor-semantics-behind-lsp.md)
- [0111 — Ship Windows without letting it gate the pipeline](0111-windows-is-supported-but-not-gating.md)
- [0112 — Publish one editor package per marketplace, and configuration everywhere else](0112-publish-one-editor-package-per-marketplace.md)
- [0113 — Bundle the released binary in the extension](0113-bundle-the-release-binary-in-the-extension.md)
- [0114 — Show rationale in a panel, and stop diagnosing what is healthy](0114-show-rationale-in-a-panel-and-stop-diagnosing-health.md)
- [0115 — Give an annotation a title](0115-an-annotation-has-a-title.md)
- [0116 — Stop reporting a stale line number as a status](0116-moved-is-not-a-status.md)
- [0117 — Relicense to AGPL-3.0-or-later with commercial dual licensing](0117-relicense-to-agpl-with-commercial-dual-licensing.md)
- [0118 — Translate opencode edits into apply_patch so one comment walker serves every client](0118-translate-opencode-edits-into-apply-patch.md)
- [0119 — Make the annotation a Kubernetes-shaped resource, cut off v1 records, and freeze the API](0119-make-the-annotation-a-kubernetes-shaped-resource.md)
- [0120 — Promote koment to v1.0.0](0120-promote-koment-to-v1-0-0.md)
- [0121 — Every committed koment file is a resource, and its schema is pinned to the API version](0121-every-committed-koment-file-is-a-resource-with-a-pinned-schema.md)
- [0122 — Run Renovate on GitHub runners behind an app we own](0122-run-renovate-on-github-runners-behind-our-own-app.md)
- [0123 — A published path never begins with a dot](0123-a-published-path-never-begins-with-a-dot.md)
- [0124 — `koment bootstrap`, a stronger agent contract, and a user-configurable comment allow-list](0124-koment-bootstrap-and-allowed-comment-patterns.md)
- [0125 — Decouple the `ci` aggregate from the `setup-action` smoke](0125-decouple-ci-aggregate-from-setup-action.md)
- [0127 — Change the VS Code extension `displayName` to `koment-dev`](0127-fix-vscode-marketplace-display-name.md)
- [0128 — Enforce Conventional Commits 1.0.0 subjects](0128-enforce-conventional-commit-names.md)
- [0130 — Delete the v1 auto-migrate path and refuse `version: 1`](0130-delete-the-v1-auto-migrate-path.md)
- [0131 — One registry drives the command help](0131-one-registry-drives-the-command-help.md)
- [0132 — Comment detection covers every file, with an opinionated catch-all](0132-comment-detection-covers-every-file.md)
- [0133 — An annotation can be edited and forgotten](0133-an-annotation-can-be-edited-and-forgotten.md)
- [0134 — The demo workspace shows every state koment can produce](0134-the-demo-workspace-shows-every-state.md)
- [0135 — Floating major aliases for consumers, SHA pins for ourselves](0135-floating-major-aliases-for-consumers.md)
- [0136 — Documentation stays in the repository and is served from elsewhere](0136-docs-stay-in-the-repository-and-are-served-elsewhere.md)
- [0137 — A feature is not done until its documentation is true](0137-a-feature-is-not-done-until-its-documentation-is-true.md)
- [0138 — Documentation has four sections, and every page belongs to one](0138-documentation-has-four-sections.md)
- [0139 — Package a Zed extension, and widen what earns a package](0139-package-a-zed-extension.md)
- [0140 — Generate the language reference from the syntax table](0140-generate-the-language-reference.md)
- [0141 — Slash commands are the human surface, the skill is the agent's](0141-slash-commands-are-the-human-surface.md)
- [0142 — The note column carries its own height, and moves to the line on a phone](0142-the-reading-view-carries-its-own-height.md)
- [0143 — Make the repository tree a closed contract](0143-make-the-repository-tree-a-closed-contract.md)
- [0144 — Configure the OpenCode plugin and fail closed](0144-configure-opencode-plugin-and-fail-closed.md)
- [0145 — License only the Zed extension under GPLv3](0145-license-the-zed-extension-under-gplv3.md)
- [0147 — Upload release assets through the creation response](0147-upload-release-assets-through-the-creation-response.md)

## Superseded decisions

- [0108 — Layer agent guidance behind an authoritative policy gate](0108-layer-agent-guidance-behind-an-authoritative-policy-gate.md) — superseded by 0124
- [0126 — Fix the VS Code extension marketplace name to `koment-dev`](0126-fix-vscode-marketplace-extension-name.md) — superseded by 0127
- [0129 — OpenCode plugin as a first-class installable](0129-opencode-plugin-distribution-surface.md) — superseded by 0144
- [0146 — Gate distribution on visible release assets](0146-gate-distribution-on-visible-release-assets.md) — superseded by 0147

Use [0000-template.md](0000-template.md) for a new decision. A new dependency or
a structural change still requires its own ADR when these decisions do not
already settle it.
