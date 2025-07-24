package core

import (
	"fmt"
)

// PipelineValidator provides comprehensive validation for pipelines
type PipelineValidator struct {
	strictMode  bool
	allowCycles bool
}

// NewPipelineValidator creates a new pipeline validator
func NewPipelineValidator() *PipelineValidator {
	return &PipelineValidator{
		strictMode:  true,
		allowCycles: false,
	}
}

// ValidationResult holds the results of pipeline validation
type ValidationResult struct {
	Valid          bool
	Errors         []PipelineValidationError
	Warnings       []ValidationWarning
	ComponentGraph *ComponentGraph
}

// PipelineValidationError represents a validation error
type PipelineValidationError struct {
	Type        ValidationErrorType
	Component   string
	Port        string
	Connection  string
	Message     string
	Severity    Severity
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Type      ValidationWarningType
	Component string
	Message   string
}

// ComponentGraph represents the component dependency graph
type ComponentGraph struct {
	Nodes         map[string]*GraphNode
	Edges         []*GraphEdge
	TopologyOrder []string
	CriticalPath  []string
}

// GraphNode represents a component in the graph
type GraphNode struct {
	Name         string
	Component    Component
	Dependencies []string
	Dependents   []string
	Level        int
}

// GraphEdge represents a connection in the graph
type GraphEdge struct {
	From       string
	To         string
	Connection *Connection
	Weight     int
}

// Validation error and warning types
type ValidationErrorType int

const (
	ValidationErrorTypeUnknown ValidationErrorType = iota
	ValidationErrorTypeMissingComponent
	ValidationErrorTypeMissingPort
	ValidationErrorTypeTypeMismatch
	ValidationErrorTypeCycle
	ValidationErrorTypeDisconnectedComponent
	ValidationErrorTypeInvalidConfiguration
	ValidationErrorTypeResourceLimit
)

type ValidationWarningType int

const (
	ValidationWarningTypeUnused ValidationWarningType = iota
	ValidationWarningTypePerformance
	ValidationWarningTypeConfiguration
)

// ValidateComprehensive performs comprehensive validation of the pipeline
func (pv *PipelineValidator) ValidateComprehensive(p *Pipeline) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]PipelineValidationError, 0),
		Warnings: make([]ValidationWarning, 0),
	}

	// Build component graph
	graph, err := pv.buildComponentGraph(p)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, PipelineValidationError{
			Type:     ValidationErrorTypeUnknown,
			Message:  fmt.Sprintf("Failed to build component graph: %v", err),
			Severity: Critical,
		})
		return result
	}
	result.ComponentGraph = graph

	// Perform various validation checks
	pv.validateComponents(p, result)
	pv.validateConnections(p, result)
	pv.validateTypes(p, result)
	pv.validateGraph(p, result)
	pv.validateConfiguration(p, result)
	pv.validateResources(p, result)

	// Check if any critical errors were found
	for _, err := range result.Errors {
		if err.Severity == Critical || err.Severity == Error {
			result.Valid = false
		}
	}

	return result
}

// buildComponentGraph builds a graph representation of the pipeline
func (pv *PipelineValidator) buildComponentGraph(p *Pipeline) (*ComponentGraph, error) {
	graph := &ComponentGraph{
		Nodes: make(map[string]*GraphNode),
		Edges: make([]*GraphEdge, 0),
	}

	// Create nodes for each component
	for name, component := range p.components {
		node := &GraphNode{
			Name:         name,
			Component:    component,
			Dependencies: make([]string, 0),
			Dependents:   make([]string, 0),
			Level:        0,
		}
		graph.Nodes[name] = node
	}

	// Create edges for each connection
	for _, conn := range p.connections {
		edge := &GraphEdge{
			From:       conn.FromComponent,
			To:         conn.ToComponent,
			Connection: &conn,
			Weight:     1,
		}
		graph.Edges = append(graph.Edges, edge)

		// Update node dependencies
		if fromNode, ok := graph.Nodes[conn.FromComponent]; ok {
			fromNode.Dependents = append(fromNode.Dependents, conn.ToComponent)
		}
		if toNode, ok := graph.Nodes[conn.ToComponent]; ok {
			toNode.Dependencies = append(toNode.Dependencies, conn.FromComponent)
		}
	}

	// Calculate topology order - don't fail if there are cycles, just skip topology calculation
	if err := pv.calculateTopologyOrder(graph); err != nil {
		// If there's a cycle, we can't calculate topology order, but we can still return the graph
		// The cycle will be detected later in validateGraph
		graph.TopologyOrder = nil
	} else {
		// Calculate critical path only if we have a valid topology order
		pv.calculateCriticalPath(graph)
	}

	return graph, nil
}

