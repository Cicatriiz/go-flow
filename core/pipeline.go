package core

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"
)

var (
	defaultEngineCreator func() ExecutionEngine
)

// SetDefaultEngineCreator sets the function used to create a default execution engine.
func SetDefaultEngineCreator(creator func() ExecutionEngine) {
	defaultEngineCreator = creator
}

// Pipeline represents a data processing pipeline with enhanced configuration and validation.
type Pipeline struct {
	// Core identification
	name        string
	version     string
	description string
	
	// Graph structure
	components  map[string]Component
	connections []Connection
	
	// Configuration and state
	config      *PipelineConfig
	metadata    map[string]interface{}
	
	// Execution context
	engine      ExecutionEngine
	context     *PipelineContext
	
	// Error tracking and validation
	errors          []error
	pipelineErrors  []PipelineError
	errorCollector  *ErrorCollector
	validator       *PipelineValidator
}

// Connection represents a connection between two component ports with enhanced configuration.
type Connection struct {
	// Basic connection information
	FromComponent string
	FromPort      string
	ToComponent   string
	ToPort        string
	
	// Connection metadata
	Transform     DataTransform
	BufferSize    int
	Backpressure  *BackpressureConfig
	
	// Connection properties
	Name         string
	Description  string
	Metadata     map[string]interface{}
}

// PipelineConfig contains execution, resource, and monitoring settings.
type PipelineConfig struct {
	// Execution configuration
	MaxConcurrency    int
	Timeout          time.Duration
	RetryPolicy      *RetryPolicy
	
	// Resource limits
	MemoryLimit      int64
	CPULimit         float64
	
	// Monitoring configuration
	MetricsEnabled   bool
	TracingEnabled   bool
	LogLevel         string
	
	// Validation settings
	StrictValidation bool
	AllowCycles      bool
	
	// Buffer configuration
	DefaultBufferSize int
	MaxBufferSize     int
}

// PipelineContext holds runtime context and state information.
type PipelineContext struct {
	ExecutionID    string
	StartTime      time.Time
	Status         PipelineStatus
	ComponentStates map[string]ComponentState
	Metrics        *PipelineMetrics
	
	// Context data
	Variables      map[string]interface{}
	Tags           map[string]string
}

// DataTransform defines a transformation function for connection data.
type DataTransform interface {
	Transform(ctx context.Context, data interface{}) (interface{}, error)
	Name() string
	Description() string
}

// BackpressureConfig defines backpressure handling configuration.
type BackpressureConfig struct {
	Strategy     BackpressureStrategy
	BufferSize   int
	DropPolicy   DropPolicy
	Timeout      time.Duration
	MaxRetries   int
}

// RetryPolicy defines retry behavior for failed operations.
type RetryPolicy struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
	RetryableErrors []ErrorType
}

// PipelineMetrics holds runtime metrics for the pipeline.
type PipelineMetrics struct {
	ComponentMetrics map[string]*ComponentMetrics
	TotalProcessed   int64
	TotalErrors      int64
	Throughput       float64
	Latency          time.Duration
	StartTime        time.Time
	LastUpdate       time.Time
}

// ComponentMetrics holds metrics for individual components.
type ComponentMetrics struct {
	ProcessedCount int64
	ErrorCount     int64
	AverageLatency time.Duration
	LastProcessed  time.Time
	Status         ComponentState
}

// Enums for various states and strategies
type PipelineStatus int
const (
	PipelineStatusIdle PipelineStatus = iota
	PipelineStatusRunning
	PipelineStatusPaused
	PipelineStatusStopped
	PipelineStatusError
)

type ComponentState int
const (
	ComponentStateIdle ComponentState = iota
	ComponentStateRunning
	ComponentStatePaused
	ComponentStateError
	ComponentStateCompleted
)

type BackpressureStrategy int
const (
	BackpressureBlock BackpressureStrategy = iota
	BackpressureDrop
	BackpressureBuffer
)

type DropPolicy int
const (
	DropOldest DropPolicy = iota
	DropNewest
	DropRandom
)

