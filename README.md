# Go-Flow

Go-Flow is a specialized Go library for building, visualizing, and executing complex data and processing pipelines using a declarative, flow-based programming model. This library provides compile-time type safety, native Go graph definition, and automatic visualization for data transformation pipelines.

## Core Features

*   **Declarative, Type-Safe Pipelines:** Define complex data pipelines in pure Go, with the compiler enforcing type safety across all connections.
*   **Compile-Time Validation:** The library automatically detects cycles, and ensures that all connections are valid.
*   **Automatic Visualization:** Generate detailed visual representations of your pipelines in DOT, SVG, or PNG format using the new CLI tool.
*   **Flexible Execution:** The library provides both concurrent and sequential execution engines to suit different needs.
*   **Extensible Component Model:** The refined `Component` interface and `BaseComponent` make it easy to create new, reusable pipeline components.

## Installation

```bash
go get github.com/forrest/go-flow
```

## Quick Start

Here is a quick example of how to define and run a simple file-processing pipeline:

```go
package main

import (
	"context"
	"fmt"

	"github.com/forrest/go-flow/components"
	"github.com/forrest/go-flow/core"
	"github.com/forrest/go-flow/execution"
)

func main() {
	p := core.NewPipeline("file-processing-example")
	p.AddComponent("reader", components.NewFileReader("input.txt"))
	p.AddComponent("grepper", components.NewGrep("go"))
	p.AddComponent("upper", components.NewUpperCase())
	p.AddComponent("writer", components.NewFileWriter("output.txt"))

	core.Connect[string](p, "reader", "output", "grepper", "input")
	core.Connect[string](p, "grepper", "output", "upper", "input")
	core.Connect[string](p, "upper", "output", "writer", "input")

	p.SetEngine(execution.NewConcurrentEngine())

	if err := p.Run(context.Background()); err != nil {
		fmt.Println("Error running pipeline:", err)
	}
}
```

## Usage

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