// validateComponents validates individual components
func (pv *PipelineValidator) validateComponents(p *Pipeline, result *ValidationResult) {
	for name, component := range p.components {
		// Validate component itself
		if err := component.Validate(); err != nil {
			result.Errors = append(result.Errors, PipelineValidationError{
				Type:      ValidationErrorTypeInvalidConfiguration,
				Component: name,
				Message:   fmt.Sprintf("Component validation failed: %v", err),
				Severity:  Error,
			})
		}

		// Check for required input ports
		for _, port := range component.InputPorts() {
			if port.Required() {
				connected := false
				for _, conn := range p.connections {
					if conn.ToComponent == name && conn.ToPort == port.Name() {
						connected = true
						break
					}
				}
				if !connected {
					result.Errors = append(result.Errors, PipelineValidationError{
						Type:      ValidationErrorTypeMissingPort,
						Component: name,
						Port:      port.Name(),
						Message:   fmt.Sprintf("Required input port '%s' is not connected", port.Name()),
						Severity:  Error,
					})
				}
			}
		}

		// Check for unused output ports
		for _, port := range component.OutputPorts() {
			connected := false
			for _, conn := range p.connections {
				if conn.FromComponent == name && conn.FromPort == port.Name() {
					connected = true
					break
				}
			}
			if !connected {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Type:      ValidationWarningTypeUnused,
					Component: name,
					Message:   fmt.Sprintf("Output port '%s' is not connected", port.Name()),
				})
			}
		}
	}
}

// validateConnections validates all connections in the pipeline
func (pv *PipelineValidator) validateConnections(p *Pipeline, result *ValidationResult) {
	for _, conn := range p.connections {
		// Validate source component exists
		fromComponent, ok := p.components[conn.FromComponent]
		if !ok {
			result.Errors = append(result.Errors, PipelineValidationError{
				Type:       ValidationErrorTypeMissingComponent,
				Component:  conn.FromComponent,
				Connection: conn.Name,
				Message:    fmt.Sprintf("Source component '%s' not found", conn.FromComponent),
				Severity:   Critical,
			})
			continue
		}

		// Validate target component exists
		toComponent, ok := p.components[conn.ToComponent]
		if !ok {
			result.Errors = append(result.Errors, PipelineValidationError{
				Type:       ValidationErrorTypeMissingComponent,
				Component:  conn.ToComponent,
				Connection: conn.Name,
				Message:    fmt.Sprintf("Target component '%s' not found", conn.ToComponent),
				Severity:   Critical,
			})
			continue
		}

		// Validate source port exists
		var fromPort Port
		found := false
		for _, port := range fromComponent.OutputPorts() {
			if port.Name() == conn.FromPort {
				fromPort = port
				found = true
				break
			}
		}
		if !found {
			result.Errors = append(result.Errors, PipelineValidationError{
				Type:       ValidationErrorTypeMissingPort,
				Component:  conn.FromComponent,
				Port:       conn.FromPort,
				Connection: conn.Name,
				Message:    fmt.Sprintf("Output port '%s' not found in component '%s'", conn.FromPort, conn.FromComponent),
				Severity:   Error,
			})
			continue
		}

		// Validate target port exists
		var toPort Port
		found = false
		for _, port := range toComponent.InputPorts() {
			if port.Name() == conn.ToPort {
				toPort = port
				found = true
				break
			}
		}
		if !found {
			result.Errors = append(result.Errors, PipelineValidationError{
				Type:       ValidationErrorTypeMissingPort,
				Component:  conn.ToComponent,
				Port:       conn.ToPort,
				Connection: conn.Name,
				Message:    fmt.Sprintf("Input port '%s' not found in component '%s'", conn.ToPort, conn.ToComponent),
				Severity:   Error,
			})
			continue
		}

		// Validate type compatibility
		if fromPort.Type() != toPort.Type() {
			result.Errors = append(result.Errors, PipelineValidationError{
				Type:       ValidationErrorTypeTypeMismatch,
				Component:  conn.FromComponent,
				Port:       conn.FromPort,
				Connection: conn.Name,
				Message:    fmt.Sprintf("Type mismatch: cannot connect %s (%s) to %s (%s)", fromPort.Type(), conn.FromPort, toPort.Type(), conn.ToPort),
				Severity:   Error,
			})
		}

		// Validate buffer size
		if conn.BufferSize <= 0 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:      ValidationWarningTypeConfiguration,
				Component: conn.FromComponent,
				Message:   fmt.Sprintf("Connection '%s' has invalid buffer size: %d", conn.Name, conn.BufferSize),
			})
		}
	}
}

