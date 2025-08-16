package mcpargs

import (
	"errors"
	"fmt"
	"reflect"
)

var ErrUnmarshalArguments = fmt.Errorf("failed to unmarshal arguments")

// Unmarshal populates the struct pointed to by v with values from the arguments map.
// Field names in the struct are converted to snake_case to match keys in the arguments map.
func Unmarshal(arguments map[string]any, v any) error { //nolint:cyclop,funlen,gocognit,gocyclo // Unfortunately reflection is complex.
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("%w: destination must be a non-nil pointer to a struct", ErrUnmarshalArguments)
	}

	// Dereference the pointer
	rv = rv.Elem()

	// Ensure v points to a struct
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("%w: destination must be a pointer to a struct, got pointer to %s", ErrUnmarshalArguments, rv.Kind())
	}

	// Get the type of the struct
	rt := rv.Type()

	var errs error

	// Iterate through each field in the struct
	for i := range rt.NumField() {
		field := rt.Field(i)

		// Check for unexported fields
		if !field.IsExported() {
			errs = errors.Join(errs, fmt.Errorf("%w: struct contains unexported field %q, which is not supported", ErrUnmarshalArguments, field.Name))
			continue
		}

		snakeName := toSnakeCase(field.Name)

		value, ok := arguments[snakeName]
		if !ok {
			if required := field.Tag.Get("mcp_required"); required == "true" {
				errs = errors.Join(errs, fmt.Errorf("%w: missing required field %q", ErrUnmarshalArguments, field.Name))
				continue
			}

			continue
		}

		// Get the field value we want to set
		fieldValue := rv.Field(i)

		// First, check if the field implements the Unmarshaler interface
		// We need to create a pointer to the field since Unmarshaler is only implemented on pointer types
		if fieldValue.CanAddr() {
			ptrToField := fieldValue.Addr()

			// Check if the pointer implements Unmarshaler
			if unmarshaler, ok := ptrToField.Interface().(Unmarshaler); ok {
				if err := unmarshaler.Unmarshal(value); err != nil {
					errs = errors.Join(errs, fmt.Errorf("%w: failed to unmarshal field %q: %w", ErrUnmarshalArguments, field.Name, err))
				}

				continue
			}
		}

		//nolint:exhaustive
		switch field.Type.Kind() {
		case reflect.Struct, reflect.Chan, reflect.Func, reflect.Map, reflect.Slice, reflect.Array,
			reflect.Interface, reflect.Ptr:
			errs = errors.Join(errs, fmt.Errorf("%w: field %q has unsupported type %s", ErrUnmarshalArguments, field.Name, field.Type.Kind()))
		}

		// Only set the field if it's settable (exported)
		if !fieldValue.CanSet() {
			errs = errors.Join(errs, fmt.Errorf("%w: field %q is not settable", ErrUnmarshalArguments, field.Name))
			continue
		}

		// If not a custom unmarshaler, try to convert the value to the field's type and set it
		if err := setValue(fieldValue, value); err != nil {
			errs = errors.Join(errs, fmt.Errorf("%w: failed to set field %q: %w", ErrUnmarshalArguments, field.Name, err))
		}
	}

	return errs
}

// setValue sets the given value to the target reflect.Value, with type conversion if needed
//
//nolint:err113 // Errors are wrapped in Unmarshal().
func setValue(target reflect.Value, value any) error { //nolint:cyclop,funlen,gocognit,gocyclo // Unfortunately reflection is complex.
	// If value is nil, we can't do much with it
	if value == nil {
		return fmt.Errorf("cannot set nil value")
	}

	// Get reflect.Value of the value
	rv := reflect.ValueOf(value)

	// Handle the case when the target is directly assignable from the value
	if rv.Type().AssignableTo(target.Type()) {
		target.Set(rv)
		return nil
	}

	//nolint:exhaustive // Unhandled cases return an error.
	switch target.Kind() {
	case reflect.String:
		// Convert value to string
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("cannot convert %T to string", value)
		}

		target.SetString(s)

	case reflect.Bool:
		// Convert value to bool
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("cannot convert %T to bool", value)
		}

		target.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Handle various number types that might need conversion
		var intVal int64

		switch v := value.(type) {
		case int:
			intVal = int64(v)
		case int8:
			intVal = int64(v)
		case int16:
			intVal = int64(v)
		case int32:
			intVal = int64(v)
		case int64:
			intVal = v
		case float32:
			intVal = int64(v)
		case float64:
			intVal = int64(v)
		default:
			return fmt.Errorf("cannot convert %T to int", value)
		}

		target.SetInt(intVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// Handle unsigned integers
		var uintVal uint64
		switch v := value.(type) {
		case uint:
			uintVal = uint64(v)
		case uint8:
			uintVal = uint64(v)
		case uint16:
			uintVal = uint64(v)
		case uint32:
			uintVal = uint64(v)
		case uint64:
			uintVal = v
		case int:
			if v < 0 {
				return fmt.Errorf("cannot convert negative value to unsigned int")
			}

			uintVal = uint64(v)

		case float64:
			if v < 0 {
				return fmt.Errorf("cannot convert negative value to unsigned int")
			}

			uintVal = uint64(v)

		default:
			return fmt.Errorf("cannot convert %T to uint", value)
		}

		target.SetUint(uintVal)

	case reflect.Float32, reflect.Float64:
		// Handle floating point
		var floatVal float64
		switch v := value.(type) {
		case float32:
			floatVal = float64(v)
		case float64:
			floatVal = v
		case int:
			floatVal = float64(v)
		case int64:
			floatVal = float64(v)
		default:
			return fmt.Errorf("cannot convert %T to float", value)
		}

		target.SetFloat(floatVal)

	default:
		return fmt.Errorf("unsupported target type: %s", target.Kind())
	}

	return nil
}

// Unmarshaler is the interface implemented by types that can unmarshal an MCP argument themselves.
// Typically used in combination with the Marshaler interface.
type Unmarshaler interface {
	Unmarshal(v any) error
}
