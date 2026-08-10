package lsp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/koment-dev/koment/internal/store"
)

func workspaceWith(t *testing.T, name, body string) workspaceFile {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, store.DirName), 0o755); err != nil {
		t.Fatal(err)
	}
	policyYAML := "apiVersion: koment.dev/v1alpha\nkind: Policy\nspec:\n  comments:\n    mode: strict\n"
	if err := os.WriteFile(filepath.Join(root, store.DirName, "policy.yaml"), []byte(policyYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	file, err := loadWorkspaceFile("file://"+path, []byte(body))
	if err != nil {
		t.Fatal(err)
	}
	return file
}

func TestCommentDiagnosticsReachNonGoFiles(t *testing.T) {
	for name, body := range map[string]string{
		"app.yaml": "# Retried five times because the upstream is flaky.\nretries: 5\n",
		"main.py":  "# Cached because recomputing is expensive.\nvalue = 1\n",
		"lib.rs":   "// Kept unsafe because the borrow checker cannot prove this.\nfn main() {}\n",
	} {
		t.Run(name, func(t *testing.T) {
			found, err := documentDiagnostics(workspaceWith(t, name, body))
			if err != nil {
				t.Fatal(err)
			}
			for _, problem := range found {
				if problem.Code == "koment.comment" {
					return
				}
			}
			t.Errorf("no comment diagnostic for %s, so the editor offers no quick fix: %+v", name, found)
		})
	}
}

func TestCommentDiagnosticsSkipProseFiles(t *testing.T) {
	found, err := documentDiagnostics(workspaceWith(t, "README.md", "# A heading, not a comment\n\nProse.\n"))
	if err != nil {
		t.Fatal(err)
	}
	for _, problem := range found {
		if problem.Code == "koment.comment" {
			t.Errorf("a Markdown heading was reported as a prohibited comment: %+v", problem)
		}
	}
}

func TestCommentDiagnosticCarriesTheTextAQuickFixNeeds(t *testing.T) {
	found, err := documentDiagnostics(workspaceWith(t, "app.yaml",
		"# Retried five times because the upstream is flaky.\nretries: 5\n"))
	if err != nil {
		t.Fatal(err)
	}
	for _, problem := range found {
		if problem.Code != "koment.comment" {
			continue
		}
		data, ok := problem.Data.(map[string]any)
		if !ok {
			t.Fatalf("diagnostic carries no data, so convert and acknowledge cannot act: %+v", problem)
		}
		if comment, _ := data["comment"].(string); !strings.Contains(comment, "upstream is flaky") {
			t.Errorf("the diagnostic does not carry the comment text: %q", comment)
		}
		return
	}
	t.Fatal("no comment diagnostic produced")
}
