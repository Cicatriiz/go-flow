package core

import (
	"context"
	"reflect"
	"testing"
	"time"
)

// TestEnhancedIntegration demonstrates the complete enhanced interface system
func TestEnhancedIntegration(t *testing.T) {
	ctx := context.Background()

	// Create a test component with enhanced features
	component := &TestEnhancedComponent{
		BaseComponent: BaseComponent{
			ComponentName:        "test-component",
			ComponentDescription: "A test component demonstrating enhanced features",
			ComponentVersion:     "2.0.0",
			ComponentTags:        []string{"test", "enhanced", "demo"},
		},
		initialized: false,
		healthy:     true,
	}

	// Set up enhanced ports with schemas and constraints
	stringSchema := NewBaseSchema(reflect.TypeOf(""), "Enhanced string schema")
	stringSchema.AddConstraint(&NotNilConstraint{})
	stringSchema.AddConstraint(&StringLengthConstraint{MinLength: 1, MaxLength: 100})

	component.Inputs = []Port{
		&BasePort{
			PortName:         "input",
			PortType:         reflect.TypeOf(""),
			IsRequired:       true,
			PortDescription:  "Enhanced input port",
			PortSchema:       stringSchema,
			PortDefaultValue: "default",
			PortExamples:     []interface{}{"hello", "world"},
			PortDocumentation: "This port accepts string input with validation",
		},
	}

	component.Outputs = []Port{
		&BasePort{
			PortName:         "output",
			PortType:         reflect.TypeOf(""),
			PortDescription:  "Enhanced output port",
			PortSchema:       stringSchema,
			PortExamples:     []interface{}{"HELLO", "WORLD"},
			PortDocumentation: "This port outputs processed string data",
		},
	}

	// Test lifecycle management
	t.Run("Lifecycle", func(t *testing.T) {
		// Test initialization
		if err := component.Initialize(ctx); err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}
		if !component.initialized {
			t.Error("Component should be initialized")
		}

		// Test health check
		if err := component.HealthCheck(ctx); err != nil {
			t.Fatalf("HealthCheck failed: %v", err)
		}

		// Test cleanup
		if err := component.Cleanup(ctx); err != nil {
			t.Fatalf("Cleanup failed: %v", err)
		}
	})

	// Test enhanced metadata
	t.Run("Metadata", func(t *testing.T) {
		if component.Description() != "A test component demonstrating enhanced features" {
			t.Error("Description mismatch")
		}
		if component.Version() != "2.0.0" {
			t.Error("Version mismatch")
		}
		tags := component.Tags()
		if len(tags) != 3 || tags[0] != "test" {
			t.Error("Tags mismatch")
		}
	})

	// Test port enhancements
	t.Run("EnhancedPorts", func(t *testing.T) {
		inputPort := component.InputPorts()[0]
		
		// Test schema validation
		schema := inputPort.Schema()
		if err := schema.Validate("valid string"); err != nil {
			t.Errorf("Valid string should pass validation: %v", err)
		}
		
		if err := schema.Validate(""); err == nil {
			t.Error("Empty string should fail validation due to length constraint")
		}

		// Test port metadata
		if inputPort.DefaultValue() != "default" {
			t.Error("Default value mismatch")
		}
		
		examples := inputPort.Examples()
		if len(examples) != 2 {
			t.Error("Examples count mismatch")
		}
		
		if inputPort.Documentation() == "" {
			t.Error("Documentation should not be empty")
		}
	})

	// Test error handling
	t.Run("ErrorHandling", func(t *testing.T) {
		// Create a pipeline error
		pipelineErr := NewPipelineError(
			"test error",
			component.Name(),
			ValidationError,
			Warning,
			true,
		).WithContext("test_key", "test_value")

		if pipelineErr.Component() != component.Name() {
			t.Error("Component name mismatch in error")
		}
		if pipelineErr.ErrorType() != ValidationError {
			t.Error("Error type mismatch")
		}
		if pipelineErr.Severity() != Warning {
			t.Error("Severity mismatch")
		}
		if !pipelineErr.Recoverable() {
			t.Error("Error should be recoverable")
		}

		context := pipelineErr.Context()
		if context["test_key"] != "test_value" {
			t.Error("Context value mismatch")
		}

		// Test error handler
		handler := NewDefaultErrorHandler(3)
		action := handler.HandleError(ctx, pipelineErr)
		if action != Continue {
			t.Error("Warning errors should continue")
		}
	})

	// Test circuit breaker
	t.Run("CircuitBreaker", func(t *testing.T) {
		cb := NewCircuitBreaker(2, 1, time.Millisecond*100)
		
		// Test successful execution
		result, err := cb.Execute(ctx, func() (interface{}, error) {
			return "success", nil
		})
		if err != nil {
			t.Errorf("Circuit breaker execution failed: %v", err)
		}
		if result != "success" {
			t.Error("Result mismatch")
		}
		
		if cb.State() != Closed {
			t.Error("Circuit should be closed after success")
		}
	})
}

