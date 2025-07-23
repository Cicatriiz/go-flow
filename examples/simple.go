package main

import (
	"context"
	"fmt"
	"reflect"

	"github.com/forrest/go-flow/core"
	"github.com/forrest/go-flow/execution"
	"github.com/forrest/go-flow/visualization"
)

// HelloWorldComponent is a simple component that prints a message.
type HelloWorldComponent struct {
	core.BaseComponent
}

// NewHelloWorldComponent creates a new HelloWorldComponent.
func NewHelloWorldComponent() *HelloWorldComponent {
	c := &HelloWorldComponent{}
	c.Inputs = []core.Port{
		&core.BasePort{PortName: "input", PortType: reflect.TypeOf("")},
	}
	c.Outputs = []core.Port{
		&core.BasePort{PortName: "output", PortType: reflect.TypeOf("")},
	}
	return c
}

func (c *HelloWorldComponent) Process(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	if msg, ok := inputs["input"]; ok {
		fmt.Printf("%s: received '%s'\n", c.Name(), msg)
	} else {
		fmt.Printf("%s: sending 'Hello'\n", c.Name())
	}

	outputs := make(map[string]interface{})
	outputs["output"] = "Hello"
	return outputs, nil
}

func simpleExample() {
	// Define the pipeline
	p := core.NewPipeline("simple-example")
	p.AddComponent("hello", NewHelloWorldComponent())
	p.AddComponent("world", NewHelloWorldComponent())
	core.Connect[string](p, "hello", "output", "world", "input")

	// Set the execution engine
	p.SetEngine(execution.NewConcurrentEngine())

	// Generate and print the DOT representation
	dot := visualization.ToDOT(p)
	fmt.Println(dot)

	// Run the pipeline
	if err := p.Run(context.Background()); err != nil {
		fmt.Println("Error running pipeline:", err)
	}
}
