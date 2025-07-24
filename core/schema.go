package core

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// BaseSchema provides a default implementation of the Schema interface
type BaseSchema struct {
	schemaType    reflect.Type
	description   string
	constraints   []Constraint
	migrationFunc func(interface{}, Schema) (interface{}, error)
}

// NewBaseSchema creates a new BaseSchema for the given type
func NewBaseSchema(t reflect.Type, description string) *BaseSchema {
	return &BaseSchema{
		schemaType:  t,
		description: description,
		constraints: make([]Constraint, 0),
	}
}

// Validate validates data against the schema
func (s *BaseSchema) Validate(data interface{}) error {
	if data == nil {
		return fmt.Errorf("data cannot be nil")
	}

	dataType := reflect.TypeOf(data)
	if !s.isCompatibleType(dataType, s.schemaType) {
		return fmt.Errorf("type mismatch: expected %s, got %s", s.schemaType, dataType)
	}

	// Apply constraints
	for _, constraint := range s.constraints {
		if err := constraint.Validate(data); err != nil {
			return fmt.Errorf("constraint validation failed: %w", err)
		}
	}

	return nil
}

// Compatible checks if this schema is compatible with another schema
func (s *BaseSchema) Compatible(other Schema) bool {
	otherBase, ok := other.(*BaseSchema)
	if !ok {
		return false
	}

	return s.isCompatibleType(s.schemaType, otherBase.schemaType)
}

// Migrate migrates data from this schema to a target schema
func (s *BaseSchema) Migrate(data interface{}, targetSchema Schema) (interface{}, error) {
	if s.migrationFunc != nil {
		return s.migrationFunc(data, targetSchema)
	}

	// Default migration: if types are compatible, return as-is
	if s.Compatible(targetSchema) {
		return data, nil
	}

	return nil, fmt.Errorf("no migration path from %s to target schema", s.schemaType)
}

// JSONSchema returns a JSON Schema representation
func (s *BaseSchema) JSONSchema() string {
	schema := map[string]interface{}{
		"type":        s.getJSONType(s.schemaType),
		"description": s.description,
	}

	if len(s.constraints) > 0 {
		constraints := make([]string, len(s.constraints))
		for i, c := range s.constraints {
			constraints[i] = c.Description()
		}
		schema["constraints"] = constraints
	}

	jsonBytes, _ := json.MarshalIndent(schema, "", "  ")
	return string(jsonBytes)
}

// AddConstraint adds a constraint to the schema
func (s *BaseSchema) AddConstraint(constraint Constraint) {
	s.constraints = append(s.constraints, constraint)
}

// SetMigrationFunc sets a custom migration function
func (s *BaseSchema) SetMigrationFunc(fn func(interface{}, Schema) (interface{}, error)) {
	s.migrationFunc = fn
}

// isCompatibleType checks if two types are compatible
func (s *BaseSchema) isCompatibleType(t1, t2 reflect.Type) bool {
	if t1 == t2 {
		return true
	}

	// Handle interface{} compatibility
	if t1.Kind() == reflect.Interface && t1.NumMethod() == 0 {
		return true
	}
	if t2.Kind() == reflect.Interface && t2.NumMethod() == 0 {
		return true
	}

	// Handle pointer types
	if t1.Kind() == reflect.Ptr && t2.Kind() == reflect.Ptr {
		return s.isCompatibleType(t1.Elem(), t2.Elem())
	}

	// Handle slice types
	if t1.Kind() == reflect.Slice && t2.Kind() == reflect.Slice {
		return s.isCompatibleType(t1.Elem(), t2.Elem())
	}

	// Handle map types
	if t1.Kind() == reflect.Map && t2.Kind() == reflect.Map {
		return s.isCompatibleType(t1.Key(), t2.Key()) && s.isCompatibleType(t1.Elem(), t2.Elem())
	}

	return false
}

// getJSONType converts Go type to JSON Schema type
func (s *BaseSchema) getJSONType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		 reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Map, reflect.Struct:
		return "object"
	default:
		return "any"
	}
}

// Common constraint implementations

// NotNilConstraint ensures the value is not nil
type NotNilConstraint struct{}

func (c *NotNilConstraint) Validate(data interface{}) error {
	if data == nil {
		return fmt.Errorf("value cannot be nil")
	}
	return nil
}

