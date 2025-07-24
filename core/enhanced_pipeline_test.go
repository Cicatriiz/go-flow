package core

import (
	"context"
	"reflect"
	"testing"
	"time"
)

// Test component for validation testing
type TestValidationComponent struct {
	*BaseComponent
}

func NewTestValidationComponent(name string) *TestValidationComponent {
	return &TestValidationComponent{
		BaseComponent: &BaseComponent{
			ComponentName:        name,
			ComponentDescription: "Test component for validation",
			ComponentVersion:     "1.0.0",
			ComponentTags:        []string{"test"},
			Inputs: []Port{
				&BasePort{
					PortName:        "input",
					PortType:        reflect.TypeOf(""),
					IsRequired:      true,
					PortDescription: "String input",
				},
			},
			Outputs: []Port{
				&BasePort{
					PortName:        "output",
					PortType:        reflect.TypeOf(""),
					IsRequired:      false,
					PortDescription: "String output",
				},
			},
		},
	}
}

func (c *TestValidationComponent) Process(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	input := inputs["input"].(string)
	return map[string]interface{}{
		"output": "processed_" + input,
	}, nil
}



func TestEnhancedPipelineCreation(t *testing.T) {
	// Test creating pipeline with default config
	pipeline := NewPipeline("test_pipeline")
	
	if pipeline.Name() != "test_pipeline" {
		t.Errorf("Expected pipeline name 'test_pipeline', got '%s'", pipeline.Name())
	}
	
	if pipeline.GetVersion() != "1.0.0" {
		t.Errorf("Expected default version '1.0.0', got '%s'", pipeline.GetVersion())
	}
	
	if pipeline.GetConfig() == nil {
		t.Error("Expected default config to be set")
	}
	
	if pipeline.GetContext() == nil {
		t.Error("Expected pipeline context to be initialized")
	}
	
	if pipeline.GetErrorCollector() == nil {
		t.Error("Expected error collector to be initialized")
	}
}

func TestEnhancedPipelineConfiguration(t *testing.T) {
	config := &PipelineConfig{
		MaxConcurrency:    5,
		Timeout:          10 * time.Second,
		MemoryLimit:      512 * 1024 * 1024,
		CPULimit:         0.5,
		MetricsEnabled:   true,
		TracingEnabled:   true,
		LogLevel:         "DEBUG",
		StrictValidation: true,
		AllowCycles:      false,
		DefaultBufferSize: 50,
		MaxBufferSize:    500,
	}
	
	pipeline := NewPipelineWithConfig("configured_pipeline", config)
	
	if pipeline.GetConfig().MaxConcurrency != 5 {
		t.Errorf("Expected MaxConcurrency 5, got %d", pipeline.GetConfig().MaxConcurrency)
	}
	
	if pipeline.GetConfig().Timeout != 10*time.Second {
		t.Errorf("Expected Timeout 10s, got %v", pipeline.GetConfig().Timeout)
	}
	
	if !pipeline.GetConfig().TracingEnabled {
		t.Error("Expected TracingEnabled to be true")
	}
}

func TestEnhancedPipelineMetadata(t *testing.T) {
	pipeline := NewPipeline("metadata_test")
	
	// Test setting and getting metadata
	pipeline.SetMetadata("author", "test_user")
	pipeline.SetMetadata("version", "2.0.0")
	
	if pipeline.GetMetadata("author") != "test_user" {
		t.Errorf("Expected author 'test_user', got '%v'", pipeline.GetMetadata("author"))
	}
	
	metadata := pipeline.GetAllMetadata()
	if len(metadata) != 2 {
		t.Errorf("Expected 2 metadata entries, got %d", len(metadata))
	}
}