// TestEnhancedComponent is a test component that implements the enhanced Component interface
type TestEnhancedComponent struct {
	BaseComponent
	initialized bool
	healthy     bool
}

func (c *TestEnhancedComponent) Initialize(ctx context.Context) error {
	c.initialized = true
	return nil
}

func (c *TestEnhancedComponent) HealthCheck(ctx context.Context) error {
	if !c.healthy {
		return NewPipelineError("component unhealthy", c.Name(), RuntimeError, Warning, true)
	}
	return nil
}

func (c *TestEnhancedComponent) Cleanup(ctx context.Context) error {
	c.initialized = false
	return nil
}

func (c *TestEnhancedComponent) Process(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	if !c.initialized {
		return nil, NewPipelineError("component not initialized", c.Name(), ConfigurationError, Error, false)
	}

	input, ok := inputs["input"].(string)
	if !ok {
		return nil, NewPipelineError("invalid input type", c.Name(), ValidationError, Error, false)
	}

	// Validate input using port schema
	inputPort := c.InputPorts()[0]
	if err := inputPort.Schema().Validate(input); err != nil {
		return nil, NewPipelineError("input validation failed", c.Name(), ValidationError, Error, false).WithOriginalError(err)
	}

	outputs := make(map[string]interface{})
	outputs["output"] = "processed: " + input
	return outputs, nil
}

// TestPipelineWithEnhancedComponents tests a complete pipeline with enhanced components
func TestPipelineWithEnhancedComponents(t *testing.T) {
	ctx := context.Background()

	// Create pipeline
	pipeline := NewPipeline("enhanced-test-pipeline")

	// Add enhanced components
	source := &TestEnhancedComponent{
		BaseComponent: BaseComponent{
			ComponentName:        "source",
			ComponentDescription: "Enhanced source component",
			ComponentVersion:     "1.0.0",
			ComponentTags:        []string{"source", "test"},
		},
		initialized: false,
		healthy:     true,
	}

	// Set up source outputs
	stringSchema := NewBaseSchema(reflect.TypeOf(""), "String output")
	source.Outputs = []Port{
		&BasePort{
			PortName:     "output",
			PortType:     reflect.TypeOf(""),
			PortSchema:   stringSchema,
			PortExamples: []interface{}{"test output"},
		},
	}

	sink := &TestEnhancedComponent{
		BaseComponent: BaseComponent{
			ComponentName:        "sink",
			ComponentDescription: "Enhanced sink component",
			ComponentVersion:     "1.0.0",
			ComponentTags:        []string{"sink", "test"},
		},
		initialized: false,
		healthy:     true,
	}

	// Set up sink inputs
	sink.Inputs = []Port{
		&BasePort{
			PortName:   "input",
			PortType:   reflect.TypeOf(""),
			IsRequired: true,
			PortSchema: stringSchema,
		},
	}

	// Add components to pipeline
	pipeline.AddComponent("source", source)
	pipeline.AddComponent("sink", sink)

	// Connect components
	Connect[string](pipeline, "source", "output", "sink", "input")

	// Test pipeline lifecycle
	t.Run("PipelineLifecycle", func(t *testing.T) {
		// Initialize pipeline
		if err := pipeline.Initialize(ctx); err != nil {
			t.Fatalf("Pipeline initialization failed: %v", err)
		}

		// Validate pipeline
		if err := pipeline.Validate(); err != nil {
			t.Fatalf("Pipeline validation failed: %v", err)
		}

		// Health check pipeline
		if err := pipeline.HealthCheck(ctx); err != nil {
			t.Fatalf("Pipeline health check failed: %v", err)
		}

		// Cleanup pipeline
		if err := pipeline.Cleanup(ctx); err != nil {
			t.Fatalf("Pipeline cleanup failed: %v", err)
		}
	})

	// Test pipeline metadata
	t.Run("PipelineMetadata", func(t *testing.T) {
		description := pipeline.Description()
		if description == "" {
			t.Error("Pipeline description should not be empty")
		}

		version := pipeline.Version()
		if version != "1.0.0" {
			t.Error("Pipeline version mismatch")
		}

		tags := pipeline.Tags()
		if len(tags) == 0 {
			t.Error("Pipeline should have tags")
		}
	})

	// Test error collection
	t.Run("ErrorCollection", func(t *testing.T) {
		collector := NewErrorCollector()

		// Collect some test errors
		err1 := NewPipelineError("error 1", "comp1", ValidationError, Error, true)
		err2 := NewPipelineError("error 2", "comp2", RuntimeError, Warning, false)
		err3 := NewPipelineError("error 3", "comp1", NetworkError, Critical, false)

		collector.Collect(err1)
		collector.Collect(err2)
		collector.Collect(err3)

		// Test error retrieval
		allErrors := collector.GetErrors()
		if len(allErrors) != 3 {
			t.Errorf("Expected 3 errors, got %d", len(allErrors))
		}

		comp1Errors := collector.GetErrorsByComponent("comp1")
		if len(comp1Errors) != 2 {
			t.Errorf("Expected 2 errors for comp1, got %d", len(comp1Errors))
		}

		criticalErrors := collector.GetErrorsBySeverity(Critical)
		if len(criticalErrors) != 1 {
			t.Errorf("Expected 1 critical error, got %d", len(criticalErrors))
		}

		// Test count
		if collector.Count() != 3 {
			t.Errorf("Expected count 3, got %d", collector.Count())
		}

		// Test clear
		collector.Clear()
		if collector.Count() != 0 {
			t.Error("Collector should be empty after clear")
		}
	})
}

