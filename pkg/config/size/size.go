package size

import "github.com/dustin/go-humanize"

// Bytes represent a type for handling data sizes in bytes, stored as an int64 value.
type Bytes int64

// UnmarshalText parses the given byte slice as a human-readable size and assigns it to the Bytes type.
func (m *Bytes) UnmarshalText(value []byte) error {
	bytes, err := humanize.ParseBytes(string(value))
	if err != nil {
		return err
	}

	*m = Bytes(bytes) // nolint: gosec
	return nil
}
