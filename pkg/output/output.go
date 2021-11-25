package output

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

// Format is a type for capturing supported output formats.
type Format string

const (
	JSON  Format = "json"
	YAML  Format = "yaml"
	TABLE Format = "table"
)

// ErrInvalidFormatType is returned when an unsupported format type is used.
var ErrInvalidFormatType = errors.New("invalid format type")

// Writer is an interface that any type can implement to write supported formats.
type Writer interface {
	// WriteTable will write tabular output into the given io.Writer, returning
	// an error if any occur.
	WriteTable(out io.Writer) error
	// WriteJSON will write JSON formatted output into the given io.Writer,
	// returning an error if any occur.
	WriteJSON(out io.Writer) error
	// WriteYAML will write YAML formatted output into the given io.Writer,
	// returning an error if any occur.
	WriteYAML(out io.Writer) error
}

// String returns the string representation of the Format.
func (o Format) String() string {
	return string(o)
}

// Write the output in the given format to the io.Writer. Unsupported formats
// will return an error.
func (o Format) Write(w io.Writer, obj interface{}) error {
	switch o {
	case JSON:
		return EncodeJSON(w, obj)
	case YAML:
		return EncodeYAML(w, obj)
	case TABLE:
		if obj, ok := obj.(*uitable.Table); ok {
			return EncodeTable(w, obj)
		}
	}
	return ErrInvalidFormatType
}

// ParseFormat takes a raw string and returns the matching Format.
// If the format does not exists, ErrInvalidFormatType is returned.
func ParseFormat(s string) (out Format, err error) {
	switch strings.ToLower(s) {
	case JSON.String():
		out, err = JSON, nil
	case YAML.String():
		out, err = YAML, nil
	case TABLE.String():
		out, err = TABLE, nil
	default:
		out, err = "", ErrInvalidFormatType
	}
	return
}

// EncodeJSON is a helper function to decorate any error message with a bit more
// context and avoid writing the same code over and over for printers.
func EncodeJSON(out io.Writer, obj interface{}) error {
	enc := json.NewEncoder(out)
	if err := enc.Encode(obj); err != nil {
		return errors.Wrap(err, "unable to write JSON output")
	}
	return nil
}

// EncodeYAML is a helper function to decorate any error message with a bit more
// context and avoid writing the same code over and over for printers.
func EncodeYAML(out io.Writer, obj interface{}) error {
	raw, err := yaml.Marshal(obj)
	if err != nil {
		return errors.Wrap(err, "unable to write YAML output")
	}

	if _, err = out.Write(raw); err != nil {
		return errors.Wrap(err, "unable to write YAML output")
	}
	return nil
}

// EncodeTable is a helper function to decorate any error message with a bit
// more context and avoid writing the same code over and over for printers.
func EncodeTable(out io.Writer, table *uitable.Table) error {
	raw := table.Bytes()
	raw = append(raw, []byte("\n")...)
	if _, err := out.Write(raw); err != nil {
		return errors.Wrap(err, "unable to write table output")
	}
	return nil
}
