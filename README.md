<div align="center">

<img src="internal/ui/assets/koment-logo.svg" alt="koment comment bubble" width="104">

# koment

**Keep the _why_ next to your code — checked, so it can't quietly rot.**

[![Release](https://img.shields.io/github/v/release/koment-dev/koment?label=release)](https://github.com/koment-dev/koment/releases/latest)
[![CI](https://img.shields.io/github/actions/workflow/status/koment-dev/koment/ci.yml?branch=main&label=ci)](https://github.com/koment-dev/koment/actions/workflows/ci.yml)
[![OpenSSF Scorecard](https://img.shields.io/ossf-scorecard/github.com/koment-dev/koment?label=openssf+scorecard)](https://scorecard.dev/viewer/?uri=github.com/koment-dev/koment)
[![Open VSX](https://img.shields.io/open-vsx/v/koment/koment-dev?label=open%20vsx)](https://open-vsx.org/extension/koment/koment-dev)
[![License](https://img.shields.io/github/license/koment-dev/koment)](https://github.com/koment-dev/koment/blob/main/LICENSE)
[![Annotations](https://img.shields.io/badge/annotations-browse-brightgreen)](https://why.koment.dev/)

</div>

## The problem

Someone wrote `retry 5 times` instead of 3, and there was a good reason. Six
months later nobody remembers it, so it gets "simplified" back to 3 and the bug
returns.

A comment was supposed to prevent that. Comments rot: they sit next to the code
without being attached to it, and when the code changes nothing tells you the
comment just became a lie. Your AI agent then reads that lie and acts on it.

## What koment does

You record the reason **outside** the source, anchored to the exact lines it
explains. Then koment watches those lines. Change them, and it says so:

```console
$ koment check
internal/auth/token.go
  drifted   gotcha        internal/auth/token.go  01KZN63VC5SZASDYJBMPDC03WB
    Without it, clients whose clock runs fast get logged out mid-request. Bit
    us in #412.
1 annotation across 1 file: 1 drifted
koment: 1 annotation no longer resolves; revisit it or update the anchor
```

Exit code 1. The build stops. Nobody silently inherits a reason that no longer
describes the code.

The records are YAML in `.koment/`, in your git repository. No database, no
service, no account. They review in the same pull request as the change that
motivated them.

**[See it running →](https://why.koment.dev/)** — koment's own annotations,
published by [a workflow you can copy](docs/guides/publish-annotations.md).

## For AI agents

This is where it earns its keep. Agents write explanatory comments constantly,
and they edit code without reading why it was written that way. koment makes
both of those fail *immediately*, while the agent is still working:

- **It writes a comment** → the edit is refused, with instructions to record the
  reasoning as an annotation instead.
- **It edits annotated code without checking** → the turn cannot finish until
  the annotation is revisited.

One MCP server serves Claude Code, Cursor, Codex, opencode, Hermes, Zed and the
rest, so every agent reads the same reasoning through the same interface.
[Set yours up →](docs/guides/agents/)

Humans get the same loop in the editor: a squiggle under the comment, and a
quick fix offering *convert to annotation* or *keep it, on the record*.

## Install

```bash
brew install koment-dev/tap/koment
```

<details>
<summary>Other ways</summary>

```bash
mise use -g github:koment-dev/koment          # mise
go install github.com/koment-dev/koment/cmd/koment@latest
docker run --rm -v "$PWD:/repo" -w /repo ghcr.io/koment-dev/koment:3 check
```

In GitHub Actions, `uses: koment-dev/koment@v3`. Or take a checksum-listed
binary for Linux, macOS or Windows on amd64/arm64 from the
[latest release](https://github.com/koment-dev/koment/releases/latest) — every
other channel is built from those same artifacts.

</details>

## Try it in a minute

From inside any git repository:

```bash
koment bootstrap
```

That sets up `.koment/` and wires whichever agents you use. Now record a reason
and watch it get checked:

```bash
koment add src/auth.go \
  --excerpt 'if token.Expiry.Before(now.Add(-clockSkew)) {' \
  --kind gotcha \
  --title 'Skew subtraction keeps fast clocks logged in' \
  --body 'Without it, clients whose clock runs fast get logged out
          mid-request. Bit us in #412.'

koment check     # ok
```

**Now edit that line and run `koment check` again.** It fails, and tells you
which reasoning no longer matches its code. That is the entire product; the rest
is delivering it to the people and agents who need it.

```bash
koment ui        # read them in a browser
koment show src/auth.go
```

## How it works

1. You annotate a snippet. koment records the prose, exact excerpt, surrounding
   source context, the commit you were on, and who you are.
2. The record lands in `.koment/annotations/<id>.yaml`, one file per
   annotation. Concurrent agents create independent files instead of replacing
   one shared list.
3. Resolution searches the current file for that excerpt and produces exactly
   one status:

   | | meaning | build |
   |---|---|---|
   | `ok` | found where it was last seen | passes |
   | `ambiguous` | several contextual candidates remain | **fails** |
   | `drifted` | file exists, the annotated code is gone | **fails** |
   | `orphaned` | the file is gone | **fails** |

4. `koment check` exits non-zero on `ambiguous`, `drifted` or `orphaned`. That
   is the whole mechanism: uncertain rationale is worse than no annotation, so
   it has to be impossible to ignore.
5. When it fails, `koment reanchor <id> --excerpt '<new text>'` repoints it —
   keeping its id and creation date, recapturing context and the line for you.
   Nothing re-attaches automatically; a person confirms the reasoning still holds.

Anchoring is by **excerpt**, never by line number — line numbers rot on the next
edit above them. The commit hash *is* recorded, but only to reconstruct history;
it never decides whether an annotation still applies. Those are two different
questions with two different mechanisms.

Local writes change the checkout you are already reviewing. Served writes never
touch a replica or push a default branch: they create an exact annotation on a
deterministic branch and return only after its pull request exists. Static
publications remain immutable and read-only.

## Three ways to run it

Pick one. Each is a place to stop, not a step you have to take, and **moving
between them is not a migration** — all three read the same `.koment/` in git,
so there is nothing to export, import or back up.

| | you run | you get |
|---|---|---|
| **local** | the CLI, `koment ui --write`, and `koment mcp --write` | humans and agents read and write the same checked records. Nothing to host. |
| **published** | [one workflow file](docs/guides/publish-annotations.md) → GitHub Pages | everyone reads the annotations in a browser. No server, no auth to design, no cost. |
| **served** | the container or the [Helm chart](#kubernetes) | authenticated, commit-stamped GitHub snapshots for several repositories, cross-repository search, reviewed annotation PRs, metrics |

## Several repositories

One deployment serves many. Identity is assigned independently of provider path,
and every refresh resolves a branch to one immutable commit before replacing a
repository's active snapshot:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/koment-dev/koment/main/schema/server.schema.json
repositories:
  - id: payments
    name: Payments API
    provider: github
    remote: you/payments
    default_branch: main
    default: true
  - id: web
    name: Customer Web
    provider: github
    remote: you/web
    default_branch: main
```

The service starts only when a non-loopback listener has either trusted-proxy
identity or scoped bearer credentials. Private repositories and reviewed writes
also require a GitHub token. The [Helm chart documentation](distribution/helm/koment/README.md)
shows the secret formats and boundary.

Local commands need none of this configuration — koment finds the owning
checkout by walking up from the working directory.

## Kubernetes

koment publishes an **OCI** Helm chart to `oci://ghcr.io/koment-dev/charts/koment`:

```bash
helm install koment oci://ghcr.io/koment-dev/charts/koment \
  --set repositories[0].remote=you/your-repo \
  --set metrics.enabled=true \
  --set metrics.serviceMonitor.enabled=true \
  --set metrics.dashboard.enabled=true
```

The application port authenticates source, rationale, UI and MCP; only liveness
and readiness are public. Metrics use a separate listener so an ingress cannot
accidentally expose them with application authentication. The chart includes a
Grafana dashboard, ServiceMonitor, hardened pod defaults, optional NetworkPolicy
and disruption controls, and a digest-pinned `helm test` probe.

## Configuration

Every flag can be set from the environment. `--flag-name` becomes
`KOMENT_FLAG_NAME`, and an explicit flag always wins.

| | |
|---|---|
| `KOMENT_CONFIG` | strict served repository YAML |
| `KOMENT_CREDENTIALS_FILE` | secret file of SHA-256 bearer hashes and repository scopes |
| `KOMENT_GITHUB_TOKEN_FILE` | provider token file for private reads and reviewed writes |
| `KOMENT_LISTEN` | local UI or unified service address |
| `KOMENT_HUMAN_WRITES` | allow identities from the trusted OIDC proxy to create reviewed annotations |
| `KOMENT_TRUSTED_PROXIES` | CIDRs allowed to assert forwarded human identity |
| `KOMENT_SYNC_INTERVAL` | provider snapshot refresh interval |
| `KOMENT_METRICS` | separate metrics listener; off unless set |
| `KOMENT_WRITE` | enable local UI or stdio MCP mutations |
| `KOMENT_OUT` | static publication output directory |

Git is the only authoritative record. Local reads resolve the YAML directly
against the working tree; disposable read models cannot restore or overwrite
Git.

`koment <command> --help` lists every flag alongside its variable.

## Work where the code lives

The reference [VS Code extension](integrations/editors/vscode/README.md) starts `koment lsp`,
renders annotation bodies as virtual inline text, reports drift and prohibited
comments as diagnostics, and adds native add, reanchor, convert and explicit
acknowledgement actions. The prose is never inserted into the source buffer.
Every release attaches a signed VSIX and publishes the same artifact to the VS
Code Marketplace and Open VSX.

Other editors can use the standard hover, diagnostics, code-lens, code-action
and execute-command surface from `koment lsp` without reimplementing storage or
anchoring.

Claude Code can install the repository's own project-scoped marketplace plugin:

```text
/plugin marketplace add koment-dev/koment
/plugin install koment@koment-dev
```

It bundles writable MCP configuration, injects the strict procedure at session
start and runs the policy gate before Claude can finish a turn. Install a
released `koment` binary and run `koment agents install` in the repository first.

OpenCode ships a parallel plugin at
[`integrations/agent-plugins/opencode/`](integrations/agent-plugins/opencode/).
Add `@koment/opencode-koment` to the `plugin` array in `opencode.json`; OpenCode
installs it at startup. It requires the released `koment` binary on `PATH`,
denies ordinary explanatory comment intent and runs the policy gate on session
end. ADR 0144 records the package and generated-adapter boundary.

## Give it to your agents

```bash
koment agents install
koment mcp --write
```

The generated repository adapters give agents a strict procedure and configure
the writable stdio server. It has three read tools:

- **`koment_get(file, repository?)`** — annotations for the file an agent is about to edit
- **`koment_search(query, repository?)`** — find reasoning by topic; omitting `repository` searches all of them
- **`koment_repositories()`** — what this deployment serves, with counts

Write mode adds four tools:

- **`koment_add`** — create agent-attributed rationale
- **`koment_reanchor`** — explicitly move an existing anchor
- **`koment_convert_comment`** — record a comment as rationale, then remove it
- **`koment_acknowledge_comment`** — retain an exceptional comment only after an explicit acknowledgement

Those four mutation tools operate on a local checkout. The authenticated served
MCP surface exposes `koment_add` and returns the deterministic branch, commit and
pull-request URL; source-mutating conversion and reanchor stay local so the
service cannot overwrite an agent's separate worktree.

Every annotation arrives with its resolution status *and* its repository, so a
stale one is never presented as current and a result is never detached from its
scope. When a path exists in several repositories `koment_get` **refuses and
names the candidates** rather than guessing.

| | | |
|---|---|---|
| [Claude Code](docs/guides/agents/claude-code.md) | [Cursor](docs/guides/agents/cursor.md) | [VS Code](docs/guides/agents/vscode.md) |
| [Zed](docs/guides/agents/zed.md) | [Codex](docs/guides/agents/codex.md) | [opencode](docs/guides/agents/opencode.md) |
| [Hermes](docs/guides/agents/hermes.md) | [OpenClaw](docs/guides/agents/openclaw.md) | [Anything else](docs/guides/agents/other.md) |

## Publish it

Your reviewers are not going to install anything. Give them a URL: one workflow
file renders every annotation to static HTML plus `annotations.json` and
`search.json`, then puts the atomic snapshot on GitHub Pages with no server,
database or authentication to design.

```yaml
- uses: actions/checkout@v7
- uses: koment-dev/koment@v3
- run: koment check
- run: koment site --out dist
```

Every page names the commit it was rendered from, so a snapshot can never pass
for the current tree. `koment check` in the same workflow means a build that
would publish drift fails first.

**[The whole workflow, ready to copy →](docs/guides/publish-annotations.md)**

A site renders your source as well as your annotations, so publishing one from a
private repository publishes that source. Grouped publications render one
snapshot per repository and connect them through the ordinary repository
switcher; there is no selector landing page.

## Documentation

- **[Publishing](docs/guides/publish-annotations.md)** — the copy-paste workflow, and moving to a served instance later
- **[Contributor setup](docs/start/contributing.md)** — establish a verified development checkout
- **[Getting started](docs/start/quickstart.md)** · **[Writing good annotations](docs/guides/write-good-annotations.md)** · **[CLI reference](docs/reference/cli.md)** · **[CI](docs/guides/enforce-in-ci.md)**

## What koment is not

- **Not general project documentation.** koment holds knowledge bound to a
  *place* in the code; broad project guidance belongs in the documentation
  system your team already uses.
- **Not a memory system.** A consolidating store paraphrases, merges and
  eventually forgets. koment is a record, not a belief.
- **Not line-precise.** Anchors are snippets.

## Prior art

[Codetations / Magic Markup](https://github.com/elmisback/codetations) is the
closest existing work — document-external annotations with LLM re-anchoring. The
anchoring problem dates to [Microsoft Research, 2001](https://www.microsoft.com/en-us/research/wp-content/uploads/2016/02/tr-2001-107.pdf)
and is still unsolved in the general case. The UI owes its shape to
[konflate](https://github.com/home-operations/konflate), which makes an
invisible layer visible for Flux the way koment tries to for rationale.

## License

[AGPL-3.0-or-later](LICENSE) is the open-source grant for the repository except
`integrations/editors/zed/`. That extension code alone is
[GPL-3.0-or-later](integrations/editors/zed/LICENSE) so Zed may build and
distribute it through its registry; the koment binary it starts remains AGPL.
Both license files are verbatim texts so automated scanners identify them
correctly. [ADR 0145](docs/explanation/decisions/0145-license-the-zed-extension-under-gplv3.md)
records the narrow exception.

The complete corresponding source is this repository,
<https://github.com/koment-dev/koment>, which satisfies the AGPL §13 obligation
for anyone interacting with a koment server over a network.

Commercial licences are available on request for organisations whose policy
excludes AGPL, or that want warranty or indemnification — write to
[licensing@koment.dev](mailto:licensing@koment.dev).

Releases ≤ v0.6.0 were distributed under the MIT licence and remain MIT in
perpetuity. The last MIT-tagged source may be forked from
[`v0.6.0`](https://github.com/koment-dev/koment/tree/v0.6.0) under those terms
indefinitely. The decision is recorded in [ADR 0117](docs/explanation/decisions/0117-relicense-to-agpl-with-commercial-dual-licensing.md).

**Does using koment make my code AGPL?** No. koment is a *tool*; annotations
it writes are data, not derivative works of the tool. The same legal class
applies as compiling a program with GCC: the tool's licence governs the tool,
not what you build with it. This is the intent of every OSI-approved AGPL
deployment that processes external input; if your organisation's counsel
disagrees, the commercial licence resolves it.
