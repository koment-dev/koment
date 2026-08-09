package cli

import (
	"flag"
	"fmt"

	"github.com/koment-dev/koment/internal/application"
)

func runEdit(args []string, env Environment) int {
	flags := flagSet("edit", env)
	title := flags.String("title", "", "replace the headline shown beside the code")
	body := flags.String("body", "", "replace the rationale; - reads it from stdin")

	id, code, ok := onePositional("edit", "an annotation id", flags, args, env)
	if !ok {
		return code
	}

	input := application.EditInput{ID: id}
	flags.Visit(func(given *flag.Flag) {
		switch given.Name {
		case "title":
			input.Title = title
		case "body":
			input.Body = body
		}
	})
	if input.Body != nil {
		text, err := bodyText(*body, env.Stdin)
		if err != nil {
			return fail(env, err)
		}
		input.Body = &text
	}

	service, _, err := openApplication()
	if err != nil {
		return fail(env, err)
	}
	mutation, err := service.Edit(input)
	if err != nil {
		return fail(env, err)
	}
	fmt.Fprintf(env.Stdout, "%s  %s %s  %s\n",
		mutation.Record.Metadata.ID, mutation.Record.Spec.Type,
		location(mutation.Record.Spec.Target.File, mutation.Record.Status.LastSeenLine), mutation.Path)
	return ExitOK
}
