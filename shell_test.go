package shell

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testShellResult struct {
	status int
	err    error
}

func newTestShellState(limit int) testShellState {
	shell := &shell{messages: make(chan message, 4096), limit: limit}
	state := testShellState{result: make(chan testShellResult, 1024)}

	state.stdin, shell.stdin = io.Pipe()
	shell.stdout, state.stdout = io.Pipe()
	shell.stderr, state.stderr = io.Pipe()

	state.shell = shell
	shell.start()

	return state
}

type testShellState struct {
	stdin  io.ReadCloser
	stdout io.WriteCloser
	stderr io.WriteCloser

	err    error
	args   []string
	result chan testShellResult

	shell *shell
}

func (state *testShellState) handler(kind MessageType, result string) error {
	if kind == StdOut {
		result = "OUT: " + result
	} else if kind == StdErr {
		result = "ERR: " + result
	}

	state.args = append(state.args, result)
	return state.err
}

func (state *testShellState) run(
	command string,
	callback func(),
) (int, error) {
	go func() {
		status, err := state.shell.Run(command, state.handler)
		state.result <- testShellResult{status, err}
	}()

	go func() {
		for {
			bytes := make([]byte, 1024)
			_, err := state.stdin.Read(bytes)
			if err != nil {
				break
			}
		}

	}()

	callback()

	result := <-state.result
	return result.status, result.err
}

func TestShellSendsCommandToStdin(test *testing.T) {
	result := make([]byte, 1024)
	length := 0
	state := newTestShellState(0)
	defer state.shell.close()
	state.run("COMMAND\n\n", func() {
		length, _ = state.stdin.Read(result)
		state.stdout.Write([]byte("__SHELL_EXIT_STATUS_0__"))
		state.stderr.Write([]byte("__SHELL_EXIT_STATUS_0__"))
	})

	expected := "COMMAND\necho -n __SHELL_EXIT_STATUS_$?__ | tee /dev/stderr\n"
	assert.Equal(test, expected, string(result[:length]))
}

func TestShellClosesNormally(test *testing.T) {
	state := newTestShellState(0)
	state.run("COMMAND\n\n", func() {
		state.stdout.Write([]byte("__SHELL_EXIT_STATUS_0__"))
		state.stderr.Write([]byte("__SHELL_EXIT_STATUS_0__"))
	})

	err := state.shell.close()
	assert.NoError(test, err)
}

func TestShellReturnsNoErrorOnRunningCommand(test *testing.T) {
	state := newTestShellState(0)
	defer state.shell.close()
	_, err := state.run("COMMAND\n\n", func() {
		state.stdout.Write([]byte("__SHELL_EXIT_STATUS_0__"))
		state.stderr.Write([]byte("__SHELL_EXIT_STATUS_0__"))
	})

	assert.NoError(test, err)
}

func TestShellReturnsExitStatus(test *testing.T) {
	state := newTestShellState(0)
	defer state.shell.close()
	status, _ := state.run("COMMAND\n\n", func() {
		state.stdout.Write([]byte("__SHELL_EXIT_STATUS_1__"))
		state.stderr.Write([]byte("__SHELL_EXIT_STATUS_1__"))
	})

	assert.Equal(test, 1, status)
}

func TestShellSendsStdErrToHandler(test *testing.T) {
	state := newTestShellState(0)
	defer state.shell.close()
	state.run("COMMAND", func() {
		state.stderr.Write([]byte("MSG"))
		state.stdout.Write([]byte("__SHELL_EXIT_STATUS_0__"))
		state.stderr.Write([]byte("__SHELL_EXIT_STATUS_0__"))
	})

	assert.Equal(test, "ERR: MSG", state.args[0])
}

func TestShellSendsStdOutToHandler(test *testing.T) {
	state := newTestShellState(0)
	defer state.shell.close()
	state.run("COMMAND", func() {
		state.stdout.Write([]byte("MSG"))
		state.stdout.Write([]byte("__SHELL_EXIT_STATUS_0__"))
		state.stderr.Write([]byte("__SHELL_EXIT_STATUS_0__"))
	})

	assert.Equal(test, "OUT: MSG", state.args[0])
}

func TestShellConcatsStdOutMessages(test *testing.T) {
	state := newTestShellState(0)
	defer state.shell.close()
	state.run("COMMAND", func() {
		state.stdout.Write([]byte("MSG1"))
		state.stdout.Write([]byte("MSG2"))
		state.stdout.Write([]byte("__SHELL_EXIT_STATUS_0__"))
		state.stderr.Write([]byte("__SHELL_EXIT_STATUS_0__"))
	})

	assert.Equal(test, "OUT: MSG1MSG2", state.args[0])
}

func TestShellSendsTwoStdOutMessagesToHandler(test *testing.T) {
	state := newTestShellState(0)
	defer state.shell.close()
	state.run("COMMAND", func() {
		state.stdout.Write([]byte("MSG1\nMSG2\n__SHELL_EXIT_STATUS_0__"))
		state.stderr.Write([]byte("__SHELL_EXIT_STATUS_0__"))
	})

	assert.Equal(test, "OUT: MSG2", state.args[1])
}