// NewPipeline creates a new pipeline with the given name.
func NewPipeline(name string) *Pipeline {
	return &Pipeline{
		name:           name,
		version:        "1.0.0",
		description:    "",
		components:     make(map[string]Component),
		connections:    make([]Connection, 0),
		config:         NewDefaultPipelineConfig(),
		metadata:       make(map[string]interface{}),
		context:        NewPipelineContext(),
		errors:         make([]error, 0),
		pipelineErrors: make([]PipelineError, 0),
		errorCollector: NewErrorCollector(),
		validator:      NewPipelineValidator(),
	}
}

// NewPipelineWithConfig creates a new pipeline with custom configuration.
func NewPipelineWithConfig(name string, config *PipelineConfig) *Pipeline {
	p := NewPipeline(name)
	p.config = config
	return p
}

// NewDefaultPipelineConfig creates a default pipeline configuration.
func NewDefaultPipelineConfig() *PipelineConfig {
	return &PipelineConfig{
		MaxConcurrency:    10,
		Timeout:          30 * time.Second,
		RetryPolicy:      NewDefaultRetryPolicy(),
		MemoryLimit:      1024 * 1024 * 1024, // 1GB
		CPULimit:         1.0,
		MetricsEnabled:   true,
		TracingEnabled:   false,
		LogLevel:         "INFO",
		StrictValidation: true,
		AllowCycles:      false,
		DefaultBufferSize: 100,
		MaxBufferSize:    1000,
	}
}

// NewDefaultRetryPolicy creates a default retry policy.
func NewDefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:      3,
		InitialDelay:    100 * time.Millisecond,
		MaxDelay:        5 * time.Second,
		BackoffFactor:   2.0,
		RetryableErrors: []ErrorType{RuntimeError, NetworkError, ResourceError},
	}
}

// NewPipelineContext creates a new pipeline context.
func NewPipelineContext() *PipelineContext {
	return &PipelineContext{
		ExecutionID:     generateExecutionID(),
		StartTime:       time.Now(),
		Status:          PipelineStatusIdle,
		ComponentStates: make(map[string]ComponentState),
		Metrics:         NewPipelineMetrics(),
		Variables:       make(map[string]interface{}),
		Tags:            make(map[string]string),
	}
}

// NewPipelineMetrics creates a new pipeline metrics instance.
func NewPipelineMetrics() *PipelineMetrics {
	return &PipelineMetrics{
		ComponentMetrics: make(map[string]*ComponentMetrics),
		TotalProcessed:   0,
		TotalErrors:      0,
		Throughput:       0.0,
		Latency:          0,
		StartTime:        time.Now(),
		LastUpdate:       time.Now(),
	}
}

// generateExecutionID generates a unique execution ID.
func generateExecutionID() string {
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}

// AddComponent adds a component to the pipeline.
func (p *Pipeline) AddComponent(name string, component Component) *Pipeline {
	component.SetName(name)
	p.components[name] = component
	return p
}

// Connect connects an output port of one component to an input port of another.
// It uses generics to enforce type safety at compile time.
func Connect[T any](p *Pipeline, fromComponent, fromPort, toComponent, toPort string) *Pipeline {
	// Validate components exist
	from, ok := p.components[fromComponent]
	if !ok {
		p.errors = append(p.errors, fmt.Errorf("source component '%s' not found", fromComponent))
		return p
	}
	to, ok := p.components[toComponent]
	if !ok {
		p.errors = append(p.errors, fmt.Errorf("target component '%s' not found", toComponent))
		return p
	}

	// Validate ports exist and types match
	if err := validatePortMatch[T](from, fromPort, to, toPort); err != nil {
		p.errors = append(p.errors, err)
		return p
	}

	connection := Connection{
		FromComponent: fromComponent,
		FromPort:      fromPort,
		ToComponent:   toComponent,
		ToPort:        toPort,
		BufferSize:    p.config.DefaultBufferSize,
		Name:          fmt.Sprintf("%s.%s -> %s.%s", fromComponent, fromPort, toComponent, toPort),
		Description:   fmt.Sprintf("Connection from %s to %s", fromComponent, toComponent),
		Metadata:      make(map[string]interface{}),
	}
	
	p.connections = append(p.connections, connection)
	return p
}

