package components

import (
	"context"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/forrest/go-flow/core"
)

// FileReader is a component that reads data from a file.
type FileReader struct {
	core.BaseComponent
	Path string
}

// FileWriter is a component that writes data to a file.
type FileWriter struct {
	core.BaseComponent
	Path string
}

// NewFileReader creates a new FileReader component.
func NewFileReader(path string) *FileReader {
	c := &FileReader{Path: path}
	c.Outputs = []core.Port{
		&core.BasePort{PortName: "output", PortType: reflect.TypeOf("")},
	}
	return c
}

// Process reads the file and sends its content to the output port.
func (c *FileReader) Process(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(c.Path)
	if err != nil {
		return nil, err
	}

	outputs := make(map[string]interface{})
	outputs["output"] = string(data)
	return outputs, nil
}

// NewFileWriter creates a new FileWriter component.
func NewFileWriter(path string) *FileWriter {
	c := &FileWriter{Path: path}
	c.Inputs = []core.Port{
		&core.BasePort{PortName: "input", PortType: reflect.TypeOf("")},
	}
	return c
}

// Process writes the input data to a file.
func (c *FileWriter) Process(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	input, ok := inputs["input"].(string)
	if !ok {
		return nil, fmt.Errorf("input is not a string")
	}

	err := ioutil.WriteFile(c.Path, []byte(input), 0644)
	return nil, err
}

// Grep is a component that filters lines from a string.
type Grep struct {
	core.BaseComponent
	Pattern string
}

// NewGrep creates a new Grep component.
func NewGrep(pattern string) *Grep {
	c := &Grep{Pattern: pattern}
	c.Inputs = []core.Port{
		&core.BasePort{PortName: "input", PortType: reflect.TypeOf("")},
	}
	c.Outputs = []core.Port{
		&core.BasePort{PortName: "output", PortType: reflect.TypeOf("")},
	}
	return c
}

// Process filters the input string.
func (c *Grep) Process(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	input, ok := inputs["input"].(string)
	if !ok {
		return nil, fmt.Errorf("input is not a string")
	}

	var result []string
	for _, line := range strings.Split(input, "\n") {
		if strings.Contains(line, c.Pattern) {
			result = append(result, line)
		}
	}

	outputs := make(map[string]interface{})
	outputs["output"] = strings.Join(result, "\n")
	return outputs, nil
}
