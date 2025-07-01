package mode

import (
	"fmt"
	"strings"
)

type Mode string

const (
	Dev  = "dev"
	Prod = "prod"
)

func (m *Mode) UnmarshalText(mode []byte) error {
	val, err := ParseMode(string(mode))
	if err != nil {
		return err
	}

	*m = val
	return nil
}

func ParseMode(mode string) (Mode, error) {
	mode = strings.ToLower(mode)
	switch mode {
	case Dev, Prod:
		return Mode(mode), nil
	default:
		return "", fmt.Errorf("unknown mode: %s", mode)
	}
}
