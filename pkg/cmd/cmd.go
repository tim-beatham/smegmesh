// cmd is a package for running commands in the different operating systems implementations
package cmd

import (
	"os/exec"
	"strings"
)

// CmdRunner: run cmd commands when instantiating a network
type CmdRunner interface {
	RunCommands(commands ...string) error
}

// UnixCmdRunner: Run UNIX commands
type UnixCmdRunner struct{}

// RunCommand: runs the unix command. It splits the command into fields
// and then runs the command accordingly
func RunCommand(cmd string) error {
	args := strings.Fields(cmd)
	c := exec.Command(args[0], args[1:]...)
	return c.Run()
}

// RunCommands: run a series of commands
func (l *UnixCmdRunner) RunCommands(commands ...string) error {
	for _, cmd := range commands {
		err := RunCommand(cmd)

		if err != nil {
			return err
		}
	}

	return nil
}
