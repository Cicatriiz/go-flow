package core

import (
	"context"
	"reflect"
	"testing"
	"time"
)

// TestEnhancedInterfaces demonstrates the enhanced core interfaces
func TestEnhancedInterfaces(t *testing.T) {
	// Test Schema functionality
	t.Run("Schema", func(t *testing.T) {
		schema := NewBaseSchema(reflect.TypeOf(""), "Test string schema")
		schema.AddConstraint(&NotNilConstraint{})
		schema.AddConstraint(&StringLengthConstraint{MinLength: 1, MaxLength: 100})

		// Test valid data
		if err := schema.Validate("hello"); err != nil {
			t.Errorf("Expected valid string to pass validation, got: %v", err)
		}

		// Test invalid data (nil)
		if err := schema.Validate(nil); err == nil {
			t.Error("Expected nil to fail validation")
		}

		// Test invalid data (wrong type)
		if err := schema.Validate(123); err == nil {
			t.Error("Expected integer to fail string validation")
		}

		// Test JSON Schema generation
		jsonSchema := schema.JSONSchema()
		if jsonSchema == "" {
			t.Error("Expected JSON schema to be generated")
		}
	})

	// Test Error handling
	t.Run("PipelineError", func(t *testing.T) {
		err := NewPipelineError("test error", "test-component", ValidationError, Error, true)
		err.WithContext("key", "value")

		if err.Component() != "test-component" {
			t.Errorf("Expected component 'test-component', got '%s'", err.Component())
		}

		if err.ErrorType() != ValidationError {
			t.Errorf("Expected ValidationError, got %v", err.ErrorType())
		}

		if err.Severity() != Error {
			t.Errorf("Expected Error severity, got %v", err.Severity())
		}

		if !err.Recoverable() {
			t.Error("Expected error to be recoverable")
		}

		context := err.Context()
		if context["key"] != "value" {
			t.Errorf("Expected context key 'value', got '%v'", context["key"])
		}
	})

	// Test Error Handler
	t.Run("ErrorHandler", func(t *testing.T) {
		handler := NewDefaultErrorHandler(3)
		
		// Test critical error
		criticalErr := NewPipelineError("critical", "comp", RuntimeError, Critical, false)
		action := handler.HandleError(context.Background(), criticalErr)
		if action != Abort {
			t.Errorf("Expected Abort for critical error, got %v", action)
		}

		// Test recoverable error
		recoverableErr := NewPipelineError("recoverable", "comp", RuntimeError, Error, true)
		action = handler.HandleError(context.Background(), recoverableErr)
		if action != Retry {
			t.Errorf("Expected Retry for recoverable error, got %v", action)
		}
	})

	// Test Circuit Breaker
	t.Run("CircuitBreaker", func(t *testing.T) {
		cb := NewCircuitBreaker(2, 1, 100*time.Millisecond)

		// Test successful execution
		result, err := cb.Execute(context.Background(), func() (interface{}, error) {
			return "success", nil
		})
		if err != nil {
			t.Errorf("Expected successful execution, got error: %v", err)
		}
		if result != "success" {
			t.Errorf("Expected 'success', got '%v'", result)
		}

		// Test circuit breaker state
		if cb.State() != Closed {
			t.Errorf("Expected Closed state, got %v", cb.State())
		}
	})

	// Test Enhanced Port
	t.Run("EnhancedPort", func(t *testing.T) {
		schema := NewBaseSchema(reflect.TypeOf(""), "String port")
		port := &BasePort{
			PortName:         "test-port",
			PortType:         reflect.TypeOf(""),
			IsRequired:       true,
			PortDescription:  "Test port description",
			PortSchema:       schema,
			PortDefaultValue: "default",
			PortExamples:     []interface{}{"example1", "example2"},
			PortDocumentation: "Detailed documentation",
		}

		if port.Name() != "test-port" {
			t.Errorf("Expected port name 'test-port', got '%s'", port.Name())
		}

		if !port.Required() {
			t.Error("Expected port to be required")
		}

		if port.Schema() != schema {
			t.Error("Expected schema to match")
		}

		if port.DefaultValue() != "default" {
			t.Errorf("Expected default value 'default', got '%v'", port.DefaultValue())
		}

		examples := port.Examples()
		if len(examples) != 2 {
			t.Errorf("Expected 2 examples, got %d", len(examples))
		}
	})
}

