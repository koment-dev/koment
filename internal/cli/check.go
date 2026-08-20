package cli

import (
	"fmt"

	"github.com/koment-dev/koment/internal/anchor"
)

func runCheck(args []string, env Environment) int {
	flags := flagSet("check", env)
	if code, ok := parse(flags, args); !ok {
		return code
	}

	active, err := openActiveRepository()
	if err != nil {
		return fail(env, err)
	}
	if active == nil {
		return ExitOK
	}
	resolved, err := resolveEverything(active.service, active.annotations, flags.Args())
	if err != nil {
		return fail(env, err)
	}

	for _, entry := range resolved {
		writeFailures(env, entry)
	}

	counted := tallyOf(resolved)
	fmt.Fprintf(env.Stdout, "%s across %s: %s\n",
		plural(counted.total(), "annotation"), plural(len(resolved), "file"), counted)

	if failures := counted.failures(); failures > 0 {
		subject, pronoun := "no longer resolves", "it"
		if failures > 1 {
			subject, pronoun = "no longer resolve", "them"
		}
		fmt.Fprintf(env.Stderr, "koment: %s %s; revisit %s or update the anchor\n",
			plural(failures, "annotation"), subject, pronoun)
		return ExitFailure
	}
	return ExitOK
}

func writeFailures(env Environment, entry fileResolutions) {
	failing := make([]anchor.Resolution, 0, len(entry.resolutions))
	for _, resolution := range entry.resolutions {
		if resolution.Status.IsFailure() {
			failing = append(failing, resolution)
		}
	}
	if len(failing) == 0 {
		return
	}

	fmt.Fprintf(env.Stdout, "%s\n", entry.file)
	for _, resolution := range failing {
		writeResolution(env.Stdout, entry.file, resolution)
	}
}
