// Package mcpargs provides utilities for working with argument marshaling and unmarshaling
// in MCP (Multipurpose Control Protocol) tools. It simplifies the process of converting
// between Go structs and MCP tool options/arguments.
//
// The package offers two main capabilities:
//
// 1. Marshaling Go structs to MCP tool options for tool definition:
//   - Use the Marshal function to convert a struct into MCP tool options
//   - Use NewTool as a convenient wrapper around mcp.NewTool for automatic marshaling
//   - Implement the Marshaler interface for custom type handling
//   - Fields are automatically converted from CamelCase to snake_case
//   - Supported struct tags:
//   - mcp_desc: Required. Provides field description
//   - mcp_required: Optional. Set to "true" if the field is required
//   - mcp_enum: Optional. For string fields, comma-separated list of allowed values
//
// 2. Unmarshaling MCP arguments back to Go structs:
//   - Use the Unmarshal function to populate a struct from an arguments map
//   - Implement the Unmarshaler interface for custom unmarshaling logic
//   - Automatic type conversions between common Go types
//
// Example usage:
//
//	// Define a struct with appropriate tags
//	type UserArgs struct {
//	  UserID    string `mcp_desc:"The ID or username of the user" mcp_required:"false"`
//	  MaxItems  int    `mcp_desc:"Maximum number of items to return" mcp_required:"false"`
//	  SortOrder string `mcp_desc:"Sort order for results" mcp_enum:"asc,desc"`
//	}
//
//	// For defining a tool:
//	func NewUserTool() mcp.Tool {
//	  return mcpargs.NewTool("get_user", UserArgs{},
//	    mcp.WithDescription("Get information about a user"))
//	}
//
//	// Or using Marshal directly:
//	opts, err := mcpargs.Marshal(UserArgs{})
//	tool := mcp.NewTool("get_user",
//	  append([]mcp.ToolOption{mcp.WithDescription("...")}, opts...)...)
//
//	// For handling tool invocation:
//	func handleUser(args map[string]any) (any, error) {
//	  var userArgs UserArgs
//	  if err := mcpargs.Unmarshal(args, &userArgs); err != nil {
//	    return nil, err
//	  }
//	  // Use userArgs...
//	}
//
//	// For custom type handling, implement Marshaler and/or Unmarshaler:
//	type CustomType string
//
//	func (c CustomType) Marshal() mcpargs.MCPType {
//	  return mcpargs.TypeString
//	}
//
//	func (c *CustomType) Unmarshal(v any) error {
//	  str, ok := v.(string)
//	  if !ok {
//	    return fmt.Errorf("expected string, got %T", v)
//	  }
//	  *c = CustomType(str)
//	  return nil
//	}
package mcpargs
