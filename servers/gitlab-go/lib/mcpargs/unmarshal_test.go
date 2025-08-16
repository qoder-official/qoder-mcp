package mcpargs

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name            string
		arguments       map[string]any
		want            any
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "basic struct with string field",
			arguments: map[string]any{
				"name": "John Doe",
			},
			want: struct {
				Name string `mcp_desc:"The name" mcp_required:"true"`
			}{
				Name: "John Doe",
			},
		},
		{
			name: "multiple field types",
			arguments: map[string]any{
				"name":    "Jane Smith",
				"age":     30,
				"active":  true,
				"score":   98.5,
				"ignored": "this field isn't in the struct",
			},
			want: struct {
				Name   string  `mcp_desc:"The name"`
				Age    int     `mcp_desc:"The age"`
				Active bool    `mcp_desc:"Active status"`
				Score  float64 `mcp_desc:"Test score"`
			}{
				Name:   "Jane Smith",
				Age:    30,
				Active: true,
				Score:  98.5,
			},
		},
		{
			name: "snake case field mapping",
			arguments: map[string]any{
				"user_name":  "jdoe",
				"user_email": "jdoe@example.com",
			},
			want: struct {
				UserName  string `mcp_desc:"Username"`
				UserEmail string `mcp_desc:"Email address"`
			}{
				UserName:  "jdoe",
				UserEmail: "jdoe@example.com",
			},
		},
		{
			name: "acronyms in field names",
			arguments: map[string]any{
				"http_server": "localhost:8080",
				"api_key":     "secret-key-123",
			},
			want: struct {
				HTTPServer string `mcp_desc:"HTTP Server address"`
				APIKey     string `mcp_desc:"API authentication key"`
			}{
				HTTPServer: "localhost:8080",
				APIKey:     "secret-key-123",
			},
		},
		{
			name: "type conversions",
			arguments: map[string]any{
				"string_val": "42",         // We don't convert string to other types
				"int_val":    42.5,         // Float to int
				"uint_val":   float64(100), // Float to uint
				"float_val":  42,           // Int to float
			},
			want: struct {
				StringVal string  `mcp_desc:"String value"`
				IntVal    int     `mcp_desc:"Integer value"`
				UintVal   uint    `mcp_desc:"Unsigned integer value"`
				FloatVal  float64 `mcp_desc:"Float value"`
			}{
				StringVal: "42",
				IntVal:    42,
				UintVal:   100,
				FloatVal:  42.0,
			},
		},
		{
			name: "partially populated struct",
			arguments: map[string]any{
				"first_name": "John",
				// last_name is missing
			},
			want: struct {
				FirstName string `mcp_desc:"First name"`
				LastName  string `mcp_desc:"Last name"`
			}{
				FirstName: "John",
				LastName:  "", // Zero value because it's not in arguments
			},
		},
		{
			name: "unexported fields result in error",
			arguments: map[string]any{
				"name":    "John",
				"private": "should cause error",
			},
			want: struct {
				Name    string `mcp_desc:"Name"`
				private string // unexported field
			}{},
			wantErr:         true,
			wantErrContains: "unexported field",
		},
		{
			name: "nested struct - should error",
			arguments: map[string]any{
				"name": "John",
				"address": map[string]any{
					"city":  "New York",
					"state": "NY",
				},
			},
			want: struct {
				Name    string `mcp_desc:"Name"`
				Address struct {
					City  string `mcp_desc:"City"`
					State string `mcp_desc:"State"`
				} `mcp_desc:"Address"`
			}{},
			wantErr:         true,
			wantErrContains: "unsupported type",
		},
		{
			name: "struct with channel field - should error",
			arguments: map[string]any{
				"name":     "John",
				"messages": nil,
			},
			want: struct {
				Name     string      `mcp_desc:"Name"`
				Messages chan string `mcp_desc:"Message channel"`
			}{},
			wantErr:         true,
			wantErrContains: "unsupported type",
		},
		{
			name: "struct with func field - should error",
			arguments: map[string]any{
				"name":     "John",
				"callback": nil,
			},
			want: struct {
				Name     string `mcp_desc:"Name"`
				Callback func() `mcp_desc:"Callback function"`
			}{},
			wantErr:         true,
			wantErrContains: "unsupported type",
		},
		{
			name: "struct with map field - should error",
			arguments: map[string]any{
				"name":     "John",
				"metadata": map[string]string{"key": "value"},
			},
			want: struct {
				Name     string            `mcp_desc:"Name"`
				Metadata map[string]string `mcp_desc:"Metadata"`
			}{},
			wantErr:         true,
			wantErrContains: "unsupported type",
		},
		{
			name: "struct with slice field - should error",
			arguments: map[string]any{
				"name": "John",
				"tags": []string{"tag1", "tag2"},
			},
			want: struct {
				Name string   `mcp_desc:"Name"`
				Tags []string `mcp_desc:"Tags"`
			}{},
			wantErr:         true,
			wantErrContains: "unsupported type",
		},
		{
			name: "error - non-struct pointer",
			arguments: map[string]any{
				"value": 42,
			},
			want:            42, // Not a struct
			wantErr:         true,
			wantErrContains: "must be a pointer to a struct",
		},
		{
			name: "group ID using string",
			arguments: map[string]any{
				"id": "owner/namespace",
			},
			want: struct {
				ID ID `mcp_desc:"The group ID" mcp_required:"true"`
			}{
				ID: ID{
					String: "owner/namespace",
				},
			},
		},
		{
			name: "group ID using integer",
			arguments: map[string]any{
				"id": "1234",
			},
			want: struct {
				ID ID `mcp_desc:"The group ID" mcp_required:"true"`
			}{
				ID: ID{
					Integer: 1234,
				},
			},
		},
		{
			name: "required field is missing - should error",
			arguments: map[string]any{
				"project_id": "1234",
				// missing: merge_request_iid
			},
			want: struct {
				ProjectID       ID  `mcp_desc:"The project ID" mcp_required:"true"`
				MergeRequestIID int `mcp_desc:"The merge request ID" mcp_required:"true"`
			}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new zero value of the concrete type
			wantType := reflect.TypeOf(tt.want)
			got := reflect.New(wantType).Elem()

			// Pass a pointer to got to the Unmarshal function
			err := Unmarshal(tt.arguments, got.Addr().Interface())

			// Check if we expect an error
			if tt.wantErr {
				if err == nil {
					t.Errorf("Unmarshal() expected error but got nil")
					return
				}

				if tt.wantErrContains != "" && !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("Unmarshal() error = %v, wantErrContains %q", err, tt.wantErrContains)
				}

				return
			}

			// Otherwise we don't expect an error
			if err != nil {
				t.Errorf("Unmarshal() error = %v", err)
				return
			}

			if diff := cmp.Diff(tt.want, got.Interface(), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Unmarshal() mismatch (-want/+got):\n%s", diff)
			}
		})
	}
}

