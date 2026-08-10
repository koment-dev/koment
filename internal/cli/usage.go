package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

type command struct {
	name    string
	args    string
	summary string
}

type section struct {
	title    string
	commands []command
}

var sections = []section{
	{
		title: "Getting started",
		commands: []command{
			{"bootstrap", "[--agents <list>] [--all] [--policy-only] [--non-interactive]",
				"set koment up in this repository"},
		},
	},
	{
		title: "Annotations",
		commands: []command{
			{"add", "<file> [--excerpt <text>] --kind <kind> --body <text|->",
				"record why the code is the way it is"},
			{"show", "<file>", "print the annotations on one file"},
			{"check", "[path...]", "resolve every anchor and fail on the stale ones"},
			{"list", "[--kind <kind>]", "list annotations across the repository"},
			{"search", "<query>", "find recorded rationale by topic"},
			{"reanchor", "<id> [--excerpt <text>] [--file <path>]",
				"point an annotation at code that moved"},
			{"edit", "<id> [--title <text>] [--body <text|->]",
				"rewrite an annotation's headline or rationale"},
			{"forget", "<id>", "delete an annotation; git keeps who removed it"},
		},
	},
	{
		title: "Policy",
		commands: []command{
			{"comments", "check|convert|acknowledge", "keep rationale out of source comments"},
			{"agents", "install|check", "generate and verify the agent adapters"},
		},
	},
	{
		title: "Read and share",
		commands: []command{
			{"ui", "[--listen <addr>] [--write]", "browse this repository in a browser"},
			{"site", "--out <dir>", "render one repository to static HTML"},
			{"serve", "--config <repositories.yaml>", "serve several repositories at once"},
		},
	},
	{
		title: "Integrations",
		commands: []command{
			{"mcp", "[--write | --http <addr> | --streamable-http <addr>]",
				"Model Context Protocol server for agents"},
			{"lsp", "", "editor protocol over stdio"},
		},
	},
	{
		commands: []command{
			{"version", "", "print the version, commit and toolchain"},
		},
	},
}

var subcommands = []command{
	{"comments check", "[path...]", "fail on comments the policy does not allow"},
	{"comments convert", "<file> --excerpt <comment> [--kind <kind>]",
		"turn an existing comment into an annotation"},
	{"comments acknowledge", "<file> --excerpt <comment> --body <text|-> --acknowledge-inline-comment",
		"keep one comment on the record, with a reason"},
	{"agents install", "[--agents <list>]", "write or refresh the generated agent adapters"},
	{"agents check", "", "fail when an adapter is missing or stale"},
}

const epilogue = `check exits non-zero when an annotation is ambiguous, drifted or orphaned.
reanchor is how you fix one while keeping its id.

Run "koment <command> --help" for the flags one command takes.`

func writeUsage(out io.Writer) {
	fmt.Fprint(out, "koment — out-of-band code annotations\n\n")
	fmt.Fprint(out, "Usage:\n  koment <command> [flags]\n")

	width := widestName()
	for _, group := range sections {
		fmt.Fprintln(out)
		if group.title != "" {
			fmt.Fprintf(out, "%s\n", group.title)
		}
		for _, entry := range group.commands {
			fmt.Fprintf(out, "  %-*s  %s\n", width, entry.name, entry.summary)
		}
	}
	fmt.Fprintf(out, "\n%s\n", epilogue)
}

func widestName() int {
	widest := 0
	for _, group := range sections {
		for _, entry := range group.commands {
			if len(entry.name) > widest {
				widest = len(entry.name)
			}
		}
	}
	return widest
}

func lookup(name string) (command, bool) {
	for _, group := range sections {
		for _, entry := range group.commands {
			if entry.name == name {
				return entry, true
			}
		}
	}
	for _, entry := range subcommands {
		if entry.name == name {
			return entry, true
		}
	}
	return command{}, false
}

func writeCommandUsage(out io.Writer, name string, flags *flag.FlagSet) {
	entry, known := lookup(name)
	if !known {
		fmt.Fprintf(out, "Usage:\n  koment %s\n", name)
		flags.PrintDefaults()
		return
	}

	invocation := strings.TrimSpace("koment " + entry.name + " " + entry.args)
	fmt.Fprintf(out, "%s\n\nUsage:\n  %s\n", entry.summary, invocation)
	if hasFlags(flags) {
		fmt.Fprint(out, "\nFlags:\n")
		flags.PrintDefaults()
	}
}

func hasFlags(flags *flag.FlagSet) bool {
	declared := false
	flags.VisitAll(func(*flag.Flag) { declared = true })
	return declared
}
