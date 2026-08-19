# 0136 — Documentation stays in the repository and is served from elsewhere

Date: 2026-08-10
Status: Accepted

## Context

`docs/` holds ten Markdown files that are the only real explanation of how to
use koment. They are reachable today only by browsing GitHub, which is fine for
a contributor and poor for someone deciding whether to adopt the tool: no
search, no navigation, no landing, and a reading experience that signals
"internal notes" rather than "product".

The obvious answer — GitHub Pages on this repository — is not available. **A
GitHub Pages site accepts one custom domain, and both of ours are spent:**

```
koment-dev/koment      → why.koment.dev   (the annotations demo, rendered by koment site)
koment-dev/koment.dev  → koment.dev       (the landing page)
```

Serving docs from `koment-dev/koment` would mean giving up `why.koment.dev`, or
nesting the documentation under it at `why.koment.dev/docs/`, which puts the
manual inside a demo.

The other constraint is where the source lives. Documentation that drifts from
the code is the failure this whole project exists to name, and the surest way to
cause it is to move the prose away from the change that invalidates it. The
files must stay in `docs/` in this repository, edited in the same pull request
as the behaviour they describe.

## Decision

**Source stays in `docs/`. Serving moves off GitHub Pages.**

`docs.koment.dev` is built from `docs/` in this repository by a static site
generator and published to a host that is not GitHub Pages — Cloudflare Pages
is the intended one, being free, custom-domain-capable, and able to build
directly from the repository without a second copy of the content.

Three properties are non-negotiable, and any host that has them is acceptable:

1. **The Markdown in `docs/` is the only source.** No parallel wiki, no
   exported copy, no CMS. A file is edited in the pull request that changes the
   behaviour it documents.
2. **The build is reproducible from the repository**, so a broken docs build is
   a red check on the pull request rather than a surprise after merge.
3. **The generator adds no runtime dependency to koment itself.** It is build
   tooling for a website; it never enters `go.mod` and it never becomes
   something a user of the CLI has to install.

`why.koment.dev` keeps serving the annotations demo, and `koment.dev` keeps
serving the landing page. The three subdomains address three different things:
the product, the manual, and the tool's own reasoning rendered by itself.

The README links the site rather than the individual files, so there is one
place to keep current.

## Consequences

What becomes easier:

- Documentation gains search and navigation, which ten files already need and
  twenty will need badly.
- Evaluating koment stops requiring a repository browse.
- The docs build fails on a pull request that breaks a link or a reference,
  which is a check the repository does not have today.

What becomes harder:

- A second hosting provider joins the surface. Cloudflare Pages is not GitHub,
  so an outage, an account change or a billing lapse is a new way for the manual
  to disappear. The mitigation is that nothing depends on it: `docs/` is still
  readable in the repository, and the site is a rendering of it.
- A static site generator is a new build dependency with its own upgrade
  treadmill. It is confined to the docs job and never reaches the binary.
- Three subdomains need explaining to a newcomer. The README does that in one
  line.

## Alternatives rejected

- **GitHub Pages on this repository.** The natural choice and the one everybody
  reaches for. Rejected because it would cost `why.koment.dev`: one Pages site,
  one custom domain, and the annotations demo is already there. Nesting the
  manual under the demo inverts which one is the product.

- **A third repository holding the documentation.** Restores GitHub Pages and
  its custom domain. Rejected on the ground that matters most: the prose would
  no longer live beside the code it describes, so the pull request that changes
  behaviour would not be the pull request that updates the manual. That is the
  precise mechanism by which documentation rots, and this project has no
  standing to ship it.

- **GitHub Wiki.** Free, hosted, no build. Rejected for the same reason plus a
  worse one: a wiki is a separate git repository that cannot take a custom
  domain, is not reviewed in pull requests, and is edited outside the change
  that motivates it.

- **A documentation SaaS (GitBook, Mintlify, Read the Docs).** Better search
  and navigation than a generic generator. Rejected as disproportionate for ten
  Markdown files, and because several of them want to own the content rather
  than render a directory.

- **Teach `koment site` to render `docs/`.** koment already publishes static
  HTML, so this looks free. Rejected under scope discipline: `koment site`
  renders annotations anchored to source, which is a different thing from a
  manual, and widening it means maintaining a general documentation generator
  inside a tool whose deterministic core is the thing that must stay trustworthy.
