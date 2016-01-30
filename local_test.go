package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testLocalShell *Local = nil
)

func newTestLocalShell() *Local {
	local, err := NewLocal(LocalConfig{})
	if err != nil {
		panic(err)
	}

	return &local
}

type testLocalState struct {
	args  []string
	shell Local
}

func (state *testLocalState) handler(kind int, result string) error {
	if kind == Stdout {
		result = "OUT: " + result
	} else if kind == Stderr {
		result = "ERR: " + result
	}

	state.args = append(state.args, result)
	return nil
}

func newTestLocalState() testLocalState {
	if testLocalShell == nil {
		testLocalShell = newTestLocalShell()
	}

	state := testLocalState{shell: *testLocalShell}
	return state
}

func TestLocalRunsCommand(test *testing.T) {
	state := newTestLocalState()

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
	status, err := state.shell.Run("echo -n TEST 1>&2 && false", state.handler)
	assert.NoError(test, err)
	assert.Equal(test, 1, status)
	assert.Equal(test, "ERR: TEST", state.args[0])
}

func TestLocalExitsWithoutError(test *testing.T) {
	shell, err := NewLocal(LocalConfig{})
	assert.NoError(test, err)
	err = shell.Close()
	assert.NoError(test, err)
}

func TestLocalExitsWithoutErrorWhileExecutingCommand(test *testing.T) {
	shell, err := NewLocal(LocalConfig{})
	assert.NoError(test, err)
	go func() { shell.Run("sleep 100", nil) }()
	err = shell.Close()
	assert.NoError(test, err)
}