func (c *NotNilConstraint) Description() string {
	return "value must not be nil"
}

// StringLengthConstraint validates string length
type StringLengthConstraint struct {
	MinLength int
	MaxLength int
}

func (c *StringLengthConstraint) Validate(data interface{}) error {
	str, ok := data.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", data)
	}

	length := len(str)
	if c.MinLength > 0 && length < c.MinLength {
		return fmt.Errorf("string length %d is less than minimum %d", length, c.MinLength)
	}
	if c.MaxLength > 0 && length > c.MaxLength {
		return fmt.Errorf("string length %d exceeds maximum %d", length, c.MaxLength)
	}

	return nil
}

func (c *StringLengthConstraint) Description() string {
	if c.MinLength > 0 && c.MaxLength > 0 {
		return fmt.Sprintf("string length must be between %d and %d", c.MinLength, c.MaxLength)
	} else if c.MinLength > 0 {
		return fmt.Sprintf("string length must be at least %d", c.MinLength)
	} else if c.MaxLength > 0 {
		return fmt.Sprintf("string length must be at most %d", c.MaxLength)
	}
	return "string length constraint"
}

// NumericRangeConstraint validates numeric ranges
type NumericRangeConstraint struct {
	Min interface{}
	Max interface{}
}

func (c *NumericRangeConstraint) Validate(data interface{}) error {
	val := reflect.ValueOf(data)
	if !val.Type().Comparable() {
		return fmt.Errorf("value type %T is not comparable", data)
	}

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal := val.Int()
		if c.Min != nil {
			if minVal, ok := c.Min.(int64); ok && intVal < minVal {
				return fmt.Errorf("value %d is less than minimum %d", intVal, minVal)
			}
		}
		if c.Max != nil {
			if maxVal, ok := c.Max.(int64); ok && intVal > maxVal {
				return fmt.Errorf("value %d exceeds maximum %d", intVal, maxVal)
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal := val.Uint()
		if c.Min != nil {
			if minVal, ok := c.Min.(uint64); ok && uintVal < minVal {
				return fmt.Errorf("value %d is less than minimum %d", uintVal, minVal)
			}
		}
		if c.Max != nil {
			if maxVal, ok := c.Max.(uint64); ok && uintVal > maxVal {
				return fmt.Errorf("value %d exceeds maximum %d", uintVal, maxVal)
			}
		}
	case reflect.Float32, reflect.Float64:
		floatVal := val.Float()
		if c.Min != nil {
			if minVal, ok := c.Min.(float64); ok && floatVal < minVal {
				return fmt.Errorf("value %f is less than minimum %f", floatVal, minVal)
			}
		}
		if c.Max != nil {
			if maxVal, ok := c.Max.(float64); ok && floatVal > maxVal {
				return fmt.Errorf("value %f exceeds maximum %f", floatVal, maxVal)
			}
		}
	default:
		return fmt.Errorf("numeric constraint cannot be applied to type %T", data)
	}

	return nil
}

func (c *NumericRangeConstraint) Description() string {
	parts := make([]string, 0, 2)
	if c.Min != nil {
		parts = append(parts, fmt.Sprintf("minimum: %v", c.Min))
	}
	if c.Max != nil {
		parts = append(parts, fmt.Sprintf("maximum: %v", c.Max))
	}
	if len(parts) == 0 {
		return "numeric range constraint"
	}
	return strings.Join(parts, ", ")
}

// RegexConstraint validates strings against a regular expression
type RegexConstraint struct {
	Pattern string
	regex   interface{} // Would be *regexp.Regexp in real implementation
}

func (c *RegexConstraint) Validate(data interface{}) error {
	str, ok := data.(string)
	if !ok {
		return fmt.Errorf("regex constraint can only be applied to strings, got %T", data)
	}

	// In a real implementation, we would compile and use the regex
	// For now, just check if the pattern is not empty
	if c.Pattern == "" {
		return fmt.Errorf("regex pattern is empty")
	}

	// Placeholder validation - in real implementation would use regexp.MatchString
	if len(str) == 0 && c.Pattern != ".*" {
		return fmt.Errorf("string does not match pattern %s", c.Pattern)
	}

	return nil
}

func (c *RegexConstraint) Description() string {
	return fmt.Sprintf("must match pattern: %s", c.Pattern)
}