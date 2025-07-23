package core

import (
	"context"
	"reflect"
)

// Component is the fundamental building block of a Go-Flow pipeline.
// It represents a single unit of processing in a data flow graph.
type Component interface {
	// Name returns the unique identifier of the component.
	Name() string
	// SetName sets the name of the component.
	SetName(name string)
	// InputPorts returns the list of input ports for the component.
	InputPorts() []Port
	// OutputPorts returns the list of output ports for the component.
	OutputPorts() []Port
	// Process executes the component's logic, transforming input data into output data.
	Process(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)
	// Validate checks if the component is configured correctly.
	Validate() error
}

// Port represents an input or output connection point for a component.
// It defines the name, type, and other properties of a data channel.
type Port interface {
	// Name returns the name of the port.
	Name() string
	// Type returns the data type of the port.
	Type() reflect.Type
	// Required indicates whether the port must be connected.
	Required() bool
	// Description provides a human-readable description of the port.
	Description() string
}

// ExecutionEngine defines the interface for a pipeline execution engine.
type ExecutionEngine interface {
	// Run executes the given pipeline.
	Run(ctx context.Context, p *Pipeline, inputs, outputs map[string]chan interface{}) error
	// Close gracefully shuts down the engine.
	Close() error
}
