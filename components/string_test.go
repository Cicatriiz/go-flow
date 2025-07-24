package components

import (
	"testing"

	"github.com/forrest/go-flow/core"
)

func TestUpperCase(t *testing.T) {
	c := NewUpperCase()
	inputs := map[string]interface{}{
		"input": "hello",
	}
	expectedOutputs := map[string]interface{}{
		"output": "HELLO",
	}
	core.TestComponent(t, c, inputs, expectedOutputs)
}
