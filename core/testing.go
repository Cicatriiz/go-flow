package core

import (
	"context"
	"testing"
)

// TestComponent is a helper function for testing components.
// It takes a component, a map of input data, and a map of expected output data.
// It runs the component's Process method and compares the actual output to the expected output.
func TestComponent(t *testing.T, c Component, inputs map[string]interface{}, expectedOutputs map[string]interface{}) {
	t.Helper()

	outputs, err := c.Process(context.Background(), inputs)
	if err != nil {
		t.Fatalf("Process() returned an unexpected error: %v", err)
	}

	if len(outputs) != len(expectedOutputs) {
		t.Fatalf("Process() returned %d outputs, but expected %d", len(outputs), len(expectedOutputs))
	}

	for name, expected := range expectedOutputs {
		actual, ok := outputs[name]
		if !ok {
			t.Errorf("Process() did not return an output named '%s'", name)
			continue
		}
		if actual != expected {
			t.Errorf("Process() returned an incorrect value for output '%s': got %v, want %v", name, actual, expected)
		}
	}
}