// validateTypes performs comprehensive type checking
func (pv *PipelineValidator) validateTypes(p *Pipeline, result *ValidationResult) {
	// This method can be extended to perform more sophisticated type checking
	// such as schema validation, constraint checking, etc.

	for _, conn := range p.connections {
		fromComponent := p.components[conn.FromComponent]
		toComponent := p.components[conn.ToComponent]

		if fromComponent == nil || toComponent == nil {
			continue // Already handled in validateConnections
		}

		// Find the ports
		var fromPort, toPort Port
		for _, port := range fromComponent.OutputPorts() {
			if port.Name() == conn.FromPort {
				fromPort = port
				break
			}
		}
		for _, port := range toComponent.InputPorts() {
			if port.Name() == conn.ToPort {
				toPort = port
				break
			}
		}

		if fromPort == nil || toPort == nil {
			continue // Already handled in validateConnections
		}

		// Validate schemas if available
		if fromPort.Schema() != nil && toPort.Schema() != nil {
			if !fromPort.Schema().Compatible(toPort.Schema()) {
				result.Errors = append(result.Errors, PipelineValidationError{
					Type:       ValidationErrorTypeTypeMismatch,
					Component:  conn.FromComponent,
					Port:       conn.FromPort,
					Connection: conn.Name,
					Message:    fmt.Sprintf("Schema incompatibility between %s.%s and %s.%s", conn.FromComponent, conn.FromPort, conn.ToComponent, conn.ToPort),
					Severity:   Error,
				})
			}
		}
	}
}

// validateGraph validates the overall graph structure
func (pv *PipelineValidator) validateGraph(p *Pipeline, result *ValidationResult) {
	// Check for cycles
	if !pv.allowCycles {
		if err := pv.detectCycles(p); err != nil {
			result.Errors = append(result.Errors, PipelineValidationError{
				Type:     ValidationErrorTypeCycle,
				Message:  err.Error(),
				Severity: Error,
			})
		}
	}

	// Check for disconnected components
	pv.validateConnectivity(p, result)
}

// validateConnectivity checks for disconnected components
func (pv *PipelineValidator) validateConnectivity(p *Pipeline, result *ValidationResult) {
	if len(p.components) == 0 {
		return
	}

	// Build adjacency list
	graph := make(map[string][]string)
	for _, conn := range p.connections {
		graph[conn.FromComponent] = append(graph[conn.FromComponent], conn.ToComponent)
		graph[conn.ToComponent] = append(graph[conn.ToComponent], conn.FromComponent)
	}

	// Find connected components using DFS
	visited := make(map[string]bool)
	var connectedComponents [][]string

	for componentName := range p.components {
		if !visited[componentName] {
			var component []string
			pv.dfsConnectivity(componentName, graph, visited, &component)
			connectedComponents = append(connectedComponents, component)
		}
	}

	// If there are multiple connected components, report as warning
	if len(connectedComponents) > 1 {
		for i, component := range connectedComponents {
			if len(component) == 1 {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Type:      ValidationWarningTypeConfiguration,
					Component: component[0],
					Message:   fmt.Sprintf("Component '%s' is disconnected from the main pipeline", component[0]),
				})
			} else {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Type:    ValidationWarningTypeConfiguration,
					Message: fmt.Sprintf("Disconnected component group %d: %v", i+1, component),
				})
			}
		}
	}
}