// Errors returns any errors that occurred during pipeline construction.
func (p *Pipeline) Errors() []error {
	return p.errors
}

// SetEngine sets the execution engine for the pipeline.
func (p *Pipeline) SetEngine(engine ExecutionEngine) *Pipeline {
	p.engine = engine
	return p
}

// Run executes the pipeline using the configured execution engine.
func (p *Pipeline) Run(ctx context.Context) error {
	if len(p.errors) > 0 {
		return fmt.Errorf("pipeline has %d construction errors", len(p.errors))
	}
	if p.engine == nil {
		return fmt.Errorf("execution engine is not set")
	}
	return p.engine.Run(ctx, p, nil, nil)
}

// GetComponents returns the components in the pipeline.
func (p *Pipeline) GetComponents() map[string]Component {
	return p.components
}

// GetConnections returns the connections in the pipeline.
func (p *Pipeline) GetConnections() []Connection {
	return p.connections
}

// Name returns the name of the pipeline.
func (p *Pipeline) Name() string {
	return p.name
}

// SetName sets the name of the pipeline.
func (p *Pipeline) SetName(name string) {
	p.name = name
}

// InputPorts returns the input ports of the pipeline.
func (p *Pipeline) InputPorts() []Port {
	// For a pipeline to be a component, we need to define its public-facing ports.
	// This is a simplified implementation where we expose all unconnected input ports.
	var ports []Port
	for _, component := range p.components {
		for _, port := range component.InputPorts() {
			isConnected := false
			for _, conn := range p.connections {
				if conn.ToComponent == component.Name() && conn.ToPort == port.Name() {
					isConnected = true
					break
				}
			}
			if !isConnected {
				ports = append(ports, port)
			}
		}
	}
	return ports
}

// OutputPorts returns the output ports of the pipeline.
func (p *Pipeline) OutputPorts() []Port {
	// Similar to InputPorts, we expose all unconnected output ports.
	var ports []Port
	for _, component := range p.components {
		for _, port := range component.OutputPorts() {
			isConnected := false
			for _, conn := range p.connections {
				if conn.FromComponent == component.Name() && conn.FromPort == port.Name() {
					isConnected = true
					break
				}
			}
			if !isConnected {
				ports = append(ports, port)
			}
		}
	}
	return ports
}

// Process runs the pipeline as a component.
func (p *Pipeline) Process(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	var wg sync.WaitGroup
	// Create a new engine for the sub-pipeline
	if defaultEngineCreator == nil {
		return nil, fmt.Errorf("no default engine creator is registered")
	}
	engine := defaultEngineCreator()

	// Create channels for the pipeline's inputs and outputs
	inputChans := make(map[string]chan interface{})
	for _, port := range p.InputPorts() {
		inputChans[port.Name()] = make(chan interface{}, 1)
	}
	outputChans := make(map[string]chan interface{})
	for _, port := range p.OutputPorts() {
		outputChans[port.Name()] = make(chan interface{}, 1)
	}

	wg.Add(1)
	// Start the pipeline in a goroutine
	go func() {
		defer wg.Done()
		if err := engine.Run(ctx, p, inputChans, outputChans); err != nil {
			// In a real implementation, we would propagate this error.
			fmt.Printf("error running sub-pipeline: %v\n", err)
		}
		engine.Close()
	}()

	// Write the external inputs to the pipeline's input channels
	for name, data := range inputs {
		if ch, ok := inputChans[name]; ok {
			ch <- data
		}
	}

	// Close the output channels
	for _, ch := range outputChans {
		close(ch)
	}

	wg.Wait()

	// Read from the output channels and return the data
	outputs := make(map[string]interface{})
	var outputWg sync.WaitGroup
	for name, ch := range outputChans {
		outputWg.Add(1)
		go func(name string, ch chan interface{}) {
			defer outputWg.Done()
			for data := range ch {
				outputs[name] = data
			}
		}(name, ch)
	}
	outputWg.Wait()
	return outputs, nil
}

