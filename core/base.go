package core

import (
	"context"
	"reflect"
)

// BasePort provides a default implementation of the Port interface.
type BasePort struct {
	PortName        string
	PortType        reflect.Type
	IsRequired      bool
	PortDescription string
}

// Name returns the name of the port.
func (p *BasePort) Name() string {
	return p.PortName
}

// Type returns the data type of the port.
func (p *BasePort) Type() reflect.Type {
	return p.PortType
}

// Required indicates whether the port must be connected.
func (p *BasePort) Required() bool {
	return p.IsRequired
}

// Description provides a human-readable description of the port.
func (p *BasePort) Description() string {
	return p.PortDescription
}

// BaseComponent provides a default implementation of the Component interface.
// It can be embedded in other structs to easily create new components.
type BaseComponent struct {
	ComponentName string
	Inputs        []Port
	Outputs       []Port
}

// Name returns the unique identifier of the component.
func (c *BaseComponent) Name() string {
	return c.ComponentName
}

// SetName sets the name of the component.
func (c *BaseComponent) SetName(name string) {
	c.ComponentName = name
}

// InputPorts returns the list of input ports for the component.
func (c *BaseComponent) InputPorts() []Port {
	return c.Inputs
}

// OutputPorts returns the list of output ports for the component.
func (c *BaseComponent) OutputPorts() []Port {
	return c.Outputs
}

// Process is a placeholder implementation.
// Components embedding BaseComponent should override this method.
func (c *BaseComponent) Process(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Default implementation does nothing.
	return make(map[string]interface{}), nil
}

// Validate is a placeholder implementation.
// Components embedding BaseComponent can override this method for custom validation.
func (c *BaseComponent) Validate() error {
	// Default implementation has no validation.
	return nil
}
