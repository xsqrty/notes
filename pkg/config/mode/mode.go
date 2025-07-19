package mode

import (
	"fmt"
	"strings"
)

// Mode represents the operational mode of the application, typically used to distinguish between environments like dev or prod.
type Mode string

const (
	// Dev represents the development environment.
	Dev = "dev"
	// Prod represents the production environment.
	Prod = "prod"
)

// UnmarshalText parses the input byte slice and assigns the corresponding Mode value, returning an error if invalid.
func (m *Mode) UnmarshalText(mode []byte) error {
	val, err := ParseMode(string(mode))
	if err != nil {
		return err
	}

	*m = val
	return nil
}

// ParseMode parses a string and returns it as a Mode type if it matches predefined modes; otherwise, it returns an error.
func ParseMode(mode string) (Mode, error) {
	mode = strings.ToLower(mode)
	switch mode {
	case Dev, Prod:
		return Mode(mode), nil
	default:
		return "", fmt.Errorf("unknown mode: %s", mode)
	}
}
