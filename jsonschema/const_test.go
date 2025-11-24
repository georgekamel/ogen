package jsonschema

import (
	"fmt"
	"testing"

	"github.com/go-faster/yaml"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/location"
)

func TestConstParsing(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		expect    *Schema
		expectErr bool
	}{
		{
			name: "string const",
			raw:  `{"type": "string", "const": "active"}`,
			expect: &Schema{
				Type:     String,
				Const:    "active",
				ConstSet: true,
			},
			expectErr: false,
		},
		{
			name: "integer const",
			raw:  `{"type": "integer", "const": 42}`,
			expect: &Schema{
				Type:     Integer,
				Const:    int64(42),
				ConstSet: true,
			},
			expectErr: false,
		},
		{
			name: "number const",
			raw:  `{"type": "number", "const": 3.14}`,
			expect: &Schema{
				Type:     Number,
				Const:    3.14,
				ConstSet: true,
			},
			expectErr: false,
		},
		{
			name: "boolean const true",
			raw:  `{"type": "boolean", "const": true}`,
			expect: &Schema{
				Type:     Boolean,
				Const:    true,
				ConstSet: true,
			},
			expectErr: false,
		},
		{
			name: "boolean const false",
			raw:  `{"type": "boolean", "const": false}`,
			expect: &Schema{
				Type:     Boolean,
				Const:    false,
				ConstSet: true,
			},
			expectErr: false,
		},
		{
			name: "null const",
			raw:  `{"type": "null", "const": null}`,
			expect: &Schema{
				Type:     Null,
				Const:    nil,
				ConstSet: true,
			},
			expectErr: false,
		},
		{
			name: "const with object property",
			raw: `{
				"type": "object",
				"properties": {
					"status": {
						"type": "string",
						"const": "active"
					}
				}
			}`,
			expect: &Schema{
				Type: Object,
				Properties: []Property{
					{
						Name: "status",
						Schema: &Schema{
							Type:     String,
							Const:    "active",
							ConstSet: true,
						},
					},
				},
			},
			expectErr: false,
		},
		// Note: Invalid const types are validated during code generation,
		// not during parsing. The parser will parse the value as-is.
		// Validation happens in gen/schema_gen_primitive.go:validateConstValue
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)
			data := []byte(tt.raw)

			var raw RawSchema
			a.NoError(yaml.Unmarshal(data, &raw))

			const filename = "test.yaml"
			out, err := NewParser(Settings{
				File: location.NewFile(filename, filename, data),
			}).Parse(&raw, testCtx())
			if tt.expectErr {
				a.Error(err)
				return
			}
			a.NoError(err)
			// Zero locator to simplify comparison.
			out.Pointer = location.Pointer{}
			if tt.expect != nil {
				// Compare const values
				a.Equal(tt.expect.ConstSet, out.ConstSet, "ConstSet")
				a.Equal(tt.expect.Const, out.Const, "Const")
				if tt.expect.Type != "" {
					a.Equal(tt.expect.Type, out.Type, "Type")
				}
				if len(tt.expect.Properties) > 0 {
					a.Equal(len(tt.expect.Properties), len(out.Properties), "Properties length")
					for i, prop := range tt.expect.Properties {
						if i < len(out.Properties) {
							a.Equal(prop.Name, out.Properties[i].Name, fmt.Sprintf("Property[%d].Name", i))
							if prop.Schema != nil {
								a.Equal(prop.Schema.ConstSet, out.Properties[i].Schema.ConstSet, fmt.Sprintf("Property[%d].ConstSet", i))
								a.Equal(prop.Schema.Const, out.Properties[i].Schema.Const, fmt.Sprintf("Property[%d].Const", i))
							}
						}
					}
				}
			}
		})
	}
}

func TestConstWithEnum(t *testing.T) {
	// Const and enum should not be used together, but we test that const takes precedence
	raw := `{"type": "string", "const": "active", "enum": ["active", "inactive"]}`
	data := []byte(raw)

	var rawSchema RawSchema
	require.NoError(t, yaml.Unmarshal(data, &rawSchema))

	const filename = "test.yaml"
	out, err := NewParser(Settings{
		File: location.NewFile(filename, filename, data),
	}).Parse(&rawSchema, testCtx())
	require.NoError(t, err)

	// Const should be set
	require.True(t, out.ConstSet)
	require.Equal(t, "active", out.Const)
	// Enum should also be parsed
	require.Len(t, out.Enum, 2)
}

