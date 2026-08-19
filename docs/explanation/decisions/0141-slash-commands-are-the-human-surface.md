# 0141 — Slash commands are the human surface, the skill is the agent's

Date: 2026-08-19
Status: Accepted

## Context

The Claude Code plugin shipped a `SessionStart` hook, a `Stop` hook and one
skill, and no commands. Every koment action a person wanted was reached by
describing it in prose and hoping the model chose the right tool.

That is a worse deal than it looks, because the operations differ sharply in how
much they cost when done wrong. `koment search` is free to get wrong. `koment
forget` destroys a record whose entire purpose was to outlive the code it
described, and the failure is silent — the gate goes green, which is what makes
it attractive to a model trying to finish.

The skill already carries the procedure. It is loaded as standing context and
describes what to do before editing a file, which is exactly right for the
agent-initiated path and exactly wrong as a place to put "confirm before you
delete this": nothing in a standing procedure fires at the moment of a specific
destructive call.

memini's plugin, read at `~/.claude/plugins/cache/memini/memini/0.7.14`, splits
these: `skills/` for standing behaviour, `commands/` for a request, and its
`commands/forget.md` spends most of its length on the confirmation protocol.

## Decision

The plugin ships both surfaces, with a stated division.

**The skill is the agent's standing procedure.** Read annotations before
editing, search before deciding, do not write inline comments, do not finish
while a gate fails. It is unchanged by this ADR.

**A command is a person asking for one thing now**, and it carries the
constraint that makes that one thing safe. Eight commands: `check`, `show`,
`search`, `add`, `convert`, `reanchor`, `forget`, `bootstrap`.

The constraints are the reason the commands exist, not decoration on them:

- `forget` requires the full body shown and explicit confirmation, refuses to
  delete more than one record per confirmation, and states that deleting to make
  `koment check` pass is prohibited.
- `add` requires the rename / extract / named-type / restructure ladder to be
  tried and named first.
- `convert` fixes the order — record, then delete — and lists the comment
  classes that must stay in source.
- `check` refuses to report success while any of the three gates fails.

The `plugins` CI job validates that every file in the declared `commands`
directory exists and carries frontmatter with a description, so renaming one
without updating the manifest fails the build rather than silently removing a
command.

Commands are Claude Code's format. Other runtimes reach the same behaviour
through the generated adapters and the MCP tools, and this ADR does not promise
them a port.

## Consequences

- The dangerous operations now state their own guard rails at the point of use,
  where a standing procedure could not reach.
- The plugin has two places describing koment's procedure, and they can
  disagree. The division — standing versus requested — is what keeps that from
  being duplication, and it has to be held deliberately when either changes.
- Eight prose files are eight more things that can rot. The CI check catches a
  missing or malformed one; it cannot catch one that describes a flag that no
  longer exists. ADR 0137 applies to them like any other documentation.
- Claude Code users get a surface OpenCode and Codex users do not, so the
  runtimes are no longer at parity.

## Alternatives rejected

**Commands only, folding the skill into them.** A command fires when typed. The
behaviour koment most needs — read the annotations *before* editing a file the
agent chose to edit — is never requested by a person, so it would simply stop
happening. The skill is the only surface that reaches the agent-initiated path.

**Skill only, adding the confirmation language to it.** This is what exists
today and it does not work: a standing instruction to confirm before deleting
competes with everything else in context at the moment the model is trying to
close a task. Attaching the guard to the invocation is the point.

**Generate the commands from the CLI registry** (ADR 0131 already drives
`--help` from one registry). Rejected because the useful content is not the
flags — it is the refusals, which no registry knows. Generating them would
produce eight synopses and lose the reason for having them.

**Generate them per-runtime through `koment agents install`.** It would keep
the runtimes at parity and prevent drift, and it is the right shape if a second
runtime adopts the format. Deferred rather than rejected: OpenCode and Codex
have no equivalent surface today, so the generator would have one target and
would be a mechanism built ahead of its need.
