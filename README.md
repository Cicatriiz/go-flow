# Go-Flow

Go-Flow is a specialized Go library for building, visualizing, and executing complex data and processing pipelines using a declarative, flow-based programming model. This library provides compile-time type safety, native Go graph definition, comprehensive validation, and automatic visualization for data transformation pipelines.

## Core Features

*   **Declarative, Type-Safe Pipelines:** Define complex data pipelines in pure Go, with the compiler enforcing type safety across all connections.
*   **Compile-Time Validation:** The library automatically detects cycles, and ensures that all connections are valid.
*   **Automatic Visualization:** Generate detailed visual representations of your pipelines in DOT, SVG, or PNG format using the new CLI tool.
*   **Flexible Execution:** The library provides both concurrent and sequential execution engines to suit different needs.
*   **Extensible Component Model:** The refined `Component` interface and `BaseComponent` make it easy to create new, reusable pipeline components.

### Pipeline Definition & Validation
*   **Enhanced Pipeline Structure:** Rich pipeline metadata, versioning, and configuration management
*   **Comprehensive Validation:** Multi-layer validation including component health, connection compatibility, cycle detection, and resource limits
*   **Graph Analysis:** Automatic topology ordering, critical path calculation, and dependency analysis
*   **Type-Safe Connections:** Compile-time type safety with generic connection functions

### Advanced Error Handling & Resilience
*   **Structured Error Management:** Hierarchical error types with severity levels and recovery strategies
*   **Circuit Breaker Pattern:** Built-in circuit breakers for component resilience
*   **Error Collection:** Centralized error aggregation and analysis
*   **Retry Policies:** Configurable retry mechanisms with exponential backoff

### Data Transformation & Flow Control
*   **Connection-Level Transforms:** Apply data transformations between components
*   **Backpressure Management:** Configurable backpressure strategies and buffer management
*   **Resource Management:** Memory and CPU limits with monitoring
*   **Flexible Execution:** Multiple execution engines for different performance needs

### Monitoring & Observability
*   **Pipeline Metrics:** Real-time performance monitoring and statistics
*   **Component Lifecycle:** Initialize, health check, and cleanup management
*   **Execution Context:** Rich runtime context with variables and tags
*   **Automatic Visualization:** Generate detailed visual representations in DOT, SVG, or PNG format

## Installation

```bash
go get github.com/forrest/go-flow
```

## Quick Start

Here is a quick example of how to define and run an enhanced file-processing pipeline:

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/forrest/go-flow/components"
	"github.com/forrest/go-flow/core"
	"github.com/forrest/go-flow/execution"
)

func main() {
	// Create pipeline with enhanced configuration
	config := &core.PipelineConfig{
		MaxConcurrency:    5,
		Timeout:          30 * time.Second,
		MetricsEnabled:   true,
		StrictValidation: true,
		DefaultBufferSize: 100,
	}
	
	p := core.NewPipelineWithConfig("file-processing-example", config)
	p.SetVersion("1.0.0")
	p.SetDescription("Enhanced file processing pipeline with validation")
	p.SetMetadata("author", "go-flow-user")

	// Add components
	p.AddComponent("reader", components.NewFileReader("input.txt"))
	p.AddComponent("grepper", components.NewGrep("go"))
	p.AddComponent("upper", components.NewUpperCase())
	p.AddComponent("writer", components.NewFileWriter("output.txt"))

	// Create type-safe connections
	core.Connect[string](p, "reader", "output", "grepper", "input")
	core.Connect[string](p, "grepper", "output", "upper", "input")
	
	// Connect with data transformation
	transform := core.NewStringToUpperTransform()
	p.ConnectWithTransform("upper", "output", "writer", "input", transform)

	// Validate pipeline before execution
	if result := p.ValidateComprehensive(); !result.Valid {
		fmt.Printf("Pipeline validation failed: %v\n", result.Errors)
		return
	}

	// Set execution engine and run
	p.SetEngine(execution.NewConcurrentEngine())
	if err := p.Run(context.Background()); err != nil {
		fmt.Println("Error running pipeline:", err)
	}
	
	// Access pipeline metrics
	metrics := p.GetContext().Metrics
	fmt.Printf("Pipeline processed %d items\n", metrics.TotalProcessed)
}
```

## Advanced Features

### Pipeline Validation

Go-Flow provides comprehensive validation to ensure pipeline correctness:

```go
// Perform comprehensive validation
result := pipeline.ValidateComprehensive()

if !result.Valid {
    for _, err := range result.Errors {
        fmt.Printf("Error: %s - %s\n", err.Type, err.Message)
    }
}

// Access validation insights
graph, _ := pipeline.GetComponentGraph()
fmt.Printf("Topology order: %v\n", graph.TopologyOrder)
fmt.Printf("Critical path: %v\n", graph.CriticalPath)
```

### Error Handling & Circuit Breakers

Built-in resilience patterns for robust pipeline execution:

```go
// Configure retry policy
retryPolicy := &core.RetryPolicy{
    MaxRetries:    3,
    InitialDelay:  100 * time.Millisecond,
    BackoffFactor: 2.0,
}

// Create circuit breaker
circuitBreaker := core.NewCircuitBreaker(5, 3, 30*time.Second)

// Handle pipeline errors
errorHandler := core.NewDefaultErrorHandler(3)
```

### Data Transformations

Apply transformations between pipeline components:

```go
// Built-in transforms
upperTransform := core.NewStringToUpperTransform()
typeTransform := core.NewTypeConversionTransform("string")

// Connect with transformation
pipeline.ConnectWithTransform("source", "output", "target", "input", upperTransform)

// Custom backpressure configuration
backpressure := &core.BackpressureConfig{
    Strategy:   core.BackpressureBuffer,
    BufferSize: 200,
    DropPolicy: core.DropOldest,
}
pipeline.ConnectWithBackpressure("source", "output", "target", "input", backpressure)
```

## CLI Usage

Go-Flow includes a powerful CLI tool for visualizing your pipelines.

**Generate a DOT file:**

```bash
go run ./cli -example file -T dot
```

**Generate an SVG image:**

```bash
go run ./cli -example file -T svg > pipeline.svg
```

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.