func TestEnhancedConnectionCreation(t *testing.T) {
	pipeline := NewPipeline("connection_test")
	
	comp1 := NewTestValidationComponent("comp1")
	comp2 := NewTestValidationComponent("comp2")
	
	pipeline.AddComponent("comp1", comp1)
	pipeline.AddComponent("comp2", comp2)
	
	// Test basic connection
	Connect[string](pipeline, "comp1", "output", "comp2", "input")
	
	connections := pipeline.GetConnections()
	if len(connections) != 1 {
		t.Errorf("Expected 1 connection, got %d", len(connections))
	}
	
	conn := connections[0]
	if conn.FromComponent != "comp1" || conn.FromPort != "output" ||
	   conn.ToComponent != "comp2" || conn.ToPort != "input" {
		t.Error("Connection details are incorrect")
	}
	
	if conn.BufferSize != pipeline.GetConfig().DefaultBufferSize {
		t.Errorf("Expected buffer size %d, got %d", pipeline.GetConfig().DefaultBufferSize, conn.BufferSize)
	}
	
	if conn.Name == "" {
		t.Error("Expected connection name to be set")
	}
}

func TestConnectionWithTransform(t *testing.T) {
	pipeline := NewPipeline("transform_test")
	
	comp1 := NewTestValidationComponent("comp1")
	comp2 := NewTestValidationComponent("comp2")
	
	pipeline.AddComponent("comp1", comp1)
	pipeline.AddComponent("comp2", comp2)
	
	// Create a transform
	transform := NewStringToUpperTransform()
	
	// Connect with transform
	pipeline.ConnectWithTransform("comp1", "output", "comp2", "input", transform)
	
	connections := pipeline.GetConnections()
	if len(connections) != 1 {
		t.Errorf("Expected 1 connection, got %d", len(connections))
	}
	
	conn := connections[0]
	if conn.Transform == nil {
		t.Error("Expected transform to be set")
	}
	
	if conn.Transform.Name() != "string_to_upper" {
		t.Errorf("Expected transform name 'string_to_upper', got '%s'", conn.Transform.Name())
	}
}

func TestConnectionWithBackpressure(t *testing.T) {
	pipeline := NewPipeline("backpressure_test")
	
	comp1 := NewTestValidationComponent("comp1")
	comp2 := NewTestValidationComponent("comp2")
	
	pipeline.AddComponent("comp1", comp1)
	pipeline.AddComponent("comp2", comp2)
	
	// Create backpressure config
	backpressure := &BackpressureConfig{
		Strategy:   BackpressureBuffer,
		BufferSize: 200,
		DropPolicy: DropOldest,
		Timeout:    5 * time.Second,
		MaxRetries: 3,
	}
	
	// Connect with backpressure
	pipeline.ConnectWithBackpressure("comp1", "output", "comp2", "input", backpressure)
	
	connections := pipeline.GetConnections()
	if len(connections) != 1 {
		t.Errorf("Expected 1 connection, got %d", len(connections))
	}
	
	conn := connections[0]
	if conn.Backpressure == nil {
		t.Error("Expected backpressure config to be set")
	}
	
	if conn.Backpressure.Strategy != BackpressureBuffer {
		t.Errorf("Expected BackpressureBuffer strategy, got %v", conn.Backpressure.Strategy)
	}
	
	if conn.Backpressure.BufferSize != 200 {
		t.Errorf("Expected buffer size 200, got %d", conn.Backpressure.BufferSize)
	}
}

