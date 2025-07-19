package formatter

import (
	"fmt"
	"strings"
)

// Formatter is a type representing the format mode for output, such as "json", "pretty", or "none".
type Formatter string

const (
	// Json represents a JSON output format.
	Json = "json"
	// Pretty represents a pretty-printed output format.
	Pretty = "pretty"
	// None represents no specific output format.
	None = ""
)

// UnmarshalText converts a byte slice to a Formatter value, validating and parsing the input. It returns an error if invalid.
func (m *Formatter) UnmarshalText(mode []byte) error {
	val, err := ParseFormatter(string(mode))
	if err != nil {
		return err
	}

	*m = val
	return nil
}

// ParseFormatter validates and parses the given formatter string into a Formatter type, returning an error if invalid.
func ParseFormatter(formatter string) (Formatter, error) {
	formatter = strings.ToLower(formatter)
	switch formatter {
	case Json, Pretty, None:
		return Formatter(formatter), nil
	default:
		return "", fmt.Errorf("unknown formatter: %s", formatter)
	}
}
