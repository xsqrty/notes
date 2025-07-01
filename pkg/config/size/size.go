package size

import "github.com/dustin/go-humanize"

type Bytes uint64

func (m *Bytes) UnmarshalText(value []byte) error {
	bytes, err := humanize.ParseBytes(string(value))
	if err != nil {
		return err
	}

	*m = Bytes(bytes)
	return nil
}
