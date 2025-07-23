package core

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

var (
	defaultEngineCreator func() ExecutionEngine
)

// SetDefaultEngineCreator sets the function used to create a default execution engine.
func SetDefaultEngineCreator(creator func() ExecutionEngine) {
	defaultEngineCreator = creator
}

// Pipeline represents a data processing pipeline.
type Pipeline struct {
	name        string
	components  map[string]Component
	connections []Connection
	errors      []error
	engine      ExecutionEngine
}

// Connection represents a connection between two component ports.
type Connection struct {
	FromComponent string
	FromPort    string
	ToComponent   string
	ToPort      string
}

// NewPipeline creates a new pipeline with the given name.
func NewPipeline(name string) *Pipeline {
	return &Pipeline{
		name:       name,
		components: make(map[string]Component),
	}
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

	p.connections = append(p.connections, Connection{
		FromComponent: fromComponent,
		FromPort:      fromPort,
		ToComponent:   toComponent,
		ToPort:        toPort,
	})
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

	// Add more validation checks here in the future

	return nil
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
