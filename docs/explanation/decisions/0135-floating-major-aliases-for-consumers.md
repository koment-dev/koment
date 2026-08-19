# 0135 — Floating major aliases for consumers, SHA pins for ourselves

Date: 2026-08-09
Status: Accepted

## Context

`docs/publishing.md` carries the copy-paste workflow that is the front door to
the published tier — the thing the README and the docs point a new user at. It
pinned `koment-dev/koment@v0.2.0`, three majors behind, alongside
`actions/checkout@v5`, `configure-pages@v5`, `upload-pages-artifact@v4` and
`deploy-pages@v4`, all superseded. A newcomer copying it got a workflow that
installed a year-old koment and produced a site missing every feature added
since.

This is documentation rot with a mechanical cause: an exact pin in an example
is guaranteed to age, and nothing fails when it does. No test covers a fenced
code block, and the person best placed to notice — someone evaluating koment
for the first time — is the person least equipped to know it is stale.

The obvious fix, `uses: koment-dev/koment@v3`, did not resolve: koment
published `v0.7.0` through `v2.2.0` and no moving major tag at all. The
container image had the same gap, tagging `{{version}}`, `{{major}}.{{minor}}`
and `latest`, so `ghcr.io/koment-dev/koment:3` did not exist either.

## Decision

koment maintains a **floating major alias on every channel that can carry one**,
and continues to pin itself to exact SHAs.

- **Git tag `vN`.** Re-pointed at each `N.x.y` release, so `koment-dev/koment@v3`
  resolves and keeps resolving. Moved by a release job that runs `needs:
  [please, verify]`, so the alias only advances once the published release has
  been verified on Linux and macOS. A version that is not a plain `X.Y.Z` — a
  release candidate — leaves the alias alone.
- **Container tag `{{major}}`.** `ghcr.io/koment-dev/koment:3` joins `3.1`,
  `3.1.0` and `latest`.
- **The action's `version` input already floats.** It defaults to `latest` and
  installs the newest koment release at run time, independently of which action
  ref the user pinned. The two dials are separate and the documentation now
  says so, because conflating them is the obvious misreading.

Channels that cannot carry an alias are left alone rather than approximated:
npm and Helm consumers express the same intent client-side with a semver range
(`^3.0.0`), and the VS Code Marketplace, Open VSX and Homebrew serve the newest
version by construction.

**This does not change how koment pins its own workflows.** Every `uses:` in
`.github/workflows/` stays a full commit SHA with Renovate maintaining it. The
asymmetry is deliberate and stated in the docs: a moving tag is a supply-chain
trust decision, and whoever can move the tag can change what runs in a
consumer's pipeline. Auto-update is the right default for evaluating a tool and
the wrong default for a repository that publishes signed artifacts.

### On `AGENTS.md` §14

§14 says: *"Never delete or re-point a tag to redo a release."* This decision
re-points a tag on every release, so the tension is worth naming rather than
leaving for someone to discover.

The rule targets **release tags** — the immutable coordinates a published
artifact is fetched by. `v3.0.0` remains permanent and is never moved, deleted
or reused; that guarantee is untouched. `v3` is not a release tag and never
names a release: it is an alias whose entire purpose is to move, in the same
way `actions/checkout@v5` does. Nothing is republished when it moves and no
version number is reused.

The rule stands unchanged. This ADR records that aliases are outside its scope.

## Consequences

What becomes easier:

- The documented workflow keeps working without edits, which is the only way an
  example in prose stays honest.
- A user gets patches by doing nothing, and cannot receive a breaking change
  without editing the major in their own file.
- Kubernetes and Compose users get the same guarantee from the image tag.

What becomes harder:

- A moving tag is mutable infrastructure. If a `3.x` release is bad, everyone on
  `@v3` gets it on their next run. The mitigation is the ordering — the alias
  moves only after `verify` passes — and the remedy is the same as for any bad
  release: cut the next patch, which moves the alias forward again.
- There is now a supply-chain claim koment makes to its users that it does not
  make to itself. The documentation states the asymmetry and the reason, but a
  reader who skims will not see it.
- Someone will eventually read the release workflow, see `git push --force` on
  a tag, and reach for §14. Hence the section above.

## Alternatives rejected

- **Leave the example pinned exactly and update it by hand.** No moving tags,
  no new trust surface. Rejected on evidence: this is what was in place, and it
  rotted by three majors across five separate pins without anyone noticing.

- **Point the example at `@main`.** Always current, no tag machinery. Rejected
  because it offers no protection at all — a breaking change lands in a
  consumer's pipeline the moment it merges, before it is even released.

- **Rely on the action's `version: latest` default and pin `uses:` exactly.**
  Tempting, and it does float the CLI. Rejected because the action wrapper
  itself still rots: input names, defaults and runner support change, and the
  user is left on whatever the wrapper looked like the day they copied it.

- **Maintain `vN.M` aliases as well.** Finer-grained pinning for cautious
  users. Rejected as unearned surface: nobody has asked for it, every extra
  alias is another mutable ref to keep correct, and a semver range already
  expresses it wherever a package manager is involved.
