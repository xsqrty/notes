package repoutil

import (
	"database/sql"
	"errors"
)

func RedefineNoRowsError(err error, newDefinition error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return newDefinition
	}

	return err
}