// Validate validates the pipeline.
func (p *Pipeline) Validate() error {
	if len(p.errors) > 0 {
		// Return the first construction error found
		return p.errors[0]
	}

	// Check for cycles in the graph
	if err := p.detectCycles(); err != nil {
		return err
	}

	// Validate all components
	for _, component := range p.components {
		if err := component.Validate(); err != nil {
			return fmt.Errorf("component %s validation failed: %w", component.Name(), err)
		}
	}

	return nil
}

// HealthCheck performs health checks on all components in the pipeline
func (p *Pipeline) HealthCheck(ctx context.Context) error {
	for _, component := range p.components {
		if err := component.HealthCheck(ctx); err != nil {
			return fmt.Errorf("component %s health check failed: %w", component.Name(), err)
		}
	}
	return nil
}

// Initialize initializes all components in the pipeline
func (p *Pipeline) Initialize(ctx context.Context) error {
	for _, component := range p.components {
		if err := component.Initialize(ctx); err != nil {
			return fmt.Errorf("component %s initialization failed: %w", component.Name(), err)
		}
	}
	return nil
}

// Cleanup cleans up all components in the pipeline
func (p *Pipeline) Cleanup(ctx context.Context) error {
	var errors []error
	
	// Clean up components in reverse order
	componentNames := make([]string, 0, len(p.components))
	for name := range p.components {
		componentNames = append(componentNames, name)
	}
	
	// Cleanup in reverse order
	for i := len(componentNames) - 1; i >= 0; i-- {
		component := p.components[componentNames[i]]
		if err := component.Cleanup(ctx); err != nil {
			errors = append(errors, fmt.Errorf("component %s cleanup failed: %w", component.Name(), err))
		}
	}
	
	// Close the execution engine if it exists
	if p.engine != nil {
		if err := p.engine.Close(); err != nil {
			errors = append(errors, fmt.Errorf("engine cleanup failed: %w", err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("pipeline cleanup had %d errors: %v", len(errors), errors)
	}
	
	return nil
}

// Description returns a description of the pipeline
func (p *Pipeline) Description() string {
	return fmt.Sprintf("Pipeline '%s' with %d components and %d connections", p.name, len(p.components), len(p.connections))
}

// Version returns the version of the pipeline
func (p *Pipeline) Version() string {
	return "1.0.0"
}

// Tags returns tags associated with the pipeline
func (p *Pipeline) Tags() []string {
	return []string{"pipeline", "composite"}
}

// detectCycles checks for cycles in the pipeline graph.
func (p *Pipeline) detectCycles() error {
	graph := make(map[string][]string)
	for _, conn := range p.GetConnections() {
		graph[conn.FromComponent] = append(graph[conn.FromComponent], conn.ToComponent)
	}

	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for component := range p.GetComponents() {
		if !visited[component] {
			if p.hasCycle(component, visited, recursionStack, graph) {
				return fmt.Errorf("cycle detected in pipeline graph involving component %s", component)
			}
		}
	}

	return nil
}

func (p *Pipeline) hasCycle(component string, visited, recursionStack map[string]bool, graph map[string][]string) bool {
	visited[component] = true
	recursionStack[component] = true

	for _, neighbor := range graph[component] {
		if !visited[neighbor] {
			if p.hasCycle(neighbor, visited, recursionStack, graph) {
				return true
			}
		} else if recursionStack[neighbor] {
			return true
		}
	}

	recursionStack[component] = false
	return false
}

// validatePortMatch checks if the ports of two components can be connected.
func validatePortMatch[T any](from Component, fromPort string, to Component, toPort string) error {
	var zero T
	expectedType := reflect.TypeOf(zero)

	outPort, err := findPort(from.OutputPorts(), fromPort, expectedType)
	if err != nil {
		return fmt.Errorf("output port validation failed for %s: %w", from.Name(), err)
	}

	inPort, err := findPort(to.InputPorts(), toPort, expectedType)
	if err != nil {
		return fmt.Errorf("input port validation failed for %s: %w", to.Name(), err)
	}

	if outPort.Type() != inPort.Type() {
		return fmt.Errorf("type mismatch: cannot connect %s (%s) to %s (%s)",
			outPort.Type(), fromPort, inPort.Type(), toPort)
	}

	return nil
}

func findPort(ports []Port, name string, expectedType reflect.Type) (Port, error) {
	for _, p := range ports {
		if p.Name() == name {
			if p.Type() == expectedType {
				return p, nil
			}
			return nil, fmt.Errorf("port '%s' has type %s, but expected %s", name, p.Type(), expectedType)
		}
	}
	return nil, fmt.Errorf("port '%s' not found", name)
}
// Enhanced Pipeline Methods

// SetVersion sets the version of the pipeline
func (p *Pipeline) SetVersion(version string) *Pipeline {
	p.version = version
	return p
}

// GetVersion returns the version of the pipeline
func (p *Pipeline) GetVersion() string {
	return p.version
}

// SetDescription sets the description of the pipeline
func (p *Pipeline) SetDescription(description string) *Pipeline {
	p.description = description
	return p
}

// GetDescription returns the description of the pipeline
func (p *Pipeline) GetDescription() string {
	return p.description
}

// SetConfig sets the pipeline configuration
func (p *Pipeline) SetConfig(config *PipelineConfig) *Pipeline {
	p.config = config
	return p
}

// GetConfig returns the pipeline configuration
func (p *Pipeline) GetConfig() *PipelineConfig {
	return p.config
}

// SetMetadata sets metadata for the pipeline
func (p *Pipeline) SetMetadata(key string, value interface{}) *Pipeline {
	p.metadata[key] = value
	return p
}

// GetMetadata returns metadata value for the given key
func (p *Pipeline) GetMetadata(key string) interface{} {
	return p.metadata[key]
}

// GetAllMetadata returns all metadata
func (p *Pipeline) GetAllMetadata() map[string]interface{} {
	return p.metadata
}

// GetContext returns the pipeline context
func (p *Pipeline) GetContext() *PipelineContext {
	return p.context
}

// GetPipelineErrors returns pipeline-specific errors
func (p *Pipeline) GetPipelineErrors() []PipelineError {
	return p.pipelineErrors
}

// AddPipelineError adds a pipeline error
func (p *Pipeline) AddPipelineError(err PipelineError) {
	p.pipelineErrors = append(p.pipelineErrors, err)
	p.errorCollector.Collect(err)
}

// GetErrorCollector returns the error collector
func (p *Pipeline) GetErrorCollector() *ErrorCollector {
	return p.errorCollector
}

// ConnectWithTransform connects components with a data transformation
func (p *Pipeline) ConnectWithTransform(fromComponent, fromPort, toComponent, toPort string, transform DataTransform) *Pipeline {
	// Find existing connection or create new one
	var connection *Connection
	for i := range p.connections {
		conn := &p.connections[i]
		if conn.FromComponent == fromComponent && conn.FromPort == fromPort &&
		   conn.ToComponent == toComponent && conn.ToPort == toPort {
			connection = conn
			break
		}
	}
	
	if connection == nil {
		// Create new connection
		newConnection := Connection{
			FromComponent: fromComponent,
			FromPort:      fromPort,
			ToComponent:   toComponent,
			ToPort:        toPort,
			BufferSize:    p.config.DefaultBufferSize,
			Name:          fmt.Sprintf("%s.%s -> %s.%s", fromComponent, fromPort, toComponent, toPort),
			Description:   fmt.Sprintf("Connection from %s to %s with transform", fromComponent, toComponent),
			Metadata:      make(map[string]interface{}),
		}
		p.connections = append(p.connections, newConnection)
		connection = &p.connections[len(p.connections)-1]
	}
	
	connection.Transform = transform
	return p
}

// ConnectWithBackpressure connects components with backpressure configuration
func (p *Pipeline) ConnectWithBackpressure(fromComponent, fromPort, toComponent, toPort string, backpressure *BackpressureConfig) *Pipeline {
	// Find existing connection or create new one
	var connection *Connection
	for i := range p.connections {
		conn := &p.connections[i]
		if conn.FromComponent == fromComponent && conn.FromPort == fromPort &&
		   conn.ToComponent == toComponent && conn.ToPort == toPort {
			connection = conn
			break
		}
	}
	
	if connection == nil {
		// Create new connection
		newConnection := Connection{
			FromComponent: fromComponent,
			FromPort:      fromPort,
			ToComponent:   toComponent,
			ToPort:        toPort,
			BufferSize:    p.config.DefaultBufferSize,
			Name:          fmt.Sprintf("%s.%s -> %s.%s", fromComponent, fromPort, toComponent, toPort),
			Description:   fmt.Sprintf("Connection from %s to %s with backpressure", fromComponent, toComponent),
			Metadata:      make(map[string]interface{}),
		}
		p.connections = append(p.connections, newConnection)
		connection = &p.connections[len(p.connections)-1]
	}
	
	connection.Backpressure = backpressure
	return p
}

// SetConnectionBufferSize sets the buffer size for a specific connection
func (p *Pipeline) SetConnectionBufferSize(fromComponent, fromPort, toComponent, toPort string, bufferSize int) *Pipeline {
	for i := range p.connections {
		conn := &p.connections[i]
		if conn.FromComponent == fromComponent && conn.FromPort == fromPort &&
		   conn.ToComponent == toComponent && conn.ToPort == toPort {
			conn.BufferSize = bufferSize
			break
		}
	}
	return p
}

// String methods for enums
func (ps PipelineStatus) String() string {
	switch ps {
	case PipelineStatusIdle:
		return "IDLE"
	case PipelineStatusRunning:
		return "RUNNING"
	case PipelineStatusPaused:
		return "PAUSED"
	case PipelineStatusStopped:
		return "STOPPED"
	case PipelineStatusError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func (cs ComponentState) String() string {
	switch cs {
	case ComponentStateIdle:
		return "IDLE"
	case ComponentStateRunning:
		return "RUNNING"
	case ComponentStatePaused:
		return "PAUSED"
	case ComponentStateError:
		return "ERROR"
	case ComponentStateCompleted:
		return "COMPLETED"
	default:
		return "UNKNOWN"
	}
}

func (bs BackpressureStrategy) String() string {
	switch bs {
	case BackpressureBlock:
		return "BLOCK"
	case BackpressureDrop:
		return "DROP"
	case BackpressureBuffer:
		return "BUFFER"
	default:
		return "UNKNOWN"
	}
}

func (dp DropPolicy) String() string {
	switch dp {
	case DropOldest:
		return "DROP_OLDEST"
	case DropNewest:
		return "DROP_NEWEST"
	case DropRandom:
		return "DROP_RANDOM"
	default:
		return "UNKNOWN"
	}
}
// ValidateComprehensive performs comprehensive validation using the validator
func (p *Pipeline) ValidateComprehensive() *ValidationResult {
	return p.validator.ValidateComprehensive(p)
}

// GetComponentGraph returns the component graph for analysis
func (p *Pipeline) GetComponentGraph() (*ComponentGraph, error) {
	return p.validator.buildComponentGraph(p)
}

// GetTopologyOrder returns the topological order of components
func (p *Pipeline) GetTopologyOrder() ([]string, error) {
	graph, err := p.GetComponentGraph()
	if err != nil {
		return nil, err
	}
	return graph.TopologyOrder, nil
}

// GetCriticalPath returns the critical path through the pipeline
func (p *Pipeline) GetCriticalPath() ([]string, error) {
	graph, err := p.GetComponentGraph()
	if err != nil {
		return nil, err
	}
	return graph.CriticalPath, nil
}