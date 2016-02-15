package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShellEscapeReturnsEscapedValue(test *testing.T) {
	actual := Escape("test")
	assert.Equal(test, "test", actual)
}

func TestShellEscapeReturnsEscapedEval(test *testing.T) {
	actual := Escape("test`eval`")
	assert.Equal(test, "test\\`eval\\`", actual)
}
