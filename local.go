package shell

import "os/exec"

type LocalConfig struct {
	LineLimit int
}

func NewLocal(config LocalConfig) (Local, error) {
	shell := Local{command: exec.Command("/bin/sh")}
	shell.limit = config.LineLimit
	shell.messages = make(chan message, 4096)

	var err error

	shell.stdin, err = shell.command.StdinPipe()
	if err != nil {
		return shell, err
	}

	shell.stdout, err = shell.command.StdoutPipe()
	if err != nil {
		return shell, err
	}

	shell.stderr, err = shell.command.StderrPipe()
	if err != nil {
		return shell, err
	}

	err = shell.command.Start()
	if err != nil {
		return shell, err
	}

	shell.start()

	return shell, nil
}

type Local struct {
	shell
	command *exec.Cmd
}

func (shell Local) Close() error {
	return shell.close()
}
