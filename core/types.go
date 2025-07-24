package core

import (
	"context"
	"reflect"
)

// Component is the fundamental building block of a Go-Flow pipeline.
// It represents a single unit of processing in a data flow graph.
type Component interface {
	// Core identification and configuration
	Name() string
	SetName(name string)
	
	// Port definitions with enhanced metadata
	InputPorts() []Port
	OutputPorts() []Port
	
	// Processing with context and error handling
	Process(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)
	
	// Validation and health checking
	Validate() error
	HealthCheck(ctx context.Context) error
	
	// Lifecycle management
	Initialize(ctx context.Context) error
	Cleanup(ctx context.Context) error
	
	// Metadata for documentation and tooling
	Description() string
	Version() string
	Tags() []string
}

// Port represents an input or output connection point for a component.
// It defines the name, type, and other properties of a data channel.
type Port interface {
	// Basic port information
	Name() string
	Type() reflect.Type
	Required() bool
	Description() string
	
	// Enhanced metadata
	Schema() Schema
	DefaultValue() interface{}
	Constraints() []Constraint
	
	// Documentation and examples
	Examples() []interface{}
	Documentation() string
}

// Schema provides data validation and evolution support
type Schema interface {
	Validate(data interface{}) error
	Compatible(other Schema) bool
	Migrate(data interface{}, targetSchema Schema) (interface{}, error)
	JSONSchema() string
}

// Constraint defines validation rules for data
type Constraint interface {
	Validate(data interface{}) error
	Description() string
}

// PipelineError provides structured error information with recovery strategies
type PipelineError interface {
	error
	Component() string
	ErrorType() ErrorType
	Severity() Severity
	Recoverable() bool
	Context() map[string]interface{}
}

// ErrorType categorizes different types of errors
type ErrorType int

const (
	ValidationError ErrorType = iota
	RuntimeError
	ConfigurationError
	ResourceError
	NetworkError
)

// Severity indicates the severity level of an error
type Severity int

const (
	Info Severity = iota
	Warning
	Error
	Critical
)

// ErrorHandler defines how to handle different types of errors
type ErrorHandler interface {
	HandleError(ctx context.Context, err PipelineError) ErrorAction
	CanRecover(err PipelineError) bool
}

// ErrorAction defines what action to take when an error occurs
type ErrorAction int

const (
	Continue ErrorAction = iota
	Retry
	Skip
	Abort
)

// CircuitBreaker implements circuit breaker pattern for resilience
type CircuitBreaker interface {
	Execute(ctx context.Context, fn func() (interface{}, error)) (interface{}, error)
	State() CircuitState
	Reset()
}

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	Closed CircuitState = iota
	Open
	HalfOpen
)

// ExecutionEngine defines the interface for a pipeline execution engine.
type ExecutionEngine interface {
	// Run executes the given pipeline.
	Run(ctx context.Context, p *Pipeline, inputs, outputs map[string]chan interface{}) error
	// Close gracefully shuts down the engine.
	Close() error
}
