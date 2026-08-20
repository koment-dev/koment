package policy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/koment-dev/koment/internal/store"
)

func TestDetectTreatsRepositoriesWithoutPolicyOrAnnotationsAsInactive(t *testing.T) {
	for name, configure := range map[string]func(string) error{
		"no repository markers": func(string) error { return nil },
		"git repository": func(root string) error {
			return os.Mkdir(filepath.Join(root, ".git"), 0o755)
		},
		"empty koment directory": func(root string) error {
			return os.Mkdir(filepath.Join(root, store.DirName), 0o755)
		},
		"empty annotations directory": func(root string) error {
			return os.MkdirAll(filepath.Join(root, store.DirName, "annotations"), 0o755)
		},
		"unrelated annotations file": func(root string) error {
			directory := filepath.Join(root, store.DirName, "annotations")
			if err := os.MkdirAll(directory, 0o755); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(directory, "notes.txt"), []byte("not an annotation"), 0o644)
		},
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			if err := configure(root); err != nil {
				t.Fatal(err)
			}
			activation, err := Detect(root)
			if err != nil || activation != nil {
				t.Fatalf("Detect() = %#v, %v", activation, err)
			}
		})
	}
}

func TestDetectActivatesAValidPolicyWithoutAnnotations(t *testing.T) {
	root := t.TempDir()
	if err := Save(root, Default()); err != nil {
		t.Fatal(err)
	}
	activation, err := Detect(root)
	if err != nil {
		t.Fatal(err)
	}
	if activation == nil || activation.Root != root || activation.Configured.APIVersion != APIVersion {
		t.Fatalf("Detect() = %#v", activation)
	}
}

func TestDetectRejectsAnnotationRecordsWithoutPolicy(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, store.DirName, "annotations")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "broken.yaml"), []byte(":"), 0o644); err != nil {
		t.Fatal(err)
	}
	activation, err := Detect(root)
	if activation != nil || err == nil {
		t.Fatalf("Detect() = %#v, %v", activation, err)
	}
	for _, wanted := range []string{".koment/annotations", FileName, "koment bootstrap"} {
		if !strings.Contains(err.Error(), wanted) {
			t.Errorf("error missing %q: %v", wanted, err)
		}
	}
}

func TestDetectRejectsAnInvalidExistingPolicy(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, store.DirName), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, FileName), []byte("apiVersion: invalid\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	activation, err := Detect(root)
	if activation != nil || err == nil || !strings.Contains(err.Error(), "incompatible") {
		t.Fatalf("Detect() = %#v, %v", activation, err)
	}
}

func TestDefaultRoundTripsAndMatchesExcludedPaths(t *testing.T) {
	root := t.TempDir()
	configured := Default()
	configured.Spec.Comments.GeneratedPaths = append(configured.Spec.Comments.GeneratedPaths, "generated/**")
	if err := Save(root, configured); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.Excludes("nested/model.generated.go") || !loaded.Excludes("generated/deep/model.go") {
		t.Fatal("expected generated paths to be excluded")
	}
	if loaded.Excludes("internal/model.go") {
		t.Fatal("ordinary source was excluded")
	}
	content, err := os.ReadFile(filepath.Join(root, FileName))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(content), "# yaml-language-server: $schema="+SchemaURL) {
		t.Fatal("policy has no schema directive")
	}
}

func TestInstallPreservesExistingPolicy(t *testing.T) {
	root := t.TempDir()
	first, created, err := Install(root)
	if err != nil || !created {
		t.Fatalf("first install = %#v, %v, %v", first, created, err)
	}
	second, created, err := Install(root)
	if err != nil || created || second.APIVersion != APIVersion {
		t.Fatalf("second install = %#v, %v, %v", second, created, err)
	}
}

