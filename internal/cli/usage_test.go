package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runArgs(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run(args, Environment{Stdout: &stdout, Stderr: &stderr}, serversThatNeverListen())
	return code, stdout.String(), stderr.String()
}

func serversThatNeverListen() Servers {
	quiet := func([]string, io.Writer) error { return nil }
	return Servers{MCP: quiet, UI: quiet, Site: quiet, Serve: quiet, LSP: quiet}
}

func TestAskingForHelpSucceeds(t *testing.T) {
	for _, args := range [][]string{
		{"--help"}, {"-h"}, {"help"},
		{"add", "--help"},
		{"check", "--help"},
		{"reanchor", "--help"},
		{"comments", "convert", "--help"},
		{"agents", "check", "--help"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			code, stdout, stderr := runArgs(t, args...)
			if code != ExitOK {
				t.Errorf("asking for help exited %d, which breaks any script that asks", code)
			}
			if stdout+stderr == "" {
				t.Error("asking for help printed nothing")
			}
		})
	}
}

func TestMisuseIsDistinguishedFromHelp(t *testing.T) {
	for _, args := range [][]string{
		{},
		{"frobnicate"},
		{"add", "--nosuchflag"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			if code, _, _ := runArgs(t, args...); code != ExitUsage {
				t.Errorf("misuse exited %d, want %d", code, ExitUsage)
			}
		})
	}
}

func TestEveryListedCommandIsDispatchable(t *testing.T) {
	for _, group := range sections {
		for _, entry := range group.commands {
			t.Run(entry.name, func(t *testing.T) {
				_, _, stderr := runArgs(t, entry.name, "--help")
				if strings.Contains(stderr, "unknown command") {
					t.Errorf("help lists %q but nothing dispatches it", entry.name)
				}
			})
		}
	}
}

func TestEverySubcommandHasItsOwnUsage(t *testing.T) {
	for _, entry := range subcommands {
		if _, known := lookup(entry.name); !known {
			t.Errorf("%q is not reachable through lookup", entry.name)
		}
		if entry.summary == "" {
			t.Errorf("%q has no summary, so its --help would be blank", entry.name)
		}
	}
}

func TestSummariesLineUpInOneColumn(t *testing.T) {
	var listing bytes.Buffer
	writeUsage(&listing)

	listed := map[string]int{}
	for _, group := range sections {
		for _, entry := range group.commands {
			listed[entry.name] = 0
		}
	}

	column, rows := -1, 0
	for _, line := range strings.Split(listing.String(), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if _, isCommand := listed[fields[0]]; !isCommand || !strings.HasPrefix(line, "  "+fields[0]) {
			continue
		}
		rows++
		afterName := 2 + len(fields[0])
		gap := line[afterName:]
		at := afterName + len(gap) - len(strings.TrimLeft(gap, " "))
		if column == -1 {
			column = at
			continue
		}
		if at != column {
			t.Errorf("summary starts at column %d, not %d, so the listing is ragged:\n%s", at, column, line)
		}
	}
	if rows != len(listed) {
		t.Fatalf("matched %d command rows, want %d", rows, len(listed))
	}
}

func TestEveryCommandIsDocumented(t *testing.T) {
	reference, err := os.ReadFile(filepath.Join("..", "..", "docs", "cli.md"))
	if err != nil {
		t.Fatal(err)
	}
	documented := map[string]bool{}
	for _, line := range strings.Split(string(reference), "\n") {
		if heading, found := strings.CutPrefix(line, "## "); found {
			documented[strings.TrimSpace(heading)] = true
		}
	}

	for _, group := range sections {
		for _, entry := range group.commands {
			if !documented[entry.name] {
				t.Errorf("koment %s has no section in docs/cli.md (ADR 0137)", entry.name)
			}
		}
	}
}
