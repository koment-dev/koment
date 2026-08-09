package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/koment-dev/koment/internal/store"
)

func annotate(t *testing.T) string {
	t.Helper()
	added := koment(t, "add", "main.go",
		"--excerpt", "\tserve()",
		"--kind", "why",
		"--title", "Short headline",
		"--body", "The original rationale.")
	if added.code != ExitOK {
		t.Fatalf("add exited %d: %s", added.code, added.output())
	}
	return strings.Fields(added.stdout)[0]
}

func TestEditReplacesTheTitleWithoutChangingIdentity(t *testing.T) {
	root := repository(t)
	id := annotate(t)

	edited := koment(t, "edit", id, "--title", "A better headline written later")
	if edited.code != ExitOK {
		t.Fatalf("edit exited %d: %s", edited.code, edited.output())
	}

	record, err := os.ReadFile(filepath.Join(root, store.DirName, "annotations", id+".yaml"))
	if err != nil {
		t.Fatal(err)
	}
	written := string(record)
	if !strings.Contains(written, "A better headline written later") {
		t.Errorf("the new title was not written:\n%s", written)
	}
	if !strings.Contains(written, id) {
		t.Errorf("editing changed the annotation id:\n%s", written)
	}
	if !strings.Contains(written, "The original rationale.") {
		t.Errorf("editing the title also changed the body:\n%s", written)
	}
}

func TestEditNeedsSomethingToChange(t *testing.T) {
	repository(t)
	id := annotate(t)

	got := koment(t, "edit", id)
	if got.code != ExitFailure {
		t.Fatalf("edit exited %d, want %d: %s", got.code, ExitFailure, got.output())
	}
	if !strings.Contains(got.stderr, "--title") {
		t.Errorf("the error does not say what is missing: %s", got.stderr)
	}
}

func TestForgetRemovesTheRecordAndSaysHowToRestoreIt(t *testing.T) {
	repository(t)
	id := annotate(t)

	got := koment(t, "forget", id)
	if got.code != ExitOK {
		t.Fatalf("forget exited %d: %s", got.code, got.output())
	}
	if !strings.Contains(got.stdout, "git checkout --") {
		t.Errorf("forget did not say how to get the record back:\n%s", got.stdout)
	}

	if checked := koment(t, "check"); strings.Contains(checked.stdout, id) {
		t.Errorf("the forgotten annotation still resolves:\n%s", checked.stdout)
	}
}

func TestForgetRefusesAnUnknownID(t *testing.T) {
	repository(t)

	got := koment(t, "forget", "01JQ8ZK3M4N5P6R7S8T9V0W1X2")
	if got.code == ExitOK {
		t.Errorf("forgetting an annotation that does not exist reported success")
	}
}