func TestPolicyRejectsBypassesAndUnknownValues(t *testing.T) {
	cases := map[string]func(*Policy){
		"non-strict mode":   func(value *Policy) { value.Spec.Comments.Mode = "off" },
		"escaping pattern":  func(value *Policy) { value.Spec.Comments.VendoredPaths = []string{"../vendor/**"} },
		"absolute pattern":  func(value *Policy) { value.Spec.Comments.VendoredPaths = []string{"/vendor/**"} },
		"unknown intrinsic": func(value *Policy) { value.Spec.Comments.Intrinsic = []Intrinsic{"anything"} },
		"unknown adapter":   func(value *Policy) { value.Spec.Agents.Adapters = []Adapter{"anything"} },
		"duplicate adapter": func(value *Policy) { value.Spec.Agents.Adapters = []Adapter{AdapterAgents, AdapterAgents} },
		"wrong api version": func(value *Policy) { value.APIVersion = "koment.dev/v1" },
		"wrong kind":        func(value *Policy) { value.Kind = "Configuration" },
		"unknown principle": func(value *Policy) { value.Spec.Agents.Principles = []Principle{"anything"} },
		"duplicate principle": func(value *Policy) {
			value.Spec.Agents.Principles = []Principle{PrincipleBackCompatEvidence, PrincipleBackCompatEvidence}
		},
		"empty annotation pattern": func(value *Policy) {
			value.Spec.Comments.AllowedAnnotations = []string{""}
		},
		"invalid annotation regexp": func(value *Policy) {
			value.Spec.Comments.AllowedAnnotations = []string{"("}
		},
	}
	for name, corrupt := range cases {
		t.Run(name, func(t *testing.T) {
			configured := Default()
			corrupt(&configured)
			if err := configured.Validate(); err == nil {
				t.Fatal("invalid policy accepted")
			}
		})
	}
}

func TestPolicyAllowedAnnotationsMatchAndPersist(t *testing.T) {
	root := t.TempDir()
	configured := Default()
	configured.Spec.Comments.AllowedAnnotations = []string{
		"# renovate[\\s:]",
		"Code generated by protoc.*DO NOT EDIT",
	}
	if err := Save(root, configured); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.MatchesAllowedAnnotation("# renovate: enable") {
		t.Error("renovate annotation was not matched")
	}
	if !loaded.MatchesAllowedAnnotation("Code generated by protoc-gen-go. DO NOT EDIT.") {
		t.Error("protoc annotation was not matched")
	}
	if loaded.MatchesAllowedAnnotation("# ordinary comment") {
		t.Error("ordinary comment matched an allowed annotation")
	}
}

func TestLoadRefusesAFlatPolicyAndLeavesItUntouched(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".koment"), 0o755); err != nil {
		t.Fatal(err)
	}
	legacy := `version: 1
comments:
  mode: strict
  intrinsic:
    - toolchain-directive
  generated_paths:
    - '**/*.gen.go'
  vendored_paths:
    - vendor/**
agents:
  adapters:
    - agents
`
	path := filepath.Join(root, FileName)
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(root)
	if err == nil {
		t.Fatal("a v1 policy was read by a binary that no longer migrates it")
	}
	for _, want := range []string{"version: 1", "koment 2.x", APIVersion, "ADR 0130"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not mention %q: %v", want, err)
		}
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != legacy {
		t.Errorf("reading a repository rewrote a policy it refused:\n%s", after)
	}
}

func TestLoadRefusesAPolicyFromAnUnknownGeneration(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".koment"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, FileName), []byte("apiVersion: koment.dev/v9\nkind: Policy\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(root)
	if err == nil || !strings.Contains(err.Error(), APIVersion) {
		t.Fatalf("err = %v", err)
	}
}

func TestOnlyEnabledPrinciplesAreStated(t *testing.T) {
	configured := Default()
	if len(configured.States()) != 1 {
		t.Fatalf("the default policy states %d principles", len(configured.States()))
	}
	configured.Spec.Agents.Principles = nil
	if stated := configured.States(); len(stated) != 0 {
		t.Fatalf("a policy with no principles stated %v", stated)
	}
}
