package core

import (
	"context"
	"reflect"
)

// BasePort provides a default implementation of the Port interface.
type BasePort struct {
	PortName         string
	PortType         reflect.Type
	IsRequired       bool
	PortDescription  string
	PortSchema       Schema
	PortDefaultValue interface{}
	PortConstraints  []Constraint
	PortExamples     []interface{}
	PortDocumentation string
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

// Schema returns the schema for data validation.
func (p *BasePort) Schema() Schema {
	return p.PortSchema
}

// DefaultValue returns the default value for the port.
func (p *BasePort) DefaultValue() interface{} {
	return p.PortDefaultValue
}

// Constraints returns the validation constraints for the port.
func (p *BasePort) Constraints() []Constraint {
	return p.PortConstraints
}

// Examples returns example values for the port.
func (p *BasePort) Examples() []interface{} {
	return p.PortExamples
}

// Documentation returns detailed documentation for the port.
func (p *BasePort) Documentation() string {
	return p.PortDocumentation
}

// BaseComponent provides a default implementation of the Component interface.
// It can be embedded in other structs to easily create new components.
type BaseComponent struct {
	ComponentName        string
	ComponentDescription string
	ComponentVersion     string
	ComponentTags        []string
	Inputs               []Port
	Outputs              []Port
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

// HealthCheck performs a health check on the component.
// Components embedding BaseComponent can override this method for custom health checks.
func (c *BaseComponent) HealthCheck(ctx context.Context) error {
	// Default implementation assumes the component is healthy.
	return nil
}

// Initialize initializes the component.
// Components embedding BaseComponent can override this method for custom initialization.
func (c *BaseComponent) Initialize(ctx context.Context) error {
	// Default implementation does nothing.
	return nil
}

// Cleanup cleans up resources used by the component.
// Components embedding BaseComponent can override this method for custom cleanup.
func (c *BaseComponent) Cleanup(ctx context.Context) error {
	// Default implementation does nothing.
	return nil
}

// Description returns a human-readable description of the component.
func (c *BaseComponent) Description() string {
	return c.ComponentDescription
}

// Version returns the version of the component.
func (c *BaseComponent) Version() string {
	if c.ComponentVersion == "" {
		return "1.0.0"
	}
	return c.ComponentVersion
}

// Tags returns the tags associated with the component.
func (c *BaseComponent) Tags() []string {
	return c.ComponentTags
}