func TestPipelineValidation(t *testing.T) {
	pipeline := NewPipeline("validation_test")
	
	// Create components with different input requirements
	comp1 := &TestValidationComponent{
		BaseComponent: &BaseComponent{
			ComponentName:        "comp1",
			ComponentDescription: "Test component for validation",
			ComponentVersion:     "1.0.0",
			ComponentTags:        []string{"test"},
			Inputs: []Port{
				&BasePort{
					PortName:        "input",
					PortType:        reflect.TypeOf(""),
					IsRequired:      false, // Make this optional for pipeline start
					PortDescription: "String input",
				},
			},
			Outputs: []Port{
				&BasePort{
					PortName:        "output",
					PortType:        reflect.TypeOf(""),
					IsRequired:      false,
					PortDescription: "String output",
				},
			},
		},
	}
	
	comp2 := NewTestValidationComponent("comp2")
	comp3 := NewTestValidationComponent("comp3")
	
	pipeline.AddComponent("comp1", comp1)
	pipeline.AddComponent("comp2", comp2)
	pipeline.AddComponent("comp3", comp3)
	
	// Create valid connections
	Connect[string](pipeline, "comp1", "output", "comp2", "input")
	Connect[string](pipeline, "comp2", "output", "comp3", "input")
	
	// Test comprehensive validation
	result := pipeline.ValidateComprehensive()
	
	if !result.Valid {
		t.Errorf("Expected pipeline to be valid, but got errors: %v", result.Errors)
	}
	
	if result.ComponentGraph == nil {
		t.Error("Expected component graph to be built")
	}
	
	// Check topology order
	if len(result.ComponentGraph.TopologyOrder) != 3 {
		t.Errorf("Expected 3 components in topology order, got %d", len(result.ComponentGraph.TopologyOrder))
	}
	
	// Check critical path
	if len(result.ComponentGraph.CriticalPath) == 0 {
		t.Error("Expected critical path to be calculated")
	}
}

func TestPipelineValidationErrors(t *testing.T) {
	pipeline := NewPipeline("error_test")
	
	comp1 := NewTestValidationComponent("comp1")
	comp2 := NewTestValidationComponent("comp2")
	
	pipeline.AddComponent("comp1", comp1)
	pipeline.AddComponent("comp2", comp2)
	
	// Create invalid connection (wrong port name)
	pipeline.connections = append(pipeline.connections, Connection{
		FromComponent: "comp1",
		FromPort:      "invalid_port",
		ToComponent:   "comp2",
		ToPort:        "input",
		BufferSize:    100,
		Name:          "invalid_connection",
	})
	
	// Test validation
	result := pipeline.ValidateComprehensive()
	
	if result.Valid {
		t.Error("Expected pipeline to be invalid due to missing port")
	}
	
	// Check for specific error
	foundError := false
	for _, err := range result.Errors {
		if err.Type == ValidationErrorTypeMissingPort {
			foundError = true
			break
		}
	}
	
	if !foundError {
		t.Error("Expected to find ValidationErrorTypeMissingPort error")
	}
}

func TestCycleDetection(t *testing.T) {
	pipeline := NewPipeline("cycle_test")
	
	comp1 := NewTestValidationComponent("comp1")
	comp2 := NewTestValidationComponent("comp2")
	comp3 := NewTestValidationComponent("comp3")
	
	pipeline.AddComponent("comp1", comp1)
	pipeline.AddComponent("comp2", comp2)
	pipeline.AddComponent("comp3", comp3)
	
	// Create a cycle: comp1 -> comp2 -> comp3 -> comp1
	Connect[string](pipeline, "comp1", "output", "comp2", "input")
	Connect[string](pipeline, "comp2", "output", "comp3", "input")
	Connect[string](pipeline, "comp3", "output", "comp1", "input")
	
	// Test cycle detection
	result := pipeline.ValidateComprehensive()
	
	if result.Valid {
		t.Error("Expected pipeline to be invalid due to cycle")
	}
	
	// Check for cycle error
	foundCycleError := false
	for _, err := range result.Errors {
		if err.Type == ValidationErrorTypeCycle {
			foundCycleError = true
			break
		}
	}
	
	if !foundCycleError {
		t.Error("Expected to find ValidationErrorTypeCycle error")
	}
}

func TestConfigurationValidation(t *testing.T) {
	// Test invalid configuration
	invalidConfig := &PipelineConfig{
		MaxConcurrency:    -1, // Invalid
		Timeout:          -5 * time.Second, // Invalid
		DefaultBufferSize: -10, // Invalid
		MaxBufferSize:    5,   // Less than default
	}
	
	pipeline := NewPipelineWithConfig("config_test", invalidConfig)
	
	result := pipeline.ValidateComprehensive()
	
	if result.Valid {
		t.Error("Expected pipeline to be invalid due to configuration errors")
	}
	
	// Should have multiple configuration errors
	configErrors := 0
	for _, err := range result.Errors {
		if err.Type == ValidationErrorTypeInvalidConfiguration {
			configErrors++
		}
	}
	
	if configErrors == 0 {
		t.Error("Expected to find configuration validation errors")
	}
}

