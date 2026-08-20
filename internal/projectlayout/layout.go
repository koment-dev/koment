package projectlayout

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const ReferencePath = "docs/reference/repository-layout.md"

const zedLicensePath = "integrations/editors/zed/LICENSE"
const zedManifestPath = "integrations/editors/zed/Cargo.toml"
const zedGPLv3SHA256 = "3972dc9744f6499f0f9b2dbf76696f2ae7ad8af9b23dde66d6af86c9dfb36986"

type Area struct {
	Path    string
	Purpose string
}

var Areas = []Area{
	{Path: ".claude-plugin/", Purpose: "Claude marketplace discovery metadata"},
	{Path: ".codex/", Purpose: "generated Codex repository adapter"},
	{Path: ".cursor/", Purpose: "generated Cursor repository adapter"},
	{Path: ".github/", Purpose: "GitHub workflows, templates and ownership"},
	{Path: ".koment/", Purpose: "authoritative annotations and policy"},
	{Path: ".mise/", Purpose: "pinned toolchain configuration"},
	{Path: ".opencode/", Purpose: "generated OpenCode repository adapter"},
	{Path: ".vscode/", Purpose: "VS Code workspace discovery and validation"},
	{Path: "cmd/", Purpose: "Go binary entry points"},
	{Path: "distribution/", Purpose: "delivery and deployment assets"},
	{Path: "docs/", Purpose: "start, guide, reference and explanation documentation"},
	{Path: "examples/", Purpose: "runnable and inspectable product examples"},
	{Path: "integrations/", Purpose: "code installed into another product"},
	{Path: "internal/", Purpose: "private Go product packages"},
	{Path: "schema/", Purpose: "versioned public schemas"},
	{Path: "scripts/", Purpose: "repository automation"},
	{Path: "testdata/", Purpose: "repository-wide fixtures"},
}

var RootFiles = []string{
	".gitignore",
	".golangci.yml",
	".lefthook.toml",
	".mcp.json",
	".release-please-manifest.json",
	".renovaterc.json5",
	"AGENTS.md",
	"CHANGELOG.md",
	"CLA.md",
	"CLAUDE.md",
	"CONTRIBUTING.md",
	"DESIGN.md",
	"Dockerfile",
	"LICENSE",
	"README.md",
	"SECURITY.md",
	"TRADEMARK.md",
	"action.yml",
	"go.mod",
	"go.sum",
	"opencode.json",
	"release-please-config.json",
	"server.json",
}

var ClosedChildren = map[string][]string{
	"distribution":                  {"helm", "package-managers"},
	"distribution/helm":             {"chart_test.go", "koment"},
	"distribution/package-managers": {"README.md", "homebrew", "naming_test.go", "registry_test.go", "scoop", "winget"},
	"docs":                          {"README.md", "explanation", "guides", "reference", "start"},
	"examples":                      {"annotated-workspace"},
	"integrations":                  {"agent-plugins", "editors"},
	"integrations/agent-plugins":    {"README.md", "claude", "hermes", "opencode"},
	"integrations/editors":          {"vscode", "zed"},
}

var LegacyRoots = []string{"charts", "editors", "packaging", "plugins", "workspace"}

type referenceMigration struct {
	Retired     string
	Replacement string
}

var referenceMigrations = []referenceMigration{
	{Retired: "docs/" + "releasing.md", Replacement: "docs/guides/release-koment.md"},
}

func RepositoryRoot(start string) (string, error) {
	root, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			return root, nil
		}
		parent := filepath.Dir(root)
		if parent == root {
			return "", fmt.Errorf("no repository root above %s", start)
		}
		root = parent
	}
}

func Check(root string) error {
	repositoryRoot, err := os.OpenRoot(root)
	if err != nil {
		return err
	}
	checkError := checkRepository(root, repositoryRoot)
	return errors.Join(checkError, repositoryRoot.Close())
}

func checkRepository(root string, repositoryRoot *os.Root) error {
	paths, err := repositoryPaths(root, repositoryRoot)
	if err != nil {
		return err
	}
	violations := ValidatePaths(paths)
	referenceViolations, err := validateRetiredReferences(repositoryRoot, paths)
	if err != nil {
		return err
	}
	violations = append(violations, referenceViolations...)
	rootViolations, err := validateLegacyRoots(repositoryRoot)
	if err != nil {
		return err
	}
	violations = append(violations, rootViolations...)
	licenseViolations, err := validateZedLicense(repositoryRoot)
	if err != nil {
		return err
	}
	violations = append(violations, licenseViolations...)
	current, err := repositoryRoot.ReadFile(ReferencePath)
	if err != nil {
		violations = append(violations, fmt.Sprintf("%s: %v", ReferencePath, err))
	} else if !bytes.Equal(current, Reference()) {
		violations = append(violations, ReferencePath+": generated reference is stale; run mise run layout-render")
	}
	if len(violations) == 0 {
		return nil
	}
	sort.Strings(violations)
	return errors.New(strings.Join(violations, "\n"))
}

func validateZedLicense(repositoryRoot *os.Root) ([]string, error) {
	license, err := repositoryRoot.ReadFile(zedLicensePath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", zedLicensePath, err)
	}
	manifest, err := repositoryRoot.ReadFile(zedManifestPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", zedManifestPath, err)
	}
	violations := []string{}
	if fmt.Sprintf("%x", sha256.Sum256(license)) != zedGPLv3SHA256 {
		violations = append(violations, zedLicensePath+": expected the verbatim GPLv3 text required by ADR 0145")
	}
	if !bytes.Contains(manifest, []byte("license = \"GPL-3.0-or-later\"")) {
		violations = append(violations, zedManifestPath+": expected license = \"GPL-3.0-or-later\" required by ADR 0145")
	}
	return violations, nil
}

