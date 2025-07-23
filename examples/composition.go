package main

import (
	"context"
	"fmt"

	"github.com/forrest/go-flow/components"
	"github.com/forrest/go-flow/core"
	"github.com/forrest/go-flow/execution"
)

func compositionExample() {
	// Create the sub-pipeline
	subPipeline := core.NewPipeline("sub-pipeline")
	subPipeline.AddComponent("reader", components.NewFileReader("examples/input.txt"))
	subPipeline.AddComponent("upper", components.NewUpperCase())
	core.Connect[string](subPipeline, "reader", "output", "upper", "input")
	subPipeline.SetEngine(execution.NewConcurrentEngine())

	// Create the main pipeline
	mainPipeline := core.NewPipeline("main-pipeline")
	mainPipeline.AddComponent("sub", subPipeline)
	mainPipeline.AddComponent("writer", components.NewFileWriter("examples/composition-output.txt"))
	core.Connect[string](mainPipeline, "sub", "output", "writer", "input")

	// Set the execution engine
	mainPipeline.SetEngine(execution.NewConcurrentEngine())

	// Run the main pipeline
	if err := mainPipeline.Run(context.Background()); err != nil {
		fmt.Println("Error running pipeline:", err)
	}
}
