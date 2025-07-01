package formatter

import (
	"fmt"
	"strings"
)

type Formatter string

const (
	Json   = "json"
	Pretty = "pretty"
	None   = ""
)

func (m *Formatter) UnmarshalText(mode []byte) error {
	val, err := ParseFormatter(string(mode))
	if err != nil {
		return err
	}

	*m = val
	return nil
}

func ParseFormatter(formatter string) (Formatter, error) {
	formatter = strings.ToLower(formatter)
	switch formatter {
	case Json, Pretty, None:
		return Formatter(formatter), nil
	default:
		return "", fmt.Errorf("unknown formatter: %s", formatter)
	}
}
