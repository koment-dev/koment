package packaging_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	yaml "go.yaml.in/yaml/v3"
)

type registryPackage struct {
	RegistryType string `json:"registryType"`
	Identifier   string `json:"identifier"`
	Version      string `json:"version"`
}

type registryServer struct {
	Version  string            `json:"version"`
	Packages []registryPackage `json:"packages"`
}

func publishedServer(t *testing.T) registryServer {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", "..", "server.json"))
	if err != nil {
		t.Fatalf("read server.json: %v", err)
	}
	var server registryServer
	if err := json.Unmarshal(content, &server); err != nil {
		t.Fatalf("decode server.json: %v", err)
	}
	if len(server.Packages) == 0 {
		t.Fatal("server.json advertises no package")
	}
	return server
}

// release-please rewrites the version fields of server.json but cannot rewrite a
// version embedded in a string. A tag committed here would silently outlive the
// release that produced it, which is the exact failure koment exists to prevent.
func TestTheRegistryImageCarriesNoVersionOfItsOwn(t *testing.T) {
	for _, advertised := range publishedServer(t).Packages {
		if advertised.RegistryType != "oci" {
			continue
		}
		if _, tag, tagged := strings.Cut(advertised.Identifier, ":"); tagged {
			t.Errorf("server.json pins %s to tag %q; release-please updates the version fields but never this string, so it will name a stale image after the next release",
				advertised.Identifier, tag)
		}
		if advertised.Version != "" {
			t.Errorf("server.json gives OCI package %s a version field of %q; the MCP registry rejects that with 400 and wants the version in the identifier instead",
				advertised.Identifier, advertised.Version)
		}
	}
}

// The release workflow supplies the tag the committed identifier deliberately
// omits. If that step stops doing so, the registry would receive an untagged
// image reference.
func TestTheReleaseWorkflowTagsTheImageItPublishesToTheRegistry(t *testing.T) {
	workflow := repositoryFile(t, ".github/workflows/release.yml")

	for _, required := range []string{
		".packages[0].identifier = $image",
		"del(.packages[0].version)",
		"IMAGE: ghcr.io/${{ github.repository }}:${{ needs.please.outputs.version }}",
	} {
		if !strings.Contains(workflow, required) {
			t.Errorf("the MCP registry step no longer supplies a tagged image: missing %q", required)
		}
	}
}

// One release version, one number everywhere it is advertised.
func TestEveryDistributionManifestAgreesOnTheVersion(t *testing.T) {
	server := publishedServer(t)
	release := releasedVersion(t)

	if server.Version != release {
		t.Errorf("server.json says %q, the release manifest says %q", server.Version, release)
	}
	for _, advertised := range server.Packages {
		if advertised.RegistryType == "oci" {
			continue
		}
		if advertised.Version != release {
			t.Errorf("server.json package %s says %q, the release manifest says %q",
				advertised.Identifier, advertised.Version, release)
		}
	}

	for manifest, path := range map[string]string{
		"the Claude plugin":     "integrations/agent-plugins/claude/.claude-plugin/plugin.json",
		"the Hermes plugin":     "integrations/agent-plugins/hermes/plugin.yaml",
		"the OpenCode manifest": "integrations/agent-plugins/opencode/plugin.json",
		"the OpenCode package":  "integrations/agent-plugins/opencode/package.json",
		"the VS Code package":   "integrations/editors/vscode/package.json",
		"the Zed extension":     "integrations/editors/zed/extension.toml",
	} {
		if declared := declaredVersion(t, path); declared != release {
			t.Errorf("%s (%s) says %q, the release manifest says %q", manifest, path, declared, release)
		}
	}

	for field, declared := range vscodeLockVersions(t) {
		if declared != release {
			t.Errorf("the VS Code lock %s says %q, the release manifest says %q", field, declared, release)
		}
	}

	chart := helmChart(t)
	if chart.Version != release {
		t.Errorf("the Helm chart says %q, the release manifest says %q", chart.Version, release)
	}
	if chart.AppVersion != release {
		t.Errorf("the Helm chart ships appVersion %q, the release manifest says %q", chart.AppVersion, release)
	}
}

func TestEveryReleaseBuildStampsTheMCPServerVersion(t *testing.T) {
	stamp := "-X github.com/koment-dev/koment/internal/mcp.serverVersion=VERSION"
	for _, path := range []string{".github/workflows/release.yml", "Dockerfile"} {
		if content := repositoryFile(t, path); !strings.Contains(content, stamp) {
			t.Errorf("%s does not stamp the MCP server with the released version", path)
		}
	}
}

func TestPluginPublicationUsesTheLockedNodeVersion(t *testing.T) {
	locked := regexp.MustCompile(`(?m)^node = "([^"]+)"$`).FindStringSubmatch(
		repositoryFile(t, "integrations/editors/vscode/mise.toml"))
	if locked == nil {
		t.Fatal("integrations/editors/vscode/mise.toml has no locked Node version")
	}
	workflow := repositoryFile(t, ".github/workflows/release.yml")
	if want := "node-version: '" + locked[1] + "'"; !strings.Contains(workflow, want) {
		t.Errorf("plugin publication does not use the locked Node %s toolchain", locked[1])
	}
}

func helmChart(t *testing.T) struct {
	Version    string `yaml:"version"`
	AppVersion string `yaml:"appVersion"`
} {
	t.Helper()
	var chart struct {
		Version    string `yaml:"version"`
		AppVersion string `yaml:"appVersion"`
	}
	content, err := os.ReadFile(filepath.Join("..", "helm", "koment", "Chart.yaml"))
	if err != nil {
		t.Fatalf("read Chart.yaml: %v", err)
	}
	if err := yaml.Unmarshal(content, &chart); err != nil {
		t.Fatalf("decode Chart.yaml: %v", err)
	}
	return chart
}

func releasedVersion(t *testing.T) string {
	t.Helper()
	var manifest map[string]string
	decodeRepositoryJSON(t, ".release-please-manifest.json", &manifest)
	release, named := manifest["."]
	if !named {
		t.Fatal(".release-please-manifest.json does not track the repository root")
	}
	return release
}

func declaredVersion(t *testing.T, path string) string {
	t.Helper()
	if filepath.Ext(path) == ".toml" {
		match := regexp.MustCompile(`(?m)^version = "([^"]+)"$`).FindStringSubmatch(repositoryFile(t, path))
		if match == nil {
			t.Fatalf("%s has no version field", path)
		}
		return match[1]
	}
	if filepath.Ext(path) == ".yaml" {
		var declared struct {
			Version string `yaml:"version"`
		}
		decodeRepositoryYAML(t, path, &declared)
		return declared.Version
	}
	var declared struct {
		Version string `json:"version"`
	}
	decodeRepositoryJSON(t, path, &declared)
	return declared.Version
}

func vscodeLockVersions(t *testing.T) map[string]string {
	t.Helper()
	var lock struct {
		Version  string `json:"version"`
		Packages map[string]struct {
			Version string `json:"version"`
		} `json:"packages"`
	}
	decodeRepositoryJSON(t, "integrations/editors/vscode/package-lock.json", &lock)
	return map[string]string{
		"root version":         lock.Version,
		"root package version": lock.Packages[""].Version,
	}
}

func decodeRepositoryJSON(t *testing.T, path string, into any) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(path)))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(content, into); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}

func decodeRepositoryYAML(t *testing.T, path string, into any) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(path)))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := yaml.Unmarshal(content, into); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}