// dfsConnectivity performs DFS for connectivity analysis
func (pv *PipelineValidator) dfsConnectivity(node string, graph map[string][]string, visited map[string]bool, component *[]string) {
	visited[node] = true
	*component = append(*component, node)

	for _, neighbor := range graph[node] {
		if !visited[neighbor] {
			pv.dfsConnectivity(neighbor, graph, visited, component)
		}
	}
}

// validateConfiguration validates pipeline configuration
func (pv *PipelineValidator) validateConfiguration(p *Pipeline, result *ValidationResult) {
	config := p.config
	if config == nil {
		result.Errors = append(result.Errors, PipelineValidationError{
			Type:     ValidationErrorTypeInvalidConfiguration,
			Message:  "Pipeline configuration is missing",
			Severity: Error,
		})
		return
	}

	// Validate concurrency settings
	if config.MaxConcurrency <= 0 {
		result.Errors = append(result.Errors, PipelineValidationError{
			Type:     ValidationErrorTypeInvalidConfiguration,
			Message:  fmt.Sprintf("Invalid MaxConcurrency: %d", config.MaxConcurrency),
			Severity: Error,
		})
	}

	// Validate timeout
	if config.Timeout <= 0 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Type:    ValidationWarningTypeConfiguration,
			Message: fmt.Sprintf("Timeout is set to %v, which may cause issues", config.Timeout),
		})
	}

	// Validate buffer sizes
	if config.DefaultBufferSize <= 0 {
		result.Errors = append(result.Errors, PipelineValidationError{
			Type:     ValidationErrorTypeInvalidConfiguration,
			Message:  fmt.Sprintf("Invalid DefaultBufferSize: %d", config.DefaultBufferSize),
			Severity: Error,
		})
	}

	if config.MaxBufferSize < config.DefaultBufferSize {
		result.Errors = append(result.Errors, PipelineValidationError{
			Type:     ValidationErrorTypeInvalidConfiguration,
			Message:  fmt.Sprintf("MaxBufferSize (%d) is less than DefaultBufferSize (%d)", config.MaxBufferSize, config.DefaultBufferSize),
			Severity: Error,
		})
	}
}

// validateResources validates resource limits and requirements
func (pv *PipelineValidator) validateResources(p *Pipeline, result *ValidationResult) {
	config := p.config
	if config == nil {
		return
	}

	// Validate memory limits
	if config.MemoryLimit <= 0 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Type:    ValidationWarningTypeConfiguration,
			Message: "No memory limit set, pipeline may consume excessive memory",
		})
	}

	// Validate CPU limits
	if config.CPULimit <= 0 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Type:    ValidationWarningTypeConfiguration,
			Message: "No CPU limit set, pipeline may consume excessive CPU",
		})
	}

	// Estimate resource requirements based on components
	estimatedMemory := int64(len(p.components) * 1024 * 1024) // 1MB per component estimate
	if config.MemoryLimit > 0 && estimatedMemory > config.MemoryLimit {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Type:    ValidationWarningTypePerformance,
			Message: fmt.Sprintf("Estimated memory usage (%d bytes) may exceed limit (%d bytes)", estimatedMemory, config.MemoryLimit),
		})
	}
}

// calculateTopologyOrder calculates the topological order of components
func (pv *PipelineValidator) calculateTopologyOrder(graph *ComponentGraph) error {
	inDegree := make(map[string]int)

	// Initialize in-degree for all nodes
	for name := range graph.Nodes {
		inDegree[name] = 0
	}

	// Calculate in-degrees
	for _, edge := range graph.Edges {
		inDegree[edge.To]++
	}

	// Queue for nodes with no incoming edges
	queue := make([]string, 0)
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	topologyOrder := make([]string, 0)

	for len(queue) > 0 {
		// Remove node from queue
		current := queue[0]
		queue = queue[1:]
		topologyOrder = append(topologyOrder, current)

		// Update in-degrees of dependent nodes
		if node, ok := graph.Nodes[current]; ok {
			for _, dependent := range node.Dependents {
				inDegree[dependent]--
				if inDegree[dependent] == 0 {
					queue = append(queue, dependent)
				}
			}
		}
	}

	// Check for cycles
	if len(topologyOrder) != len(graph.Nodes) {
		return fmt.Errorf("cycle detected in component graph")
	}

	graph.TopologyOrder = topologyOrder

	// Set levels based on topology order
	for i, name := range topologyOrder {
		if node, ok := graph.Nodes[name]; ok {
			node.Level = i
		}
	}

	return nil
}

