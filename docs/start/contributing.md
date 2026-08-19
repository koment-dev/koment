# Set up a development checkout

Follow these steps in order before changing koment.

## 1. Read the contracts

Read [AGENTS.md](../../AGENTS.md) in full. Then read
[DESIGN.md](../../DESIGN.md) and the
[active architecture decisions](../explanation/decisions/README.md).

The source file tells you what the code does. Its koment annotations tell you
why that exact code exists. The architecture decisions record choices that
apply to more than one source location.

## 2. Install the pinned tools

Install [mise](https://mise.jdx.dev/), clone the repository, enter the checkout
and run:

```sh
mise install
```

The committed mise configuration and lock file select the same Go and repository
tools that CI uses. The install also enables the repository's Lefthook hooks.

## 3. Establish a clean baseline

Run the complete local gate before attributing a failure to your change:

```sh
mise run build
mise run fmt-check
mise run tidy-check
mise run vet
mise run test
mise run lint
mise run vulncheck
mise run workflow-lint
mise run generate-check
mise run layout-check
mise run annotations
mise run comments
mise run agent-policy
```

A release or pull request additionally requires its GitHub Actions checks to
pass. Local output cannot certify CI.

## 4. Read the rationale for the files you will change

Use the MCP tools when your client exposes them:

```text
koment_get("<repository-relative path>")
koment_search("<the decision or behavior>")
```

Otherwise use the local binary:

```sh
go run ./cmd/koment show <repository-relative-path>
go run ./cmd/koment search '<the decision or behavior>'
```

Stop and surface any ambiguous, drifted or orphaned annotation before editing.

## 5. Make one reviewable change

Update `DESIGN.md` and obtain agreement before implementing a change beyond a
bugfix. Record a project-wide decision in a new ADR. Record rationale bound to a
specific source location as a koment annotation.

Do not add an ordinary explanatory source comment. Rename, extract, introduce a
named type or constant, or restructure first.

## 6. Finish with the repository gates

Run:

```sh
koment check
koment comments check
koment agents check
git diff --check
```

Then run the relevant build and test tasks again. Report the exact output, what
was not verified and anything that went wrong.

Continue with [development](../guides/develop-koment.md) for project conventions,
[the repository layout](../reference/repository-layout.md) for file ownership,
and [the release procedure](../guides/release-koment.md) only when preparing a
release.
