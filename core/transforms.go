package core

import (
	"context"
	"fmt"
	"strings"
)

// BaseDataTransform provides a basic implementation of DataTransform
type BaseDataTransform struct {
	name        string
	description string
	transformFn func(ctx context.Context, data interface{}) (interface{}, error)
}

// NewBaseDataTransform creates a new base data transform
func NewBaseDataTransform(name, description string, fn func(ctx context.Context, data interface{}) (interface{}, error)) *BaseDataTransform {
	return &BaseDataTransform{
		name:        name,
		description: description,
		transformFn: fn,
	}
}

// Transform applies the transformation function to the data
func (t *BaseDataTransform) Transform(ctx context.Context, data interface{}) (interface{}, error) {
	if t.transformFn == nil {
		return data, nil // Pass-through if no transform function
	}
	return t.transformFn(ctx, data)
}

// Name returns the name of the transform
func (t *BaseDataTransform) Name() string {
	return t.name
}

// Description returns the description of the transform
func (t *BaseDataTransform) Description() string {
	return t.description
}

// Common transform implementations

// IdentityTransform passes data through unchanged
type IdentityTransform struct {
	*BaseDataTransform
}

// NewIdentityTransform creates a new identity transform
func NewIdentityTransform() *IdentityTransform {
	return &IdentityTransform{
		BaseDataTransform: NewBaseDataTransform(
			"identity",
			"Passes data through unchanged",
			func(ctx context.Context, data interface{}) (interface{}, error) {
				return data, nil
			},
		),
	}
}

// StringToUpperTransform converts string data to uppercase
type StringToUpperTransform struct {
	*BaseDataTransform
}

// NewStringToUpperTransform creates a new string-to-upper transform
func NewStringToUpperTransform() *StringToUpperTransform {
	return &StringToUpperTransform{
		BaseDataTransform: NewBaseDataTransform(
			"string_to_upper",
			"Converts string data to uppercase",
			func(ctx context.Context, data interface{}) (interface{}, error) {
				if str, ok := data.(string); ok {
					return strings.ToUpper(str), nil
				}
				return nil, fmt.Errorf("expected string, got %T", data)
			},
		),
	}
}

// TypeConversionTransform converts data between types
type TypeConversionTransform struct {
	*BaseDataTransform
	targetType string
}

// NewTypeConversionTransform creates a new type conversion transform
func NewTypeConversionTransform(targetType string) *TypeConversionTransform {
	return &TypeConversionTransform{
		BaseDataTransform: NewBaseDataTransform(
			fmt.Sprintf("convert_to_%s", targetType),
			fmt.Sprintf("Converts data to %s", targetType),
			func(ctx context.Context, data interface{}) (interface{}, error) {
				switch targetType {
				case "string":
					return fmt.Sprintf("%v", data), nil
				default:
					return data, nil
				}
			},
		),
		targetType: targetType,
	}
}