func ValidatePaths(paths []string) []string {
	allowedRoots := map[string]bool{}
	for _, area := range Areas {
		allowedRoots[strings.TrimSuffix(area.Path, "/")] = true
	}
	allowedFiles := map[string]bool{}
	for _, file := range RootFiles {
		allowedFiles[file] = true
	}
	legacy := map[string]bool{}
	for _, name := range LegacyRoots {
		legacy[name] = true
	}
	closed := map[string]map[string]bool{}
	for parent, children := range ClosedChildren {
		closed[parent] = map[string]bool{}
		for _, child := range children {
			closed[parent][child] = true
		}
	}

	violations := []string{}
	for _, repositoryPath := range paths {
		parts := strings.Split(filepath.ToSlash(repositoryPath), "/")
		if len(parts) == 1 {
			if !allowedFiles[parts[0]] {
				violations = append(violations, repositoryPath+": root file is outside the layout contract")
			}
			continue
		}
		if legacy[parts[0]] {
			violations = append(violations, repositoryPath+": legacy root must be migrated")
			continue
		}
		if !allowedRoots[parts[0]] {
			violations = append(violations, repositoryPath+": root directory is outside the layout contract")
			continue
		}
		for depth := 1; depth < len(parts); depth++ {
			parent := strings.Join(parts[:depth], "/")
			children, isClosed := closed[parent]
			if isClosed && !children[parts[depth]] {
				violations = append(violations, repositoryPath+": "+parts[depth]+" is not allowed under "+parent)
				break
			}
		}
	}
	return violations
}

func Reference() []byte {
	var output strings.Builder
	output.WriteString("# Repository layout\n\n")
	output.WriteString("Generated from `internal/projectlayout`; edit the executable specification, not this file.\n\n")
	output.WriteString("The repository root is a closed contract. A tracked or non-ignored path outside the areas and exact root files below fails `mise run layout-check`.\n\n")
	output.WriteString("The same check rejects repository-controlled references to paths retired by completed migrations. Historical path provenance under `.koment/` is excluded.\n\n")
	output.WriteString("## Architectural areas\n\n")
	output.WriteString("| Path | Owner |\n|---|---|\n")
	for _, area := range Areas {
		fmt.Fprintf(&output, "| `%s` | %s |\n", area.Path, area.Purpose)
	}
	output.WriteString("\n## Closed categories\n\n")
	parents := make([]string, 0, len(ClosedChildren))
	for parent := range ClosedChildren {
		parents = append(parents, parent)
	}
	sort.Strings(parents)
	for _, parent := range parents {
		children := append([]string(nil), ClosedChildren[parent]...)
		sort.Strings(children)
		fmt.Fprintf(&output, "- `%s/`: `%s`\n", parent, strings.Join(children, "`, `"))
	}
	output.WriteString("\n## Exact root files\n\n")
	files := append([]string(nil), RootFiles...)
	sort.Strings(files)
	for _, file := range files {
		fmt.Fprintf(&output, "- `%s`\n", file)
	}
	output.WriteString("\n## Changing the contract\n\n")
	output.WriteString("A boundary change must supersede ADR 0143, demonstrate why no existing area can own the capability, update `DESIGN.md` and this specification, regenerate this page, and migrate every path and reference in the same change. Convenience, file count, implementation language and symmetry are insufficient reasons.\n")
	return []byte(output.String())
}

func WriteReference(root string) error {
	repositoryRoot, err := os.OpenRoot(root)
	if err != nil {
		return err
	}
	writeError := repositoryRoot.WriteFile(ReferencePath, Reference(), 0o644)
	return errors.Join(writeError, repositoryRoot.Close())
}

func repositoryPaths(root string, repositoryRoot *os.Root) ([]string, error) {
	command := exec.Command("git", "ls-files", "--cached", "--others", "--exclude-standard", "-z")
	command.Dir = root
	content, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("list repository paths: %w", err)
	}
	paths := []string{}
	for _, item := range bytes.Split(content, []byte{0}) {
		if len(item) == 0 {
			continue
		}
		path := string(item)
		if _, err := repositoryRoot.Lstat(filepath.FromSlash(path)); err == nil {
			paths = append(paths, path)
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("inspect %s: %w", path, err)
		}
	}
	return paths, nil
}

func validateRetiredReferences(repositoryRoot *os.Root, paths []string) ([]string, error) {
	violations := []string{}
	for _, repositoryPath := range paths {
		if strings.HasPrefix(filepath.ToSlash(repositoryPath), ".koment/") {
			continue
		}
		content, err := repositoryRoot.ReadFile(filepath.FromSlash(repositoryPath))
		if err != nil {
			return nil, fmt.Errorf("inspect references in %s: %w", repositoryPath, err)
		}
		for _, migration := range referenceMigrations {
			if bytes.Contains(content, []byte(migration.Retired)) {
				violations = append(violations, fmt.Sprintf("%s: retired reference %q must be replaced with %q", repositoryPath, migration.Retired, migration.Replacement))
			}
		}
	}
	return violations, nil
}

func validateLegacyRoots(repositoryRoot *os.Root) ([]string, error) {
	violations := []string{}
	for _, name := range LegacyRoots {
		_, err := repositoryRoot.Stat(name)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("inspect legacy root %s: %w", name, err)
		}
		violations = append(violations, name+"/: legacy root must be migrated")
	}
	return violations, nil
}