// Additional test for errors.
func TestUnmarshalErrors(t *testing.T) {
	tests := []struct {
		name            string
		arguments       map[string]any
		target          any
		wantErrContains string
	}{
		{
			name: "nil pointer",
			arguments: map[string]any{
				"name": "John",
			},
			target:          nil,
			wantErrContains: "must be a non-nil pointer",
		},
		{
			name: "non-pointer target",
			arguments: map[string]any{
				"name": "John",
			},
			target:          struct{}{},
			wantErrContains: "must be a non-nil pointer",
		},
		{
			name: "pointer to non-struct",
			arguments: map[string]any{
				"value": 42,
			},
			target:          new(int),
			wantErrContains: "must be a pointer to a struct",
		},
		{
			name: "type conversion error - string to int",
			arguments: map[string]any{
				"age": "not-a-number",
			},
			target: &struct {
				Age int `mcp_desc:"Age"`
			}{},
			wantErrContains: "cannot convert",
		},
		{
			name: "negative value to unsigned int",
			arguments: map[string]any{
				"count": -10,
			},
			target: &struct {
				Count uint `mcp_desc:"Count"`
			}{},
			wantErrContains: "negative value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal(tt.arguments, tt.target)

			// We expect an error in all these cases
			if err == nil {
				t.Errorf("Unmarshal() expected error but got nil")
				return
			}

			// Check if error message contains expected string
			if !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Errorf("Unmarshal() error = %v, wantErrContains %q", err, tt.wantErrContains)
			}
		})
	}
}
