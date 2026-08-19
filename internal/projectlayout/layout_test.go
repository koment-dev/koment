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
