package mcpargs

import (
	"fmt"
	"strconv"
)

// ID is a type that can represent an ID either as a string or an integer.
// It implements the Unmarshaler interface to unmarshal from either string or int types.
type ID struct { //nolint:recvcheck // Unmarshal requires pointer receiver.
	String  string
	Integer int
}

// Interface tests.
var _ Marshaler = ID{}
var _ Unmarshaler = &ID{}

// Unmarshal sets the ID value from a string or integer.
// It implements the Unmarshaler interface.
func (id *ID) Unmarshal(v any) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w: cannot unmarshal ID from %T", ErrUnmarshalArguments, v)
	}

	if i, err := strconv.Atoi(s); err == nil {
		id.Integer = i
		id.String = ""

		return nil
	}

	id.String = s
	id.Integer = 0

	return nil
}

// Marshal implements the Marshaler interface.
func (ID) Marshal() MCPType {
	return TypeString
}

// Value returns either the integer or the string value of the ID.
func (id ID) Value() any {
	if id.Integer != 0 {
		return id.Integer
	}

	return id.String
}

// IsZero returns true if the ID is zero (both String and Integer are zero).
func (id ID) IsZero() bool {
	return id.Integer == 0 && id.String == ""
}
