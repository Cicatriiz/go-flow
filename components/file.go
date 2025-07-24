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
	c.ComponentDescription = "Reads content from a file"
	c.ComponentVersion = "1.0.0"
	c.ComponentTags = []string{"source", "file", "io"}
	
	stringSchema := core.NewBaseSchema(reflect.TypeOf(""), "File content as string")
	c.Outputs = []core.Port{
		&core.BasePort{
			PortName:         "output",
			PortType:         reflect.TypeOf(""),
			PortDescription:  "File content",
			PortSchema:       stringSchema,
			PortExamples:     []interface{}{"file content", "text data"},
			PortDocumentation: "Outputs the entire content of the file as a string",
		},
	}
	return c
}

// Initialize checks if the file exists and is readable
func (c *FileReader) Initialize(ctx context.Context) error {
	if c.Path == "" {
		return core.NewPipelineError("file path is empty", c.Name(), core.ConfigurationError, core.Error, false)
	}
	
	// Check if file exists and is readable
	if _, err := ioutil.ReadFile(c.Path); err != nil {
		return core.NewPipelineError(
			fmt.Sprintf("cannot read file: %v", err),
			c.Name(),
			core.ResourceError,
			core.Error,
			true,
		).WithOriginalError(err)
	}
	
	return nil
}

// HealthCheck verifies the file is still accessible
func (c *FileReader) HealthCheck(ctx context.Context) error {
	if _, err := ioutil.ReadFile(c.Path); err != nil {
		return core.NewPipelineError(
			fmt.Sprintf("file health check failed: %v", err),
			c.Name(),
			core.ResourceError,
			core.Warning,
			true,
		).WithOriginalError(err)
	}
	return nil
}

// Process reads the file and sends its content to the output port.
func (c *FileReader) Process(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(c.Path)
	if err != nil {
		return nil, core.NewPipelineError(
			fmt.Sprintf("failed to read file: %v", err),
			c.Name(),
			core.RuntimeError,
			core.Error,
			true,
		).WithOriginalError(err)
	}

	outputs := make(map[string]interface{})
	outputs["output"] = string(data)
	return outputs, nil
}

// NewFileWriter creates a new FileWriter component.
func NewFileWriter(path string) *FileWriter {
	c := &FileWriter{Path: path}
	c.ComponentDescription = "Writes content to a file"
	c.ComponentVersion = "1.0.0"
	c.ComponentTags = []string{"sink", "file", "io"}
	
	stringSchema := core.NewBaseSchema(reflect.TypeOf(""), "Content to write to file")
	stringSchema.AddConstraint(&core.NotNilConstraint{})
	c.Inputs = []core.Port{
		&core.BasePort{
			PortName:         "input",
			PortType:         reflect.TypeOf(""),
			IsRequired:       true,
			PortDescription:  "Content to write to file",
			PortSchema:       stringSchema,
			PortExamples:     []interface{}{"Hello World", "File content"},
			PortDocumentation: "String content that will be written to the specified file",
		},
	}
	return c
}

// Initialize validates the file path and checks write permissions
func (c *FileWriter) Initialize(ctx context.Context) error {
	if c.Path == "" {
		return core.NewPipelineError("file path is empty", c.Name(), core.ConfigurationError, core.Error, false)
	}
	
	// Test write permissions by attempting to create/write to the file
	if err := ioutil.WriteFile(c.Path, []byte(""), 0644); err != nil {
		return core.NewPipelineError(
			fmt.Sprintf("cannot write to file: %v", err),
			c.Name(),
			core.ResourceError,
			core.Error,
			true,
		).WithOriginalError(err)
	}
	
	return nil
}

// HealthCheck verifies the file is still writable
func (c *FileWriter) HealthCheck(ctx context.Context) error {
	// Test write permissions
	if err := ioutil.WriteFile(c.Path, []byte(""), 0644); err != nil {
		return core.NewPipelineError(
			fmt.Sprintf("file write health check failed: %v", err),
			c.Name(),
			core.ResourceError,
			core.Warning,
			true,
		).WithOriginalError(err)
	}
	return nil
}

// Process writes the input data to a file.
func (c *FileWriter) Process(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	input, ok := inputs["input"].(string)
	if !ok {
		return nil, core.NewPipelineError(
			"input is not a string",
			c.Name(),
			core.ValidationError,
			core.Error,
			false,
		)
	}

	if err := ioutil.WriteFile(c.Path, []byte(input), 0644); err != nil {
		return nil, core.NewPipelineError(
			fmt.Sprintf("failed to write file: %v", err),
			c.Name(),
			core.RuntimeError,
			core.Error,
			true,
		).WithOriginalError(err)
	}
	
	return nil, nil
}

// Grep is a component that filters lines from a string.
type Grep struct {
	core.BaseComponent
	Pattern string
}

// NewGrep creates a new Grep component.
func NewGrep(pattern string) *Grep {
	c := &Grep{Pattern: pattern}
	c.ComponentDescription = "Filters lines containing a specific pattern"
	c.ComponentVersion = "1.0.0"
	c.ComponentTags = []string{"filter", "string", "pattern"}
	
	stringSchema := core.NewBaseSchema(reflect.TypeOf(""), "Text content")
	stringSchema.AddConstraint(&core.NotNilConstraint{})
	
	c.Inputs = []core.Port{
		&core.BasePort{
			PortName:         "input",
			PortType:         reflect.TypeOf(""),
			IsRequired:       true,
			PortDescription:  "Text to filter",
			PortSchema:       stringSchema,
			PortExamples:     []interface{}{"line1\nline2\nline3", "hello\nworld\ntest"},
			PortDocumentation: "Multi-line text input that will be filtered by pattern",
		},
	}
	c.Outputs = []core.Port{
		&core.BasePort{
			PortName:         "output",
			PortType:         reflect.TypeOf(""),
			PortDescription:  "Filtered text output",
			PortSchema:       stringSchema,
			PortExamples:     []interface{}{"matching line1\nmatching line2", "filtered content"},
			PortDocumentation: "Lines from input that contain the specified pattern",
		},
	}
	return c
}

// Initialize validates the pattern
func (c *Grep) Initialize(ctx context.Context) error {
	if c.Pattern == "" {
		return core.NewPipelineError("pattern cannot be empty", c.Name(), core.ConfigurationError, core.Error, false)
	}
	return nil
}

// Process filters the input string.
func (c *Grep) Process(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	input, ok := inputs["input"].(string)
	if !ok {
		return nil, core.NewPipelineError(
			"input is not a string",
			c.Name(),
			core.ValidationError,
			core.Error,
			false,
		)
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