// TestComponentLifecycle tests the enhanced component lifecycle
func TestComponentLifecycle(t *testing.T) {
	component := &BaseComponent{
		ComponentName:        "test-component",
		ComponentDescription: "Test component for lifecycle",
		ComponentVersion:     "1.0.0",
		ComponentTags:        []string{"test", "lifecycle"},
	}

	ctx := context.Background()

	// Test initialization
	if err := component.Initialize(ctx); err != nil {
		t.Errorf("Expected successful initialization, got: %v", err)
	}

	// Test health check
	if err := component.HealthCheck(ctx); err != nil {
		t.Errorf("Expected successful health check, got: %v", err)
	}

	// Test validation
	if err := component.Validate(); err != nil {
		t.Errorf("Expected successful validation, got: %v", err)
	}

	// Test metadata
	if component.Description() != "Test component for lifecycle" {
		t.Errorf("Expected description to match, got '%s'", component.Description())
	}

	if component.Version() != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", component.Version())
	}

	tags := component.Tags()
	if len(tags) != 2 || tags[0] != "test" || tags[1] != "lifecycle" {
		t.Errorf("Expected tags [test, lifecycle], got %v", tags)
	}

	// Test cleanup
	if err := component.Cleanup(ctx); err != nil {
		t.Errorf("Expected successful cleanup, got: %v", err)
	}
}

// TestPipelineEnhancements tests the enhanced pipeline functionality
func TestPipelineEnhancements(t *testing.T) {
	pipeline := NewPipeline("test-pipeline")
	
	// Add a test component
	component := &BaseComponent{
		ComponentName:        "test-comp",
		ComponentDescription: "Test component",
		ComponentVersion:     "1.0.0",
	}
	pipeline.AddComponent("test", component)

	ctx := context.Background()

	// Test pipeline initialization
	if err := pipeline.Initialize(ctx); err != nil {
		t.Errorf("Expected successful pipeline initialization, got: %v", err)
	}

	// Test pipeline health check
	if err := pipeline.HealthCheck(ctx); err != nil {
		t.Errorf("Expected successful pipeline health check, got: %v", err)
	}

	// Test pipeline validation
	if err := pipeline.Validate(); err != nil {
		t.Errorf("Expected successful pipeline validation, got: %v", err)
	}

	// Test pipeline metadata
	if pipeline.Description() == "" {
		t.Error("Expected pipeline description to be set")
	}

	if pipeline.Version() != "1.0.0" {
		t.Errorf("Expected pipeline version '1.0.0', got '%s'", pipeline.Version())
	}

	tags := pipeline.Tags()
	if len(tags) == 0 {
		t.Error("Expected pipeline to have tags")
	}

	// Test pipeline cleanup
	if err := pipeline.Cleanup(ctx); err != nil {
		t.Errorf("Expected successful pipeline cleanup, got: %v", err)
	}
}

// TestConstraints tests various constraint implementations
func TestConstraints(t *testing.T) {
	t.Run("NotNilConstraint", func(t *testing.T) {
		constraint := &NotNilConstraint{}
		
		if err := constraint.Validate("not nil"); err != nil {
			t.Errorf("Expected non-nil value to pass, got: %v", err)
		}
		
		if err := constraint.Validate(nil); err == nil {
			t.Error("Expected nil value to fail")
		}
	})

	t.Run("StringLengthConstraint", func(t *testing.T) {
		constraint := &StringLengthConstraint{MinLength: 2, MaxLength: 10}
		
		if err := constraint.Validate("hello"); err != nil {
			t.Errorf("Expected valid string to pass, got: %v", err)
		}
		
		if err := constraint.Validate("a"); err == nil {
			t.Error("Expected short string to fail")
		}
		
		if err := constraint.Validate("this is too long"); err == nil {
			t.Error("Expected long string to fail")
		}
	})

	t.Run("NumericRangeConstraint", func(t *testing.T) {
		constraint := &NumericRangeConstraint{Min: int64(1), Max: int64(100)}
		
		if err := constraint.Validate(int64(50)); err != nil {
			t.Errorf("Expected valid number to pass, got: %v", err)
		}
		
		if err := constraint.Validate(int64(0)); err == nil {
			t.Error("Expected number below min to fail")
		}
		
		if err := constraint.Validate(int64(101)); err == nil {
			t.Error("Expected number above max to fail")
		}
	})
}

// TestErrorCollector tests the error collection functionality
func TestErrorCollector(t *testing.T) {
	collector := NewErrorCollector()
	
	err1 := NewPipelineError("error 1", "comp1", ValidationError, Error, true)
	err2 := NewPipelineError("error 2", "comp2", RuntimeError, Warning, false)
	err3 := NewPipelineError("error 3", "comp1", ConfigurationError, Critical, false)
	
	collector.Collect(err1)
	collector.Collect(err2)
	collector.Collect(err3)
	
	// Test total count
	if collector.Count() != 3 {
		t.Errorf("Expected 3 errors, got %d", collector.Count())
	}
	
	// Test get by component
	comp1Errors := collector.GetErrorsByComponent("comp1")
	if len(comp1Errors) != 2 {
		t.Errorf("Expected 2 errors for comp1, got %d", len(comp1Errors))
	}
	
	// Test get by severity
	criticalErrors := collector.GetErrorsBySeverity(Critical)
	if len(criticalErrors) != 1 {
		t.Errorf("Expected 1 critical error, got %d", len(criticalErrors))
	}
	
	// Test clear
	collector.Clear()
	if collector.Count() != 0 {
		t.Errorf("Expected 0 errors after clear, got %d", collector.Count())
	}
}