func TestTopologyOrder(t *testing.T) {
	pipeline := NewPipeline("topology_test")
	
	comp1 := NewTestValidationComponent("comp1")
	comp2 := NewTestValidationComponent("comp2")
	comp3 := NewTestValidationComponent("comp3")
	comp4 := NewTestValidationComponent("comp4")
	
	pipeline.AddComponent("comp1", comp1)
	pipeline.AddComponent("comp2", comp2)
	pipeline.AddComponent("comp3", comp3)
	pipeline.AddComponent("comp4", comp4)
	
	// Create connections: comp1 -> comp2, comp1 -> comp3, comp2 -> comp4, comp3 -> comp4
	Connect[string](pipeline, "comp1", "output", "comp2", "input")
	Connect[string](pipeline, "comp1", "output", "comp3", "input")
	Connect[string](pipeline, "comp2", "output", "comp4", "input")
	Connect[string](pipeline, "comp3", "output", "comp4", "input")
	
	topologyOrder, err := pipeline.GetTopologyOrder()
	if err != nil {
		t.Errorf("Failed to get topology order: %v", err)
	}
	
	if len(topologyOrder) != 4 {
		t.Errorf("Expected 4 components in topology order, got %d", len(topologyOrder))
	}
	
	// comp1 should be first, comp4 should be last
	if topologyOrder[0] != "comp1" {
		t.Errorf("Expected comp1 to be first, got %s", topologyOrder[0])
	}
	
	if topologyOrder[len(topologyOrder)-1] != "comp4" {
		t.Errorf("Expected comp4 to be last, got %s", topologyOrder[len(topologyOrder)-1])
	}
}

func TestCriticalPath(t *testing.T) {
	pipeline := NewPipeline("critical_path_test")
	
	comp1 := NewTestValidationComponent("comp1")
	comp2 := NewTestValidationComponent("comp2")
	comp3 := NewTestValidationComponent("comp3")
	
	pipeline.AddComponent("comp1", comp1)
	pipeline.AddComponent("comp2", comp2)
	pipeline.AddComponent("comp3", comp3)
	
	// Create linear path: comp1 -> comp2 -> comp3
	Connect[string](pipeline, "comp1", "output", "comp2", "input")
	Connect[string](pipeline, "comp2", "output", "comp3", "input")
	
	criticalPath, err := pipeline.GetCriticalPath()
	if err != nil {
		t.Errorf("Failed to get critical path: %v", err)
	}
	
	expectedPath := []string{"comp1", "comp2", "comp3"}
	if len(criticalPath) != len(expectedPath) {
		t.Errorf("Expected critical path length %d, got %d", len(expectedPath), len(criticalPath))
	}
	
	for i, expected := range expectedPath {
		if i >= len(criticalPath) || criticalPath[i] != expected {
			t.Errorf("Expected critical path component %d to be %s, got %s", i, expected, criticalPath[i])
		}
	}
}

func TestPipelineErrorCollection(t *testing.T) {
	pipeline := NewPipeline("error_collection_test")
	
	// Add a pipeline error
	pipelineError := NewPipelineError("Test error", "test_component", RuntimeError, Error, true)
	pipeline.AddPipelineError(pipelineError)
	
	errors := pipeline.GetPipelineErrors()
	if len(errors) != 1 {
		t.Errorf("Expected 1 pipeline error, got %d", len(errors))
	}
	
	if errors[0].Component() != "test_component" {
		t.Errorf("Expected error component 'test_component', got '%s'", errors[0].Component())
	}
	
	// Check error collector
	collectedErrors := pipeline.GetErrorCollector().GetErrors()
	if len(collectedErrors) != 1 {
		t.Errorf("Expected 1 collected error, got %d", len(collectedErrors))
	}
}