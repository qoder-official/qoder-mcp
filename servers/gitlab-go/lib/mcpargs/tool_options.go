package mcpargs

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/mark3labs/mcp-go/mcp"
)

var ErrMarshalArguments = errors.New("failed to marshal arguments")

// NewTool is a wrapper around mcp.NewTool that automatically marshals the struct to MCP tool options.
// It panics if the marshaling fails.
func NewTool(name string, v any, opts ...mcp.ToolOption) mcp.Tool {
	var toolOpts []mcp.ToolOption

	toolOpts = append(toolOpts, opts...)

	structOpts, err := Marshal(v)
	if err != nil {
		panic(err)
	}

	toolOpts = append(toolOpts, structOpts...)

	return mcp.NewTool(name, toolOpts...)
}

// Marshal accepts a struct and returns a slice of MCP tool options based on the struct's
// fields and tags. Supported fields are string, boolean, int, and float, as
// well as fields implementing the MCPTyper interface.
//
// Required tags:
//   - `mcp_desc`: Describes the field and is required.
//   - `mcp_required`: Indicates if the field is required.
//   - `mcp_enum`: Specifies allowed values for a string field in a comma-separated list.
func Marshal(v any) ([]mcp.ToolOption, error) { //nolint:cyclop,funlen,gocognit,gocyclo // Unfortunately reflection is complex.
	// Get the type of the struct
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: expected struct, got %s", ErrMarshalArguments, t.Kind())
	}

	var (
		toolOpts []mcp.ToolOption
		errs     error
	)

	// Iterate through all fields of the struct
	for i := range t.NumField() {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get field type and name
		fieldType := field.Type
		fieldName := field.Name

		// Convert field name to snake_case for tool option name
		optName := toSnakeCase(fieldName)

		var propOpts []mcp.PropertyOption

		// Set the description.
		description := field.Tag.Get("mcp_desc")
		if description == "" {
			errs = errors.Join(errs, fmt.Errorf(`%w: missing "mcp_desc" tag on field %q`, ErrMarshalArguments, fieldName))
			continue
		}

		propOpts = append(propOpts, mcp.Description(description))

		// Set the required flag if appropriate.
		switch field.Tag.Get("mcp_required") {
		case "true":
			propOpts = append(propOpts, mcp.Required())
		case "false", "":
			// no-op
		default:
			errs = errors.Join(errs, fmt.Errorf(`%w: invalid "mcp_required" tag on field %q: must be "true" or "false"`, ErrMarshalArguments, fieldName))
			continue
		}

		// Set enum values if provided.
		if enumTag := field.Tag.Get("mcp_enum"); enumTag != "" {
			if fieldType.Kind() != reflect.String {
				errs = errors.Join(errs, fmt.Errorf(`%w: invalid "mcp_enum" tag on field %q: must be a string`, ErrMarshalArguments, fieldName))
				continue
			}

			enumValues := strings.Split(enumTag, ",")
			propOpts = append(propOpts, mcp.Enum(enumValues...))
		}

		// If the field implements the MCPTyper interface, use the MCPType method to determine the tool option type.
		var fieldValue reflect.Value
		if reflect.ValueOf(v).Kind() == reflect.Ptr {
			fieldValue = reflect.ValueOf(v).Elem().Field(i)
		} else {
			fieldValue = reflect.ValueOf(v).Field(i)
		}

		if typer, ok := fieldValue.Interface().(Marshaler); ok {
			switch typer.Marshal() {
			case TypeString:
				toolOpts = append(toolOpts, mcp.WithString(optName, propOpts...))
			case TypeBoolean:
				toolOpts = append(toolOpts, mcp.WithBoolean(optName, propOpts...))
			case TypeNumber:
				toolOpts = append(toolOpts, mcp.WithNumber(optName, propOpts...))
			default:
				errs = errors.Join(errs, fmt.Errorf("%w: unsupported field type %v for field %q", ErrMarshalArguments, typer.Marshal(), fieldName))
				continue
			}

			continue
		}

		//nolint:exhaustive // Unhandled cases return an error.
		switch fieldType.Kind() {
		case reflect.String:
			toolOpts = append(toolOpts, mcp.WithString(optName, propOpts...))
		case reflect.Bool:
			toolOpts = append(toolOpts, mcp.WithBoolean(optName, propOpts...))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			toolOpts = append(toolOpts, mcp.WithNumber(optName, propOpts...))
		default:
			errs = errors.Join(errs, fmt.Errorf("%w: unsupported field type %q for field %q", ErrMarshalArguments, fieldType.Kind(), fieldName))
			continue
		}
	}

	if errs != nil {
		return nil, errs
	}

	return toolOpts, nil
}

// MCPType specifies which MCP type a field should be mapped to.
type MCPType int

const (
	TypeString MCPType = iota
	TypeNumber
	TypeBoolean
)

// Marshaler is an interface for types that can return their MCPType.
// This allows the type to be identified without requiring knowledge of the specific type.
// In particular, it allows MCP tool parameters to work with composite types.
// Typically used in combination with the Unmarshaler interface.
type Marshaler interface {
	Marshal() MCPType
}

// toSnakeCase converts a camelCase string to snake_case.
func toSnakeCase(s string) string {
	var result strings.Builder

	// Special case for "IDs", which we want to convert to "_ids", not "_i_ds".
	if strings.HasSuffix(s, "IDs") {
		s = s[:len(s)-3] + "Ids"
	}

	for i, r := range s {
		// Check if current character is uppercase
		if unicode.IsUpper(r) {
			// Add underscore if:
			// 1. Not the first character, and
			// 2. Either the previous character is lowercase, or
			// 3. Not at the end and the next character is lowercase (end of acronym)
			if i > 0 && (unicode.IsLower(rune(s[i-1])) ||
				(i < len(s)-1 && unicode.IsLower(rune(s[i+1])))) {
				result.WriteRune('_')
			}
		}

		result.WriteRune(unicode.ToLower(r))
	}

	return result.String()
}
