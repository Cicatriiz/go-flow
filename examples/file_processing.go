package main

import (
	"context"
	"fmt"

	"github.com/forrest/go-flow/components"
	"github.com/forrest/go-flow/core"
	"github.com/forrest/go-flow/execution"
)

func fileProcessingExample() {
	core.StartMetricsServer(":8080")

	p := core.NewPipeline("file-processing")
	p.AddComponent("reader", components.NewFileReader("examples/input.txt"))
	p.AddComponent("upper", components.NewUpperCase())
	p.AddComponent("writer", components.NewFileWriter("examples/output.txt"))
	core.Connect[string](p, "reader", "output", "upper", "input")
	core.Connect[string](p, "upper", "output", "writer", "input")

	p.SetEngine(execution.NewConcurrentEngine())

	if err := p.Run(context.Background()); err != nil {
		fmt.Println("Error running pipeline:", err)
	}

	// Keep the server running
	select {}
}
