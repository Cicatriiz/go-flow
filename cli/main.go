package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/forrest/go-flow/components"
	"github.com/forrest/go-flow/core"
	"github.com/forrest/go-flow/visualization"
)

// This is a placeholder for a real component from your library
type ExampleSource struct {
	core.BaseComponent
}

func main() {
	format := flag.String("T", "dot", "Output format (dot, svg, png)")
	example := flag.String("example", "simple", "Example pipeline to generate (simple, file)")
	flag.Parse()

	var p *core.Pipeline

	switch *example {
	case "simple":
		p = createSimplePipeline()
	case "file":
		p = create_file_processing_pipeline()
	default:
		fmt.Printf("Unknown example: %s\n", *example)
		os.Exit(1)
	}

	dot := visualization.ToDOT(p)

	if *format == "dot" {
		fmt.Println(dot)
	} else {
		cmd := exec.Command("dot", "-T"+*format)
		cmd.Stdin = strings.NewReader(dot)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error running dot: %v\n", err)
			os.Exit(1)
		}
	}
}

func createSimplePipeline() *core.Pipeline {
	p := core.NewPipeline("simple-example")
	p.AddComponent("source", components.NewStringSource("hello world"))
	p.AddComponent("upper", components.NewUpperCase())
	p.AddComponent("sink", components.NewStringSink())
	core.Connect[string](p, "source", "output", "upper", "input")
	core.Connect[string](p, "upper", "output", "sink", "input")
	return p
}

func create_file_processing_pipeline() *core.Pipeline {
	p := core.NewPipeline("file-processing-example")
	p.AddComponent("reader", components.NewFileReader("input.txt"))
	p.AddComponent("grepper", components.NewGrep("go"))
	p.AddComponent("upper", components.NewUpperCase())
	p.AddComponent("writer", components.NewFileWriter("output.txt"))

	core.Connect[string](p, "reader", "output", "grepper", "input")
	core.Connect[string](p, "grepper", "output", "upper", "input")
	core.Connect[string](p, "upper", "output", "writer", "input")
	return p
}
