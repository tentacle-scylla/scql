package schema

import (
	"encoding/json"
	"io"
	"os"
)

// ToJSON serializes the schema to JSON bytes.
func (s *Schema) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// ToJSONIndent serializes the schema to indented JSON bytes.
func (s *Schema) ToJSONIndent() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// ParseJSON parses a schema from JSON bytes.
func ParseJSON(data []byte) (*Schema, error) {
	var s Schema
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// LoadFromJSON loads a schema from a JSON file.
func LoadFromJSON(path string) (*Schema, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return ParseJSON(data)
}

// SaveToJSON saves the schema to a JSON file with indentation.
func (s *Schema) SaveToJSON(path string) error {
	data, err := s.ToJSONIndent()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