func TestShellDetectsStatusFromTwoDifferentMessages(test *testing.T) {
	state := newTestShellState(0)
	defer state.shell.close()
	status, _ := state.run("COMMAND", func() {
		state.stdout.Write([]byte("__SHELL_EXIT_STATUS_"))
		state.stdout.Write([]byte("1__"))
		state.stderr.Write([]byte("__SHELL_EXIT_"))
		state.stderr.Write([]byte("STATUS_1__"))
	})

	assert.Equal(test, 1, status)
}

func TestShellIgnoresStuffInStdOutAfterComplete(test *testing.T) {
	state := newTestShellState(0)
	defer state.shell.close()
	state.run("COMMAND", func() {
		state.stdout.Write([]byte("__SHELL_EXIT_STATUS_1__"))
		state.stdout.Write([]byte("TEST\n"))
		state.stderr.Write([]byte("__SHELL_EXIT_STATUS_1__"))
	})

	assert.Equal(test, 0, len(state.args))
}

func TestShellIgnoresStuffInStdErrAfterComplete(test *testing.T) {
	state := newTestShellState(0)
	defer state.shell.close()
	state.run("COMMAND", func() {
		state.stdout.Write([]byte("__SHELL_EXIT_STATUS_1__"))
		state.stderr.Write([]byte("__SHELL_EXIT_STATUS_1__"))
		state.stderr.Write([]byte("TEST\n"))
	})

	assert.Equal(test, 0, len(state.args))
}

func TestShellLimitsEachLineToLimit(test *testing.T) {
	state := newTestShellState(len("MESSAGE1"))
	defer state.shell.close()
	state.run("COMMAND", func() {
		state.stdout.Write([]byte("MESSAGE1"))
		state.stdout.Write([]byte("MESSAGE2"))
		state.stdout.Write([]byte("__SHELL_EXIT_STATUS_1__"))
		state.stderr.Write([]byte("__SHELL_EXIT_STATUS_1__"))
	})

	assert.Equal(test, "OUT: MESSAGE2", state.args[0])
}

func TestShellRunsTwoCommands(test *testing.T) {
	state := newTestShellState(len("MESSAGE1"))
	defer state.shell.close()
	state.run("COMMAND1", func() {
		state.stdout.Write([]byte("MESSAGE1"))
		state.stdout.Write([]byte("__SHELL_EXIT_STATUS_1__"))
		state.stderr.Write([]byte("__SHELL_EXIT_STATUS_1__"))
	})

	state.run("COMMAND2", func() {
		state.stdout.Write([]byte("MESSAGE2"))
		state.stdout.Write([]byte("__SHELL_EXIT_STATUS_1__"))
		state.stderr.Write([]byte("__SHELL_EXIT_STATUS_1__"))
	})

	assert.Equal(test, 2, len(state.args))
}

func TestShellReceivesStdErrAfterStdOut(test *testing.T) {
	state := newTestShellState(len("MESSAGE1"))
	defer state.shell.close()
	state.run("COMMAND1", func() {
		state.stdout.Write([]byte("MESSAGE1"))
		state.stdout.Write([]byte("__SHELL_EXIT_STATUS_1__"))
		state.stderr.Write([]byte("__SHELL_EXIT_STATUS_1__"))
	})

	state.run("COMMAND2", func() {
		state.stderr.Write([]byte("MESSAGE2"))
		state.stdout.Write([]byte("__SHELL_EXIT_STATUS_1__"))
		state.stderr.Write([]byte("__SHELL_EXIT_STATUS_1__"))
	})

	assert.Equal(test, "ERR: MESSAGE2", state.args[1])
}

func TestShellReturnsErrorOnStdOutError(test *testing.T) {
	state := newTestShellState(len("MESSAGE1"))
	defer state.shell.close()
	_, err := state.run("COMMAND", func() {
		state.stdout.(*io.PipeWriter).CloseWithError(errors.New("ERROR"))
	})

	assert.Error(test, err)
}

func TestShellReturnsErrorOnStdErrError(test *testing.T) {
	state := newTestShellState(len("MESSAGE1"))
	defer state.shell.close()
	_, err := state.run("COMMAND", func() {
		state.stderr.(*io.PipeWriter).CloseWithError(errors.New("ERROR"))
	})

	assert.Error(test, err)
}

func TestShellReturnsErrorOnUnknownStatus(test *testing.T) {
	state := newTestShellState(len("MESSAGE1"))
	defer state.shell.close()
	_, err := state.run("COMMAND", func() {
		state.stdout.Write([]byte("__SHELL_EXIT_STATUS_WRONG__"))
		state.stderr.Write([]byte("__SHELL_EXIT_STATUS_WRONG__"))
	})

	assert.Error(test, err)
}

func TestShellReturnsErrorIfHandlerReturnsError(test *testing.T) {
	state := newTestShellState(len("MESSAGE1"))
	defer state.shell.close()
	state.err = errors.New("TEST")
	_, err := state.run("COMMAND", func() {
		state.stdout.Write([]byte("TEST"))
		state.stdout.Write([]byte("__SHELL_EXIT_STATUS_1__"))
		state.stderr.Write([]byte("__SHELL_EXIT_STATUS_1__"))
	})

	assert.Error(test, err)
}
