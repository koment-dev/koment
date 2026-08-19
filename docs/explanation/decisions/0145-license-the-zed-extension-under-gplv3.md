# 0145 — License only the Zed extension under GPLv3

Date: 2026-08-19
Status: Accepted

## Context

ADR 0117 made AGPL-3.0-or-later the repository license and kept commercial
licensing available. The Zed extension registry does not accept AGPL. Its
published license requirements do accept GNU GPLv3, require the accepted
license file inside an extension subdirectory, and explicitly say that the
requirement applies to the extension code rather than a language server or
other program the extension invokes.

The extension in `integrations/editors/zed/` is a small WebAssembly adapter. It
starts the separately installed `koment` binary and contains no product server,
anchoring logic or binary payload. Keeping it under AGPL would leave the
implemented integration permanently outside Zed's registry for a distinction
that does not protect the AGPL product.

## Decision

License only the contents of `integrations/editors/zed/` under
GPL-3.0-or-later. Put the verbatim GPLv3 text in that directory and declare the
same SPDX expression in its Cargo manifest.

The repository-root AGPL-3.0-or-later license remains the default for every
other path, including the `koment` binary the extension starts. Commercial
licensing under ADR 0117 remains available for that AGPL work. This decision
narrows ADR 0117 at one distribution boundary; it does not supersede its
product-license decision.

The repository layout check verifies the extension's license text and Cargo
declaration. A release promising Zed support cannot pass if either side of the
boundary drifts.

## Consequences

Easier:

- Zed can validate, compile and distribute the extension through its registry.
- Users can install the integration from Zed without changing the license of
  the binary, server or any annotated project.
- The exception is machine-checked at the same gate that protects the repository
  structure.

Harder:

- The repository carries two copyleft license texts and contributors must know
  which directory they are changing.
- Code moved across the Zed directory boundary changes its public license and
  therefore needs explicit review.
- A future Zed license-policy change may require a new decision and another
  registry-compatible boundary.

Committed to:

- Keeping the GPL grant confined to `integrations/editors/zed/`.
- Keeping license claims in the manifest, documentation and local license file
  synchronized.
- Treating any broader relicensing as a separate owner decision.

## Alternatives rejected

- **Keep the extension AGPL and omit registry publication.** Preserves one
  repository-wide license but makes the completed integration undiscoverable to
  the users it serves. Zed CI rejects it, so documentation could not honestly
  call the extension released.
- **Relicense the whole repository under GPLv3.** Satisfies Zed but removes the
  network-use source obligation chosen for `koment serve` in ADR 0117. An
  editor registry constraint is not a reason to weaken the product license.
- **Use MIT or Apache-2.0 for the extension.** Both are accepted, but a
  permissive exception grants more than Zed needs. GPLv3 is the closest
  accepted license to the repository's existing copyleft terms.
- **Dual-license the extension as AGPL or GPL.** Adds an unnecessary choice and
  risks automated validation ambiguity. A single accepted SPDX expression and
  one local text make the distribution terms unambiguous.
- **Move the extension to another repository.** Avoids a mixed-license tree at
  the cost of split versioning, duplicated release coordination and easier
  drift between the adapter and the binary protocol. The existing closed
  integration boundary already expresses the separation.
