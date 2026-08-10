package cli

import (
	"fmt"

	"github.com/koment-dev/koment/internal/store"
)

func runForget(args []string, env Environment) int {
	flags := flagSet("forget", env)
	id, code, ok := onePositional("forget", "an annotation id", flags, args, env)
	if !ok {
		return code
	}

	service, _, err := openApplication()
	if err != nil {
		return fail(env, err)
	}
	removed, err := service.Forget(id)
	if err != nil {
		return fail(env, err)
	}
	fmt.Fprintf(env.Stdout, "forgot %s  %s %s\n  %s\n  git restores it: git checkout -- %s\n",
		removed.Metadata.ID, removed.Spec.Type,
		location(removed.Spec.Target.File, removed.Status.LastSeenLine),
		removed.Headline(), recordFile(removed.Metadata.ID))
	return ExitOK
}

func recordFile(id string) string {
	return store.DirName + "/annotations/" + id + ".yaml"
}
