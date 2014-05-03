package fake_command_runner_matchers

import (
	"fmt"
	"os/exec"

	"github.com/cloudfoundry/gunk/command_runner/fake_command_runner"
)

func prettySpecs(specs []fake_command_runner.CommandSpec) string {
	out := ""

	for _, spec := range specs {
		out += prettySpec(spec)
	}

	return out
}

func prettyCommands(commands []*exec.Cmd) string {
	out := ""

	for _, command := range commands {
		out += fmt.Sprintf(`
	'%s'
		with arguments %v
		and environment %v
		in directory '%s'`, command.Path, command.Args, command.Env, command.Dir)
	}

	return out
}

func prettySpec(spec fake_command_runner.CommandSpec) string {
	return fmt.Sprintf(`
	'%s'
		with arguments %v
		and environment %v
		and input '%s'
		in directory '%s'`, spec.Path, spec.Args, spec.Env, spec.Stdin, spec.Dir)
}
