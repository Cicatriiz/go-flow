package components

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/forrest/go-flow/core"
)

// StringSource is a component that produces a string.
type StringSource struct {
	core.BaseComponent
	Data string
}

// NewStringSource creates a new StringSource component.
func NewStringSource(data string) *StringSource {
	c := &StringSource{Data: data}
	c.ComponentDescription = "Produces a string value"
	c.ComponentVersion = "1.0.0"
	c.ComponentTags = []string{"source", "string"}
	
	stringSchema := core.NewBaseSchema(reflect.TypeOf(""), "String output")
	c.Outputs = []core.Port{
		&core.BasePort{
			PortName:         "output",
			PortType:         reflect.TypeOf(""),
			PortDescription:  "String output",
			PortSchema:       stringSchema,
			PortExamples:     []interface{}{"hello", "world"},
			PortDocumentation: "Outputs the configured string value",
		},
	}
	return c
}

// Process produces the string.
func (c *StringSource) Process(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	outputs := make(map[string]interface{})
	outputs["output"] = c.Data
	return outputs, nil
}

// StringSink is a component that consumes a string.
type StringSink struct {
	core.BaseComponent
}

// NewStringSink creates a new StringSink component.
func NewStringSink() *StringSink {
	c := &StringSink{}
	c.ComponentDescription = "Consumes and prints a string value"
	c.ComponentVersion = "1.0.0"
	c.ComponentTags = []string{"sink", "string", "output"}
	
	stringSchema := core.NewBaseSchema(reflect.TypeOf(""), "String input")
	stringSchema.AddConstraint(&core.NotNilConstraint{})
	c.Inputs = []core.Port{
		&core.BasePort{
			PortName:         "input",
			PortType:         reflect.TypeOf(""),
			IsRequired:       true,
			PortDescription:  "String input to print",
			PortSchema:       stringSchema,
			PortExamples:     []interface{}{"hello world", "test message"},
			PortDocumentation: "Accepts any string value and prints it to stdout",
		},
	}
	return c
}

// Process consumes the string.
func (c *StringSink) Process(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
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
	fmt.Printf("StringSink: %s\n", input)
	return nil, nil
}

// UpperCase is a component that converts a string to uppercase.
type UpperCase struct {
	core.BaseComponent
}

// NewUpperCase creates a new UpperCase component.
func NewUpperCase() *UpperCase {
	c := &UpperCase{}
	c.ComponentDescription = "Converts string input to uppercase"
	c.ComponentVersion = "1.0.0"
	c.ComponentTags = []string{"transform", "string", "case"}
	
	stringSchema := core.NewBaseSchema(reflect.TypeOf(""), "String data")
	stringSchema.AddConstraint(&core.NotNilConstraint{})
	stringSchema.AddConstraint(&core.StringLengthConstraint{MinLength: 1, MaxLength: 10000})
	
	c.Inputs = []core.Port{
		&core.BasePort{
			PortName:         "input",
			PortType:         reflect.TypeOf(""),
			IsRequired:       true,
			PortDescription:  "String to convert to uppercase",
			PortSchema:       stringSchema,
			PortExamples:     []interface{}{"hello", "world", "test"},
			PortDocumentation: "Input string that will be converted to uppercase",
		},
	}
	c.Outputs = []core.Port{
		&core.BasePort{
			PortName:         "output",
			PortType:         reflect.TypeOf(""),
			PortDescription:  "Uppercase string output",
			PortSchema:       stringSchema,
			PortExamples:     []interface{}{"HELLO", "WORLD", "TEST"},
			PortDocumentation: "The input string converted to uppercase",
		},
	}
	return c
}

// Process converts the input string to uppercase.
func (c *UpperCase) Process(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
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

	outputs := make(map[string]interface{})
	outputs["output"] = strings.ToUpper(input)
	return outputs, nil
}
