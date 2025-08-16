package mcpargs

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestToolOptions(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    []mcp.ToolOption
		wantErr string
	}{
		{
			name: "basic struct with string field",
			input: struct {
				Name string `mcp_desc:"The name" mcp_required:"true"`
			}{},
			want: []mcp.ToolOption{
				mcp.WithString("name", mcp.Description("The name"), mcp.Required()),
			},
		},
		{
			name: "struct with different field types",
			input: struct {
				Name   string  `mcp_desc:"The name"`
				Age    int     `mcp_desc:"The age" mcp_required:"true"`
				Active bool    `mcp_desc:"Is active"`
				Score  float64 `mcp_desc:"The score"`
			}{},
			want: []mcp.ToolOption{
				mcp.WithString("name", mcp.Description("The name")),
				mcp.WithNumber("age", mcp.Description("The age"), mcp.Required()),
				mcp.WithBoolean("active", mcp.Description("Is active")),
				mcp.WithNumber("score", mcp.Description("The score")),
			},
		},
		{
			name: "string field with enum values",
			input: struct {
				Status string `mcp_desc:"The status" mcp_enum:"pending,active,completed"`
			}{},
			want: []mcp.ToolOption{
				mcp.WithString("status", mcp.Description("The status"), mcp.Enum("pending", "active", "completed")),
			},
		},
		{
			name: "camel case conversion",
			input: struct {
				UserName  string `mcp_desc:"The username"`
				UserEmail string `mcp_desc:"The email"`
			}{},
			want: []mcp.ToolOption{
				mcp.WithString("user_name", mcp.Description("The username")),
				mcp.WithString("user_email", mcp.Description("The email")),
			},
		},
		{
			name: "acronyms in field names",
			input: struct {
				HTTPServer  string `mcp_desc:"HTTP server"`
				APIEndpoint string `mcp_desc:"API endpoint"`
			}{},
			want: []mcp.ToolOption{
				mcp.WithString("http_server", mcp.Description("HTTP server")),
				mcp.WithString("api_endpoint", mcp.Description("API endpoint")),
			},
		},
		{
			name: "unexported fields are skipped",
			input: struct {
				Name string `mcp_desc:"The name"`
				id   string
			}{},
			want: []mcp.ToolOption{
				mcp.WithString("name", mcp.Description("The name")),
			},
		},
		{
			name: "pointer to struct",
			input: &struct {
				Name string `mcp_desc:"The name"`
			}{},
			want: []mcp.ToolOption{
				mcp.WithString("name", mcp.Description("The name")),
			},
		},
		{
			name:    "non-struct input",
			input:   "not a struct",
			want:    nil,
			wantErr: "expected struct, got string",
		},
		{
			name: "missing mcp_desc tag",
			input: struct {
				Name string
			}{},
			want:    nil,
			wantErr: `missing "mcp_desc" tag on field "Name"`,
		},
		{
			name: "invalid mcp_required tag",
			input: struct {
				Name string `mcp_desc:"The name" mcp_required:"yes"`
			}{},
			want:    nil,
			wantErr: `invalid "mcp_required" tag on field "Name": must be "true" or "false"`,
		},
		{
			name: "mcp_enum on non-string field",
			input: struct {
				Count int `mcp_desc:"The count" mcp_enum:"1,2,3"`
			}{},
			want:    nil,
			wantErr: `invalid "mcp_enum" tag on field "Count": must be a string`,
		},
		{
			name: "unsupported field type",
			input: struct {
				Data []string `mcp_desc:"The data"`
			}{},
			want:    nil,
			wantErr: `unsupported field type "slice" for field "Data"`,
		},
		{
			name: "multiple errors",
			input: struct {
				Name  string
				Count int `mcp_desc:"The count" mcp_enum:"1,2,3"`
			}{},
			want:    nil,
			wantErr: `missing "mcp_desc" tag on field "Name"`,
		},
		{
			name: "field implementing Marshaler",
			input: struct {
				ID ID `mcp_desc:"The resource ID" mcp_required:"true"`
			}{},
			want: []mcp.ToolOption{
				mcp.WithString("id", mcp.Description("The resource ID"), mcp.Required()),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)

			// Check errors first
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("ToolOptions() error = nil, wantErr %q", tt.wantErr)
					return
				}

				if !errors.Is(err, err) && !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("ToolOptions() error = %v, wantErr %q", err, tt.wantErr)
				}

				return
			}

			if err != nil {
				t.Errorf("ToolOptions() unexpected error = %v", err)
				return
			}

			wantTool := mcp.NewTool("test_tool", tt.want...)
			gotTool := mcp.NewTool("test_tool", got...)

			if diff := cmp.Diff(wantTool, gotTool, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ToolOptions() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestToolOptionsMultipleErrors tests that multiple errors are correctly joined together.
func TestToolOptionsMultipleErrors(t *testing.T) {
	input := struct {
		Name  string
		Count int               `mcp_desc:"The count" mcp_enum:"1,2,3"`
		Data  map[string]string `mcp_desc:"The data"`
	}{}

	_, err := Marshal(input)

	if err == nil {
		t.Fatal("ToolOptions() error = nil, want multiple errors")
	}

	expectedErrs := []string{
		`missing "mcp_desc" tag on field "Name"`,
		`invalid "mcp_enum" tag on field "Count": must be a string`,
		`unsupported field type "map" for field "Data"`,
	}

	for _, expected := range expectedErrs {
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("ToolOptions() error does not contain %q, got: %v", expected, err)
		}
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Field", "field"},
		{"FieldName", "field_name"},
		{"HTTPServer", "http_server"},
		{"ExportDNS", "export_dns"},
		{"AssigneeIDs", "assignee_ids"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := toSnakeCase(test.input)
			if got != test.want {
				t.Errorf("toSnakeCase(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}
