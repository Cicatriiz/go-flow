package components

import (
	"testing"

	"github.com/forrest/go-flow/core"
)

func TestToUpper(t *testing.T) {
	c := NewToUpper()
	inputs := map[string]interface{}{
		"input": "hello",
	}
	expectedOutputs := map[string]interface{}{
		"output": "HELLO",
	}
	core.TestComponent(t, c, inputs, expectedOutputs)
}
