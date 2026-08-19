# 0143 — Make the repository tree a closed contract

Date: 2026-08-19
Status: Accepted

## Context

The product core follows Go's conventional `cmd/` and `internal/` boundaries,
but the surrounding repository grew one surface at a time. `editors/` describes
the consumer, `plugins/` describes a packaging mechanism, `charts/` names one
deployment format, `packaging/` mixes package-manager templates with their
tests, and `workspace/` does not say that it is an example. A contributor has to
learn these exceptions before they can predict where the next file belongs.

The plugin tree also puts three independently installed agent integrations
inside one nominal Claude plugin. That obscures artifact boundaries and has
already made it possible for release packaging to select the manifest without
the commands, hooks, scripts and skills that make up the plugin. The current
tree is therefore not only aesthetic debt; it makes incomplete artifacts easier
to produce.

ADR 0138 defined four documentation sections but deliberately deferred moving
the existing pages. The repository now contains both those sections and the old
flat organization. A rule that applies only to new pages preserves the migration
as permanent debt and leaves two plausible homes for documentation.

The project treats stale documentation and workspace clutter as technical debt.
A diagram maintained by convention would become another claim that can drift.
The target tree needs one machine-readable authority, an exact human reference
derived from it and a required check that rejects paths outside it.

Some root paths cannot be normalized. GitHub Actions resolves `action.yml` from
the repository root, tools discover files such as `AGENTS.md`, `CLAUDE.md`,
`server.json` and their dot-directories at fixed locations, and ADR 0121 freezes
the published `schema/v1alpha` paths. Those constraints must be explicit
exceptions rather than precedents for arbitrary root files.

## Decision

The repository is organized by capability around an unchanged Go core:

```text
.
├── cmd/
├── internal/
├── schema/
├── integrations/
│   ├── agent-plugins/
│   │   ├── claude/
│   │   ├── hermes/
│   │   └── opencode/
│   └── editors/
│       ├── vscode/
│       └── zed/
├── distribution/
│   ├── helm/
│   │   └── koment/
│   └── package-managers/
│       ├── homebrew/
│       ├── scoop/
│       └── winget/
├── examples/
│   └── annotated-workspace/
├── docs/
│   ├── README.md
│   ├── start/
│   ├── guides/
│   │   ├── agents/
│   │   └── editors/
│   ├── reference/
│   └── explanation/
│       └── decisions/
├── scripts/
└── testdata/
```

`cmd/` contains binary entry points. `internal/` contains the private Go
product. This decision does not rename or regroup its packages merely to mirror
the outer tree. `schema/` remains at its existing public paths. `testdata/`
contains fixtures shared across product packages; package-specific fixtures
remain beside their package.

`integrations/` contains code installed into another product. An integration is
owned first by the kind of host and then by that host. Each agent plugin is a
self-contained installable root: the Claude directory owns its manifest,
commands, hooks, scripts and skills; Hermes and OpenCode own only their own
manifests and runtime files. Editor extensions follow the same rule.

`distribution/` contains the assets that deliver or deploy koment without
becoming part of another product. The Helm chart and its repository-level tests
move under `distribution/helm/`. Homebrew, Scoop and Winget templates and their
shared registry tests move under `distribution/package-managers/`.

`examples/` contains runnable or inspectable material that is neither a product
package nor a repository-wide test fixture. The maintained annotated workspace
moves there and is named for what it demonstrates.

The documentation migration from ADR 0138 is completed rather than left as a
compatibility layer. Getting-started sequences move to `start/`; task pages,
including agent and editor setup, move to `guides/`; exhaustive command, syntax,
policy and layout material moves to `reference/`; the data model, integration
architecture and ADRs move to `explanation/`. Pages with more than one job are
split. Every inbound repository link is updated in the same migration, and old
paths are removed rather than copied or redirected.

The repository root is a closed allow-list. It contains the top-level areas
above plus conventional project entry points, governance files and exact paths
required by external tools. The allowed root files and tool-owned dot-directories
are enumerated by name; an existing root entry is not automatically allowed.

The layout has one executable specification in `internal/projectlayout`. It
provides:

- the allowed root files and directories;
- the closed children of `integrations/`, `distribution/` and `docs/`;
- the purpose and ownership text for every architectural area;
- the legacy paths that must no longer exist; and
- validation of all paths returned by
  `git ls-files --cached --others --exclude-standard`.

The same package generates `docs/reference/repository-layout.md`. A generation
check fails when the checked-in reference differs. `mise run layout-check`, the
pre-commit hooks and the required `ci` aggregate run both validations. Ignored
compiler output and dependency caches are outside the path check because their
presence is governed by `.gitignore`; a non-ignored scratch file is repository
clutter and fails.

A later layout change must supersede this ADR. The replacement must name a
capability that cannot truthfully live in any existing area, explain why every
plausible existing owner was rejected, update `DESIGN.md` and the executable
specification, regenerate the reference and migrate all affected paths and
links in the same change. Preference, a growing file count, a new implementation
language or a desire for symmetry is not sufficient justification.

The migration lands as one pull request with reviewable commits grouped by
capability. No commit that is intended to merge may leave imports, release
inputs, documentation links or annotations pointing at an old path. The drift
gate becomes required only after the complete target tree exists.

## Consequences

- A reader can predict ownership from the path, and each independently shipped
  integration has a self-contained source directory.
- The deferred ADR 0138 migration is paid down instead of supported forever.
- A new root area or architectural category is a deliberate design change that
  cannot pass CI through an undocumented directory addition.
- The human layout reference cannot silently diverge from the rule CI applies.
- External fixed paths remain visually noisy at the root, but each one is an
  enumerated compatibility constraint.
- Moving ADRs, examples, charts, package templates and integrations creates a
  large mechanical diff and invalidates annotation paths. The migration must
  re-anchor or move every affected annotation and verify all release inputs.
- Existing links to repository files outside the repository may break. The
  project does not preserve internal source paths as a public API; published
  artifact names, schema URLs and consumer installation contracts do not move.

## Alternatives rejected

**Keep the existing top-level directories and only finish the documentation
migration.** This is the smallest change, but it leaves integrations divided by
inconsistent concepts and keeps the artifact-boundary defect that motivated the
review.

**Organize by implementation language or build tool.** Roots such as `go/`,
`rust/` and `node/` make tool choice more visible than product ownership. A Zed
extension would move if its SDK changed languages even though its purpose did
not.

**Put every external surface under one `plugins/` directory.** Helm charts and
package-manager templates are distribution mechanisms, not host plugins. The
name also fails to distinguish an editor extension from an agent plugin and
recreates the mixed artifact boundary.

**Regroup the Go product into feature directories at the same time.** The
existing `cmd/` and `internal/` packages already express Go visibility and
product responsibilities. Moving them would add import churn without resolving
the surrounding ownership problem and would mix a code architecture change into
a repository navigation change.

**Describe the target tree in prose and rely on review.** This is the mechanism
that allowed the documentation migration and historical roots to coexist. It
cannot distinguish a deliberate exception from unreviewed clutter after the
fact.

**Keep compatibility copies, symlinks or redirect stubs at old paths.** GitHub
source paths are not published product interfaces, and duplicate locations
would make ownership ambiguous. The migration updates repository-controlled
references atomically; immutable release artifacts and versioned schema paths
remain unchanged.

**Allow the layout checker to accept any new directory accompanied by an ADR.**
That proves a document exists, not that the current architecture could not own
the capability. Requiring a superseding decision and explicit rejected
placements sets the intended higher bar.
