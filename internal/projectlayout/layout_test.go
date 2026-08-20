package projectlayout

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryMatchesLayout(t *testing.T) {
	root, err := RepositoryRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	if err := Check(root); err != nil {
		t.Fatal(err)
	}
}

func TestUnknownRootFileIsRejected(t *testing.T) {
	violations := ValidatePaths([]string{"notes.txt"})
	if len(violations) != 1 || !strings.Contains(violations[0], "root file") {
		t.Fatalf("unexpected violations: %v", violations)
	}
}

func TestUnknownClosedChildIsRejected(t *testing.T) {
	violations := ValidatePaths([]string{"integrations/editors/new-editor/manifest.json"})
	if len(violations) != 1 || !strings.Contains(violations[0], "new-editor is not allowed") {
		t.Fatalf("unexpected violations: %v", violations)
	}
}

func TestReferenceIsCurrent(t *testing.T) {
	root, err := RepositoryRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(root, ReferencePath))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != string(Reference()) {
		t.Fatal("repository layout reference is stale")
	}
}

func TestRetiredReferenceIsRejected(t *testing.T) {
	root := t.TempDir()
	retired := strings.Join([]string{"docs", "releasing.md"}, "/")
	if err := os.WriteFile(filepath.Join(root, "guide.md"), []byte(retired), 0o644); err != nil {
		t.Fatal(err)
	}
	repositoryRoot, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := repositoryRoot.Close(); err != nil {
			t.Error(err)
		}
	})
	violations, err := validateRetiredReferences(repositoryRoot, []string{"guide.md"})
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 1 || !strings.Contains(violations[0], "docs/guides/release-koment.md") {
		t.Fatalf("unexpected violations: %v", violations)
	}
}

func TestHistoricalAnnotationReferenceIsAllowed(t *testing.T) {
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := root.Close(); err != nil {
			t.Error(err)
		}
	})
	violations, err := validateRetiredReferences(root, []string{".koment/annotations/history.yaml"})
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Fatalf("unexpected violations: %v", violations)
	}
}
