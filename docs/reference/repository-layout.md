# Repository layout

Generated from `internal/projectlayout`; edit the executable specification, not this file.

The repository root is a closed contract. A tracked or non-ignored path outside the areas and exact root files below fails `mise run layout-check`.

## Architectural areas

| Path | Owner |
|---|---|
| `.claude-plugin/` | Claude marketplace discovery metadata |
| `.codex/` | generated Codex repository adapter |
| `.cursor/` | generated Cursor repository adapter |
| `.github/` | GitHub workflows, templates and ownership |
| `.koment/` | authoritative annotations and policy |
| `.mise/` | pinned toolchain configuration |
| `.opencode/` | generated OpenCode repository adapter |
| `.vscode/` | VS Code workspace discovery and validation |
| `cmd/` | Go binary entry points |
| `distribution/` | delivery and deployment assets |
| `docs/` | start, guide, reference and explanation documentation |
| `examples/` | runnable and inspectable product examples |
| `integrations/` | code installed into another product |
| `internal/` | private Go product packages |
| `schema/` | versioned public schemas |
| `scripts/` | repository automation |
| `testdata/` | repository-wide fixtures |

## Closed categories

- `distribution/`: `helm`, `package-managers`
- `distribution/helm/`: `chart_test.go`, `koment`
- `distribution/package-managers/`: `README.md`, `homebrew`, `naming_test.go`, `registry_test.go`, `scoop`, `winget`
- `docs/`: `README.md`, `explanation`, `guides`, `reference`, `start`
- `examples/`: `annotated-workspace`
- `integrations/`: `agent-plugins`, `editors`
- `integrations/agent-plugins/`: `README.md`, `claude`, `hermes`, `opencode`
- `integrations/editors/`: `vscode`, `zed`

## Exact root files

- `.gitignore`
- `.golangci.yml`
- `.lefthook.toml`
- `.mcp.json`
- `.release-please-manifest.json`
- `.renovaterc.json5`
- `AGENTS.md`
- `CHANGELOG.md`
- `CLA.md`
- `CLAUDE.md`
- `CONTRIBUTING.md`
- `DESIGN.md`
- `Dockerfile`
- `LICENSE`
- `README.md`
- `SECURITY.md`
- `TRADEMARK.md`
- `action.yml`
- `go.mod`
- `go.sum`
- `opencode.json`
- `release-please-config.json`
- `server.json`

## Changing the contract

A boundary change must supersede ADR 0143, demonstrate why no existing area can own the capability, update `DESIGN.md` and this specification, regenerate this page, and migrate every path and reference in the same change. Convenience, file count, implementation language and symmetry are insufficient reasons.
