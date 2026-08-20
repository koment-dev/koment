# Cutting a release

**This procedure is mandatory. Follow it in order. Do not improvise a release.**

A release publishes signed artifacts to GitHub, GHCR, the MCP Registry, the VS
Code Marketplace and Open VSX. Marketplace versions are permanent: a version
number cannot be reused, replaced or withdrawn. A mistake here is not
recoverable by editing a file, so every step below exists because skipping it
produces something a user installs and cannot undo.

Agents may prepare steps 0–3, perform the read-only verification in step 6 and
prepare the development-pin change in step 8. **Steps 4, 5 and 7 change the
outside world and require explicit human approval in the conversation.**

---

## 0. Preconditions

Verify all of these before starting. If any fails, stop and report it.

Before a release that promises Zed registry support, verify that
`integrations/editors/zed/LICENSE` is the GPLv3 text and that the crate declares
`GPL-3.0-or-later`. Zed requires an accepted license inside an extension
subdirectory; `mise run layout-check` enforces both files as part of ADR 0145.

```sh
mise run fmt-check && mise run vet && mise run tidy-check && mise run generate-check
mise run lint && mise run test
mise run annotations && mise run comments && mise run agent-policy
mise run workflow-lint && mise run release-helper-test
mise run helm-lint && mise run helm-template && mise run vulncheck
mise run extension-test && mise run layout-check
mise --cd integrations/editors/zed run check
mise --cd integrations/editors/zed run build
```

All eighteen must pass. `koment check` failing means a release would ship
annotations that no longer describe the code.

The read-only publication verification in step 6 also needs authenticated
`gh`, a current supported `cosign`, `curl`, `docker` with Buildx, `helm`, `npm`
and `sha256sum` on `PATH`:

```sh
for command in gh cosign curl docker helm npm sha256sum; do command -v "$command"; done
gh auth status
```

`main` is protected by a ruleset — pull request required, signed commits,
linear history, and one required status check: **`ci`**. That is the
aggregating job in `.github/workflows/ci.yml`; it depends on every gating job,
so adding a gating job to CI does not require a ruleset edit. The setup-action
and Windows archive jobs are advisory because they test the last published
release rather than the pull request. `cla`, `codeql` and `scorecard` also
report separately, and `cla` cannot be required because a release pull request
opened by `GITHUB_TOKEN` never gets a `cla` run. The classic
branch-protection API returns 404 for this repository; that means the rules
live in a ruleset, not that the branch is unprotected. Check with:

```sh
gh api repos/koment-dev/koment/rulesets
```

## 1. Land the work

Merge every change through a pull request with a conventional subject. The
subject decides the version, so it is a release decision, not a formatting one:

| Subject | Effect |
|---|---|
| `feat:` | minor bump — 1.0.0 → 1.1.0 |
| `fix:`, `perf:`, `refactor:` | patch bump |
| `docs:`, `test:`, `build:`, `ci:` | patch bump, listed in the changelog |
| `chore:` | no release on its own |
| any `!` or `BREAKING CHANGE:` | major bump — 1.4.2 → 2.0.0 |

`bump-minor-pre-major` was turned off when 1.0.0 shipped (ADR 0120), so a `!`
is a major version and not a quiet minor one. Before writing `!`, check whether
the change is breaking at all: a claim of backward compatibility needs a
migration the binary performs or an ADR naming the version the old shape was
cut off at. Without either, it is breaking, and the subject has to say so.

## 2. Let release-please open the release pull request

Pushing to `main` runs the `release` workflow, whose first job opens or updates
a pull request titled `chore(release): <version>`. It edits the changelog, the
manifest, and every file that carries the version:

- `.release-please-manifest.json`
- `distribution/helm/koment/Chart.yaml`
- `integrations/editors/vscode/package.json`
- `integrations/editors/vscode/package-lock.json`
- `integrations/editors/zed/extension.toml`
- `integrations/agent-plugins/claude/.claude-plugin/plugin.json`
- `integrations/agent-plugins/hermes/plugin.yaml`
- `integrations/agent-plugins/opencode/plugin.json`
- `integrations/agent-plugins/opencode/package.json`
- `server.json` — the top-level `.version`; OCI packages are versioned by tag

Do not edit these by hand and do not bump a version in a feature branch.
The package-manager parity tests fail the build when they disagree.

## 3. Unblock that pull request's checks

**Expect its CI to sit at `action_required` with a 0s duration.** GitHub does
not run workflows for events created by `GITHUB_TOKEN`, so the required checks
never start and the pull request cannot merge on its own. This is normal and is
not a failure.

