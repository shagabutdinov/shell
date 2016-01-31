package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testLocalState struct {
	args  []string
	shell *Local
}

func (state *testLocalState) handler(kind MessageType, result string) error {
	if kind == StdOut {
		result = "OUT: " + result
	} else if kind == StdErr {
		result = "ERR: " + result
	}

	state.args = append(state.args, result)
	return nil
}

func newTestLocalState() testLocalState {
	local, err := NewLocal(LocalConfig{})
	if err != nil {
		panic(err)
	}

	return testLocalState{shell: local}
}

func TestLocalRunsCommand(test *testing.T) {
	state := newTestLocalState()
	defer state.shell.Close()

	status, err := state.shell.Run("cd /var/lib", state.handler)
	assert.NoError(test, err)
	assert.Equal(test, 0, status)

	status, err = state.shell.Run("echo `pwd`", state.handler)
	assert.NoError(test, err)
	assert.Equal(test, 0, status)
	assert.Equal(test, "OUT: /var/lib", state.args[0])
}

func TestLocalRunsErrorCommand(test *testing.T) {
	state := newTestLocalState()
	defer state.shell.Close()

	status, err := state.shell.Run("echo -n TEST 1>&2 && false", state.handler)
	assert.NoError(test, err)
	assert.Equal(test, 1, status)
	assert.Equal(test, "ERR: TEST", state.args[0])
}

func TestLocalExitsWithoutError(test *testing.T) {
	shell, err := NewLocal(LocalConfig{})
	defer shell.Close()

	assert.NoError(test, err)
	err = shell.Close()
	assert.NoError(test, err)
}

func TestLocalExitsWithoutErrorWhileExecutingCommand(test *testing.T) {
	shell, err := NewLocal(LocalConfig{})
	defer shell.Close()

	assert.NoError(test, err)
	go func() { shell.Run("sleep 100", nil) }()
	err = shell.Close()
	assert.NoError(test, err)
}
