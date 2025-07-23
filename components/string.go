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
	c.Outputs = []core.Port{
		&core.BasePort{PortName: "output", PortType: reflect.TypeOf("")},
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
	c.Inputs = []core.Port{
		&core.BasePort{PortName: "input", PortType: reflect.TypeOf("")},
	}
	return c
}

// Process consumes the string.
func (c *StringSink) Process(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	input, ok := inputs["input"].(string)
	if !ok {
		return nil, fmt.Errorf("input is not a string")
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
	c.Inputs = []core.Port{
		&core.BasePort{PortName: "input", PortType: reflect.TypeOf("")},
	}
	c.Outputs = []core.Port{
		&core.BasePort{PortName: "output", PortType: reflect.TypeOf("")},
	}
	return c
}

// Process converts the input string to uppercase.
func (c *UpperCase) Process(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	input, ok := inputs["input"].(string)
	if !ok {
		return nil, fmt.Errorf("input is not a string")
	}

	outputs := make(map[string]interface{})
	outputs["output"] = strings.ToUpper(input)
	return outputs, nil
}
