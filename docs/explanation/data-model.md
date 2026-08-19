# How annotations remain trustworthy

koment stores the reason for code beside the code rather than inside it. An
annotation is prose attached to a source excerpt and committed under
`.koment/annotations/`. When that excerpt no longer resolves, koment reports
the uncertainty instead of serving the prose as current fact.

## The model

```text
Deployment
└── Repository        assigned identity and synchronization state
    └── Commit        historical context captured when rationale was written
        └── File      repository-relative path at that commit
            └── Annotation
```

A local checkout discovers its repository by walking upward for `.koment/`
and then `.git/`. A configured service can assign several repositories and
serve atomic snapshots of their commits.

One annotation record looks like this:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/koment-dev/koment/main/schema/v1alpha/annotation.schema.json
apiVersion: koment.dev/v1alpha
kind: Annotation
metadata:
  id: 01KYW1ETE3CVB6S0ND70GGZVWM
  created: "2026-08-02T14:02:00Z"
spec:
  target:
    file: internal/store/ulid.go
  type: gotcha
  body: |-
    26 Crockford characters carry 130 bits but a ULID holds 128 …
  anchor:
    scope: excerpt
    excerpt: paddingBits = ulidLength*bitsPerChar - 8*(...)
    before: const (
    after: )
  git:
    commit: 9f3c1a4d8e2b7c5a...
    path: internal/store/ulid.go
    line: 18
  author:
    name: Jan Pucilowski
    kind: human
    source: git-config
status:
  lastSeenLine: 18
```

## Applicability and history are separate

The excerpt asks whether the explanation still applies now. Resolution searches
the current file for that exact text and captured context. The Git block records
what was true when the annotation was written; it never selects an anchor.
Deleting every Git block would change no current resolution status.

Resolution returns `ok`, `ambiguous`, `drifted` or `orphaned`. The last
three fail `koment check`. There is no moved status: a uniquely resolved
excerpt is `ok` wherever it appears.

## Git remains authoritative

Each `.koment/annotations/<id>.yaml` file is the reviewed record. CLI, UI, MCP,
LSP, static pages and served snapshots derive their views from those files and
the source tree. Their indexes and rendered output are disposable; none can
restore or overwrite the Git record.

[ADR 0100](decisions/0100-one-git-record-per-annotation.md) records the
one-record model, [ADR 0101](decisions/0101-fail-ambiguous-anchor-resolution.md)
defines resolution, and
[ADR 0102](decisions/0102-one-repository-snapshot-for-every-reader.md) keeps all
read surfaces on the same snapshot.
