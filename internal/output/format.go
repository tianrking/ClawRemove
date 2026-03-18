package output

import (
	"encoding/json"
	"io"

	"gopkg.in/yaml.v3"
)

// OutputFormat represents the output format type.
type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
	FormatYAML OutputFormat = "yaml"
	FormatHTML OutputFormat = "html"
)

// ParseFormat parses a string into an OutputFormat.
func ParseFormat(s string) OutputFormat {
	switch s {
	case "json":
		return FormatJSON
	case "yaml", "yml":
		return FormatYAML
	case "html":
		return FormatHTML
	default:
		return FormatText
	}
}

// Encoder provides methods to encode data in different formats.
type Encoder struct {
	format OutputFormat
	w      io.Writer
}

// NewEncoder creates a new Encoder for the given format.
func NewEncoder(w io.Writer, format OutputFormat) *Encoder {
	return &Encoder{format: format, w: w}
}

// Encode encodes the given value in the configured format.
func (e *Encoder) Encode(v any) error {
	switch e.format {
	case FormatJSON:
		enc := json.NewEncoder(e.w)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	case FormatYAML:
		enc := yaml.NewEncoder(e.w)
		enc.SetIndent(2)
		return enc.Encode(v)
	default:
		// Text format should be handled by specific print functions
		return nil
	}
}

// IsStructured returns true if the format is structured (JSON or YAML).
func (e *Encoder) IsStructured() bool {
	return e.format == FormatJSON || e.format == FormatYAML
}

// Format returns the current format.
func (e *Encoder) Format() OutputFormat {
	return e.format
}

// EncodeJSON encodes v as JSON to w.
func EncodeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// EncodeYAML encodes v as YAML to w.
func EncodeYAML(w io.Writer, v any) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	return enc.Encode(v)
}