// detectCycles checks for cycles in the pipeline graph using DFS
func (pv *PipelineValidator) detectCycles(p *Pipeline) error {
	graph := make(map[string][]string)
	for _, conn := range p.GetConnections() {
		graph[conn.FromComponent] = append(graph[conn.FromComponent], conn.ToComponent)
	}

	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for component := range p.GetComponents() {
		if !visited[component] {
			if pv.hasCycle(component, visited, recursionStack, graph) {
				return fmt.Errorf("cycle detected in pipeline graph involving component %s", component)
			}
		}
	}

	return nil
}

// hasCycle performs DFS to detect cycles
func (pv *PipelineValidator) hasCycle(component string, visited, recursionStack map[string]bool, graph map[string][]string) bool {
	visited[component] = true
	recursionStack[component] = true

	for _, neighbor := range graph[component] {
		if !visited[neighbor] {
			if pv.hasCycle(neighbor, visited, recursionStack, graph) {
				return true
			}
		} else if recursionStack[neighbor] {
			return true
		}
	}

	recursionStack[component] = false
	return false
}

// calculateCriticalPath calculates the critical path through the pipeline
func (pv *PipelineValidator) calculateCriticalPath(graph *ComponentGraph) {
	if len(graph.TopologyOrder) == 0 {
		return
	}

	// For now, use a simple approach - the longest path through the graph
	// In a real implementation, this would consider actual execution times
	distances := make(map[string]int)
	predecessors := make(map[string]string)

	// Initialize distances
	for name := range graph.Nodes {
		distances[name] = 0
	}

	// Process nodes in topological order
	for _, current := range graph.TopologyOrder {
		if node, ok := graph.Nodes[current]; ok {
			for _, dependent := range node.Dependents {
				newDistance := distances[current] + 1
				if newDistance > distances[dependent] {
					distances[dependent] = newDistance
					predecessors[dependent] = current
				}
			}
		}
	}

	// Find the node with maximum distance
	maxDistance := 0
	endNode := ""
	for name, distance := range distances {
		if distance > maxDistance {
			maxDistance = distance
			endNode = name
		}
	}

	// Reconstruct critical path
	criticalPath := make([]string, 0)
	current := endNode
	for current != "" {
		criticalPath = append([]string{current}, criticalPath...)
		current = predecessors[current]
	}

	graph.CriticalPath = criticalPath
}

// String methods for validation types
func (vet ValidationErrorType) String() string {
	switch vet {
	case ValidationErrorTypeUnknown:
		return "UNKNOWN"
	case ValidationErrorTypeMissingComponent:
		return "MISSING_COMPONENT"
	case ValidationErrorTypeMissingPort:
		return "MISSING_PORT"
	case ValidationErrorTypeTypeMismatch:
		return "TYPE_MISMATCH"
	case ValidationErrorTypeCycle:
		return "CYCLE"
	case ValidationErrorTypeDisconnectedComponent:
		return "DISCONNECTED_COMPONENT"
	case ValidationErrorTypeInvalidConfiguration:
		return "INVALID_CONFIGURATION"
	case ValidationErrorTypeResourceLimit:
		return "RESOURCE_LIMIT"
	default:
		return "UNKNOWN"
	}
}

func (vwt ValidationWarningType) String() string {
	switch vwt {
	case ValidationWarningTypeUnused:
		return "UNUSED"
	case ValidationWarningTypePerformance:
		return "PERFORMANCE"
	case ValidationWarningTypeConfiguration:
		return "CONFIGURATION"
	default:
		return "UNKNOWN"
	}
}