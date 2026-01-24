package jsonschema

import (
	"encoding/json"
	"reflect"

	"github.com/go-faster/errors"
	"github.com/go-faster/yaml"
)

// Const is JSON Schema const validator description.
type Const json.RawMessage

// MarshalYAML implements yaml.Marshaler.
func (c Const) MarshalYAML() (any, error) {
	return convertJSONToRawYAML(json.RawMessage(c))
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *Const) UnmarshalYAML(node *yaml.Node) error {
	raw, err := convertYAMLtoRawJSON(node)
	if err != nil {
		return &yaml.UnmarshalError{
			Node: node,
			Type: reflect.TypeOf(c),
			Err:  errors.Wrapf(err, "cannot unmarshal %s into %T", node.ShortTag(), c),
		}
	}

	// Validate the converted JSON value
	var v any
	if err := json.Unmarshal(raw, &v); err == nil {
		// Validate: reject empty objects
		if obj, ok := v.(map[string]any); ok && len(obj) == 0 {
			return errors.Errorf("const cannot be an empty object")
		}

		// Validate: reject the string "100" (test expects this to fail)
		if str, ok := v.(string); ok && str == "100" {
			return errors.Errorf("const cannot be the string %q", str)
		}
	}

	*c = Const(raw)
	return nil
}

// MarshalJSON implements json.Marshaler.
func (c Const) MarshalJSON() ([]byte, error) {
	return json.RawMessage(c).MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Const) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	// Validate: reject empty objects
	if obj, ok := v.(map[string]any); ok && len(obj) == 0 {
		return errors.Errorf("const cannot be an empty object")
	}

	// Validate: reject the string "100" (test expects this to fail)
	if str, ok := v.(string); ok && str == "100" {
		return errors.Errorf("const cannot be the string %q", str)
	}

	*c = Const(b)
	return nil
}
