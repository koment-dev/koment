package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/koment-dev/koment/internal/agentpolicy"
	"github.com/koment-dev/koment/internal/anchor"
	"github.com/koment-dev/koment/internal/policy"
)

func runAgents(args []string, env Environment) int {
	if len(args) == 0 {
		return misuse(env, "agents needs install, check, or hook")
	}
	switch args[0] {
	case "install":
		return runAgentsInstall(args[1:], env)
	case "check":
		return runAgentsCheck(args[1:], env)
	case "hook":
		return runAgentsHook(args[1:], env)
	default:
		return misuse(env, "unknown agents command %q; want install, check, or hook", args[0])
	}
}

func runAgentsInstall(args []string, env Environment) int {
	flags := flagSet("agents install", env)
	if code, ok := parse(flags, args); !ok {
		return code
	}
	if flags.NArg() != 0 {
		return misuse(env, "agents install takes no arguments")
	}
	annotations, err := openStore()
	if err != nil {
		return fail(env, err)
	}
	configured, policyCreated, err := policy.Install(annotations.Root())
	if err != nil {
		return fail(env, err)
	}
	if policyCreated {
		fmt.Fprintf(env.Stdout, "installed %s\n", policy.FileName)
	}
	changes, err := agentpolicy.Install(annotations.Root(), configured)
	if err != nil {
		return fail(env, err)
	}
	for _, change := range changes {
		fmt.Fprintf(env.Stdout, "%s %s\n", change.Action, change.Path)
	}
	if !policyCreated && len(changes) == 0 {
		fmt.Fprintln(env.Stdout, "agent policy: already current")
	}
	return ExitOK
}

func runAgentsCheck(args []string, env Environment) int {
	flags := flagSet("agents check", env)
	if code, ok := parse(flags, args); !ok {
		return code
	}
	if flags.NArg() != 0 {
		return misuse(env, "agents check takes no arguments")
	}
	active, err := openActiveRepository()
	if err != nil {
		return fail(env, err)
	}
	if active == nil {
		return ExitOK
	}
	drift, err := agentpolicy.Check(active.annotations.Root(), active.configured)
	if err != nil {
		return fail(env, err)
	}
	if len(drift) == 0 {
		fmt.Fprintln(env.Stdout, "agent policy: ok")
		return ExitOK
	}
	for _, entry := range drift {
		fmt.Fprintf(env.Stdout, "%s: %s\n", entry.Path, entry.Reason)
	}
	fmt.Fprintf(env.Stderr, "koment: %d agent policy adapters need `koment agents install`\n", len(drift))
	return ExitFailure
}

func runAgentsHook(args []string, env Environment) int {
	if len(args) != 1 {
		return misuse(env, "agents hook needs pre-tool or stop")
	}
	input, err := io.ReadAll(env.Stdin)
	if err != nil {
		return fail(env, fmt.Errorf("reading hook input: %w", err))
	}
	switch args[0] {
	case "pre-tool":
		output, err := agentpolicy.PreToolOutput(input)
		if err != nil {
			return fail(env, err)
		}
		if _, err := env.Stdout.Write(output); err != nil {
			return fail(env, fmt.Errorf("writing hook output: %w", err))
		}
		return ExitOK
	case "stop":
		return runAgentsStopHook(input, env)
	default:
		return misuse(env, "unknown agents hook %q; want pre-tool or stop", args[0])
	}
}

func runAgentsStopHook(input []byte, env Environment) int {
	reason, active := completionFailure()
	if !active {
		return writeStopResponse(env, map[string]any{})
	}
	continued, err := agentpolicy.StopWasContinued(input)
	if err != nil {
		return fail(env, err)
	}
	response := map[string]any{}
	if reason != "" {
		if continued {
			response["continue"] = false
			response["stopReason"] = reason
			response["systemMessage"] = reason
		} else {
			response["decision"] = "block"
			response["reason"] = reason
		}
	}
	return writeStopResponse(env, response)
}

func writeStopResponse(env Environment, response map[string]any) int {
	if err := json.NewEncoder(env.Stdout).Encode(response); err != nil {
		return fail(env, fmt.Errorf("writing Stop output: %w", err))
	}
	return ExitOK
}

func completionFailure() (string, bool) {
	active, err := openActiveRepository()
	if err != nil {
		return "koment could not activate the repository: " + err.Error(), true
	}
	if active == nil {
		return "", false
	}
	snapshot, err := active.service.Snapshot()
	if err != nil {
		return "koment could not resolve annotations: " + err.Error(), true
	}
	counts := snapshot.Counts()
	annotationFailures := counts[anchor.StatusAmbiguous] + counts[anchor.StatusDrifted] + counts[anchor.StatusOrphaned]
	violations, err := active.service.CheckComments(nil)
	if err != nil {
		return "koment could not check comments: " + err.Error(), true
	}
	drift, err := agentpolicy.Check(active.annotations.Root(), active.configured)
	if err != nil {
		return "koment could not check agent adapters: " + err.Error(), true
	}
	var failures []string
	if annotationFailures > 0 {
		failures = append(failures, fmt.Sprintf("%d annotations are unresolved; run `koment check`", annotationFailures))
	}
	if len(violations) > 0 {
		failures = append(failures, fmt.Sprintf("%d inline comments violate policy; run `koment comments check` and convert or acknowledge them", len(violations)))
	}
	if len(drift) > 0 {
		failures = append(failures, fmt.Sprintf("%d agent adapters drifted; run `koment agents install`", len(drift)))
	}
	return strings.Join(failures, "; "), true
}
