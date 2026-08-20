# 0149 — A policy activates automatic enforcement

Date: 2026-08-20
Status: Accepted

## Context

Agent integrations can be installed globally while koment itself remains a
repository-local policy. The hooks currently treat every Git worktree as if it
had opted in: a completion hook in an unrelated repository tries to read
`.koment/policy.yaml`, reports that the file is missing and encourages the
agent to repair a configuration that was never meant to exist. The pre-tool
hook is worse in the opposite direction because it silently substitutes the
strict default policy and rejects comments in a repository that does not use
koment.

The same mismatch exists in the published OpenCode plugin. Its disposal hook
runs `koment check`, `koment comments check` and `koment agents check`
directly, so fixing only one client hook would leave another globally installed
surface producing the same false work.

Absence and damage need different treatment. A repository with no koment state
has made no policy decision. A repository that already contains annotation
records but has lost its policy is incomplete, and silently ignoring those
records would hide real damage.

## Decision

`.koment/policy.yaml` activates automatic koment enforcement for a repository.
A valid policy makes the repository active even when it contains no annotation
records.

When repository discovery finds no policy and no files matching
`.koment/annotations/*.yaml`, the repository is inactive. The three direct
gates — `koment check`, `koment comments check` and `koment agents check` —
exit successfully with no output. Agent pre-tool and completion hooks return
their empty protocol response without parsing or enforcing the request. A
session-start integration emits neither policy instructions nor a warning.

An empty `.koment/` directory, an empty annotations directory and unrelated
files below `.koment/annotations/` do not activate enforcement. This keeps
abandoned directories and editor-created paths from manufacturing a policy.

When one or more annotation YAML files exist but the policy is missing,
discovery fails with an error naming `.koment/policy.yaml` and instructing the
user to run `koment bootstrap`. An existing policy that is unreadable, invalid
or incompatible also fails. Once activated, checks retain the fail-loud
semantics chosen by the earlier agent-policy decisions.

Mutation and onboarding commands do not use this automatic activation gate.
`koment bootstrap`, `koment agents install` and annotation mutations can still
create the repository state they own.

## Consequences

- Global agent integrations are inert in repositories that have not adopted
  koment, so agents do not invent setup work from a missing file.
- A policy-only repository remains active and can enforce comment and adapter
  rules before its first annotation is written.
- Losing a policy after annotations exist is visible rather than silently
  disabling enforcement.
- Automatic commands share one discovery rule instead of each client deciding
  what absence means.
- Explicit read and mutation commands can still report that no repository was
  configured; the no-op contract is limited to automatic policy gates.

## Alternatives rejected

**Use the strict default whenever the policy is absent.** This was the pre-tool
hook's previous fallback. It is safe only after a repository has opted in; in
an unrelated worktree it imposes a policy no owner selected.

**Require every integration to be installed at project scope.** Project scope
is still a useful default, but global configuration is supported by several
clients and easily outlives the repository that motivated it. Correct behavior
cannot depend on every user maintaining perfect client configuration.

**Treat any `.koment/` directory as active.** Empty directories can be created
by an interrupted bootstrap, an editor or an earlier experiment. A directory
alone contains no enforceable decision.

**Ignore a missing policy even when annotation records exist.** This would make
the global case quiet by hiding an incomplete real repository. Annotation
records are durable project state and their missing policy requires repair.

**Fix only the Claude Stop hook.** Hermes, OpenCode, Codex and the MCP pre-tool
surface reach the same binary paths. Client-specific exceptions would drift
and would not solve globally installed integrations consistently.
