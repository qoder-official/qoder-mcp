package mcpargs

// OptionalBool is an optional boolean type. It supports three states: true, false, and "not set".
// The zero value is "not set".
type OptionalBool struct { //nolint:recvcheck // Unmarshal requires pointer receiver.
	value bool
	isSet bool
}

// Interface tests.
var _ Marshaler = OptionalBool{}
var _ Unmarshaler = &OptionalBool{}

// Unmarshal sets the OptionalBool value from a string or bool.
func (o *OptionalBool) Unmarshal(v any) error {
	if b, ok := v.(bool); ok {
		o.value = b
		o.isSet = true

		return nil
	}

	o.value = false
	o.isSet = false

	return nil
}

func (OptionalBool) Marshal() MCPType {
	return TypeBoolean
}

func (o OptionalBool) Ptr() *bool {
	if !o.isSet {
		return nil
	}

	// Create a copy to avoid mutation of the internal value.
	v := o.value

	return &v
}