// TestSchemaCompatibility tests schema compatibility and migration
func TestSchemaCompatibility(t *testing.T) {
	// Create two compatible schemas
	schema1 := NewBaseSchema(reflect.TypeOf(""), "String schema 1")
	schema2 := NewBaseSchema(reflect.TypeOf(""), "String schema 2")

	if !schema1.Compatible(schema2) {
		t.Error("String schemas should be compatible")
	}

	// Test migration
	data := "test data"
	migrated, err := schema1.Migrate(data, schema2)
	if err != nil {
		t.Errorf("Migration failed: %v", err)
	}
	if migrated != data {
		t.Error("Migrated data should be unchanged for compatible schemas")
	}

	// Test JSON schema generation
	jsonSchema := schema1.JSONSchema()
	if jsonSchema == "" {
		t.Error("JSON schema should not be empty")
	}
}

// TestConstraintValidation tests various constraint implementations
func TestConstraintValidation(t *testing.T) {
	t.Run("StringLengthConstraint", func(t *testing.T) {
		constraint := &StringLengthConstraint{MinLength: 3, MaxLength: 10}

		// Valid string
		if err := constraint.Validate("hello"); err != nil {
			t.Errorf("Valid string should pass: %v", err)
		}

		// Too short
		if err := constraint.Validate("hi"); err == nil {
			t.Error("Short string should fail validation")
		}

		// Too long
		if err := constraint.Validate("this is too long"); err == nil {
			t.Error("Long string should fail validation")
		}

		// Wrong type
		if err := constraint.Validate(123); err == nil {
			t.Error("Non-string should fail validation")
		}
	})

	t.Run("NumericRangeConstraint", func(t *testing.T) {
		constraint := &NumericRangeConstraint{Min: int64(0), Max: int64(100)}

		// Valid number
		if err := constraint.Validate(int64(50)); err != nil {
			t.Errorf("Valid number should pass: %v", err)
		}

		// Too small
		if err := constraint.Validate(int64(-1)); err == nil {
			t.Error("Small number should fail validation")
		}

		// Too large
		if err := constraint.Validate(int64(101)); err == nil {
			t.Error("Large number should fail validation")
		}
	})

	t.Run("RegexConstraint", func(t *testing.T) {
		constraint := &RegexConstraint{Pattern: "test.*"}

		// Test basic validation (simplified)
		if err := constraint.Validate("test string"); err != nil {
			t.Errorf("Valid string should pass: %v", err)
		}

		// Wrong type
		if err := constraint.Validate(123); err == nil {
			t.Error("Non-string should fail validation")
		}
	})
}