```sh
gh run list --branch release-please--branches--main --limit 5
gh api -X POST repos/koment-dev/koment/actions/runs/<run-id>/approve
```

Then wait for `commit-lint`, `container`, `editor`, `helm`, `lint`, `plugins`,
`quality`, `test` and `zed` to pass and for the aggregate `ci` job to pass.
Never merge a release pull request whose checks did not run—an unapproved run
is not a passing run. Quote `gh pr checks <number>` before asking to merge.

## 4. Merge the release pull request — human approval required

Merging tags the release and starts publication. Everything after this point is
public and permanent.

Before merging, confirm:

- the version in the title is the one you intend;
- the changelog describes real changes;
- `ci` is green and not skipped, and every job it aggregates ran.

## 5. Watch publication — human approval required to retry anything

Merging runs the rest of the `release` workflow in this order. The order is a
decision, not an accident (ADR 0109): canonical artifacts first, downstream
channels second.

```
please ──┬─> binaries ──┬─> editor
         │              ├─> tap
         │              └─> verify ──> alias
         └─> image

binaries + image ──┬─> plugins
                   ├─> mcp-registry
                   └─> chart
```

| Job | Publishes |
|---|---|
| `binaries` | six archives, `koment_<version>_checksums.txt`, a cosign signature, and rendered Homebrew/Scoop/WinGet metadata |
| `plugins` | Three self-contained plugin archives (Claude, Hermes and OpenCode), per-archive cosign signatures, a combined `koment-plugins_<version>_checksums.txt`, and its cosign signature; the OpenCode package also goes to npm |
| `image` | `ghcr.io/koment-dev/koment:<version>`, multi-arch, SBOM and provenance, cosign-signed |
| `editor` | seven VSIX — six carrying that platform's released binary, one universal — signed, attached, then pushed to both marketplaces |
| `tap` | the rendered formula in `koment-dev/homebrew-tap` after the binary assets exist |
| `verify` | installs the new GitHub release through this repository's setup Action on Linux and macOS |
| `alias` | moves the floating major tag only after both setup-Action verification jobs pass |
| — | the Zed extension is **not** built or attached by the release. Zed builds it from the submodule, so step 7 publishes it by hand |
| `mcp-registry` | MCP Registry metadata via GitHub OIDC |
| `chart` | `oci://ghcr.io/koment-dev/charts/koment`, cosign-signed |

`plugins` waits for both `binaries` and `image`. Its source archives do not
consume either artifact, but ADRs 0109 and 0129 require canonical artifacts to
exist before a downstream npm package or agent integration is published.
The chart and MCP Registry metadata use the same gate so a missing GitHub
release asset cannot leave them as the only completed distribution channels.

```sh
gh run watch "$(gh run list --workflow=release --limit 1 --json databaseId --jq '.[0].databaseId')"
```

If `binaries` fails, `editor` does not run. That is deliberate: the extension
bundles the released binary, so an extension built without one would ship
something that was never signed (ADR 0113).

## 6. Verify the release, do not assume it

```sh
tag=v<version>
gh release view "$tag" --json assets --jq '.assets[].name' | sort
curl -fsSLI -o /dev/null -w '%{http_code}\n' "https://open-vsx.org/api/koment/koment-dev"
curl -fsSLI -o /dev/null -w '%{http_code}\n' "https://marketplace.visualstudio.com/items?itemName=koment.koment-dev"
```

Expect six platform archives, the WinGet submission bundle, the binary checksum
manifest and its signature, rendered Homebrew and Scoop metadata, three plugin
archives and their three signatures, the plugin checksum manifest and its
signature, the chart and its signature, and seven VSIX files with seven
signatures. A release missing the archives breaks
every workflow using `koment-dev/koment@v<version>`, because the setup action
downloads them.

Download and verify the immutable assets rather than relying on their names:

```sh
mkdir "koment-release-<version>"
cd "koment-release-<version>"
gh release download "$tag" --repo koment-dev/koment
sha256sum --check "koment_<version>_checksums.txt"
sha256sum --check "koment-plugins_<version>_checksums.txt"
certificate_identity="https://github.com/koment-dev/koment/.github/workflows/release.yml@refs/heads/main"
certificate_issuer="https://token.actions.githubusercontent.com"
cosign verify-blob \
  --bundle "koment_<version>_checksums.sigstore.json" \
  --certificate-identity "$certificate_identity" \
  --certificate-oidc-issuer "$certificate_issuer" \
  "koment_<version>_checksums.txt"
cosign verify-blob \
  --bundle "koment-plugins_<version>_checksums.sigstore.json" \
  --certificate-identity "$certificate_identity" \
  --certificate-oidc-issuer "$certificate_issuer" \
  "koment-plugins_<version>_checksums.txt"
for artifact in koment-plugin-*.tar.gz koment-vscode_*.vsix koment-<version>.tgz; do
  cosign verify-blob \
    --bundle "${artifact}.sigstore.json" \
    --certificate-identity "$certificate_identity" \
    --certificate-oidc-issuer "$certificate_issuer" \
    "$artifact"
done
npm view "@koment/opencode-koment@<version>" version
cosign verify \
  --certificate-identity "$certificate_identity" \
  --certificate-oidc-issuer "$certificate_issuer" \
  "ghcr.io/koment-dev/koment:<version>"
cosign verify \
  --certificate-identity "$certificate_identity" \
  --certificate-oidc-issuer "$certificate_issuer" \
  "ghcr.io/koment-dev/charts/koment:<version>"
docker buildx imagetools inspect "ghcr.io/koment-dev/koment:<version>"
helm show chart "oci://ghcr.io/koment-dev/charts/koment" --version "<version>"
curl -fsSL "https://raw.githubusercontent.com/koment-dev/homebrew-tap/main/Formula/koment.rb" \
  | grep -F 'version "<version>"'
```

Run these in a new temporary directory and remove it after recording the real
output. Do not run `gh release download` into the repository checkout.

## 7. Publish the Zed extension — human approval required

Zed's registry is not driven by this repository's release workflow. It builds the
extension from a git submodule, so publishing is a pull request in another
organisation's repository and a person opens it (ADR 0139).

Do not start this step until the GPLv3 boundary required by step 0 passes the
local layout check.

Nothing else in this procedure waits on it, and the release is not fully
published until it merges.

1. Fork or update your fork of `zed-industries/extensions`.
2. Point the `koment` submodule at the release tag, and set the version in the
   top-level `extensions.toml` to the version just released:

   ```toml
   [koment]
   submodule = "extensions/koment"
   path = "integrations/editors/zed"
   version = "<version>"
   ```

3. Run `pnpm sort-extensions` so `extensions.toml` and `.gitmodules` stay
   sorted. A pull request that skips this is rejected.
4. Open the pull request. The submodule commit must be reachable on a branch,
   not detached, and the submodule URL must be HTTPS rather than SSH.

On the first publication the entry and the submodule are both new; afterwards
only the submodule commit and the `version` field change.

Verify after it merges:

```sh
curl -fsSLI -o /dev/null -w '%{http_code}\n' "https://zed.dev/extensions/koment"
```

## 8. Bump the development pin

`.mise/config.toml` pins `github:koment-dev/koment` to a released version, which is
the `koment` a contributor gets in their shell. It is not what any gate runs —
every `mise run` task uses `go run ./cmd/koment` — so it lags a release rather
than blocking one. It still has to be caught up, because a pinned binary older
than the record shape in `.koment/` cannot read this repository at all.

```sh
mise use "github:koment-dev/koment@<version>"
mise run annotations
```

Land it as `chore:`, which release-please does not turn into a release of its
own.

---

## When something goes wrong

**Never delete a tag or a published release to "redo" it.** Republish nothing.
Cut the next patch version instead. A tag that once existed has been fetched by
someone.

| Symptom | Cause | Action |
|---|---|---|
| release pull request checks show `action_required`, 0s | `GITHUB_TOKEN` created the pull request | approve the run (step 3) |
| release asset upload returns an HTTP error | the create-release response or publishing token cannot upload to its exact asset endpoint | do not rerun a published registry version; inspect the response, fix the cause, and cut the next patch |
| `editor` job skipped | `binaries` failed | fix the binaries, cut a new patch version |
| `ovsx publish` fails on the first ever publish | the namespace did not exist | the workflow now creates it; if it still fails, the token lacks the Publisher Agreement |
| `vsce publish` rejects the version | that version already exists on the marketplace | cut the next version, never reuse one |
| `Windows Archive (advisory)` is red | advisory by decision (ADR 0111) | it does not block; read it and file a task |
| version files disagree | someone hand-edited one | revert the edit, let release-please own them |

## What an agent must never do

- Publish a VSIX, image, chart or binary by hand, outside this workflow.
  ADR 0112 rejected out-of-band publishing: a marketplace would carry a version
  the release never produced.
- Hand-edit any version-bearing file listed in step 2.
- Merge a release pull request whose checks did not run.
- Delete, move or re-point a tag, or force-push `main`.
- Re-run a publish job hoping it works the second time, without first
  establishing why it failed.
- Claim a release succeeded without running step 6 and quoting its output.
