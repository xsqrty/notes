package httpio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// validate is an instance of a new validator used for input validation.
// ErrJsonParse is an error returned when JSON parsing fails.
var (
	validate     = validator.New()
	ErrJsonParse = errors.New("json parse error")
)

// init initializes the custom tag name function for the validator to use JSON field names instead of struct field names.
func init() {
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		json := fld.Tag.Get("json")
		if json == "-" {
			return ""
		}

		return strings.Split(json, ",")[0]
	})
}

// Parse decodes the JSON payload from the provided body into a struct of type T and validates the resulting struct.
// Returns an error if decoding or validation fails.
func Parse[T any](body io.ReadCloser) (T, error) {
	var result T
	err := json.NewDecoder(body).Decode(&result)
	if err != nil {
		return result, errors.Join(ErrJsonParse, err)
	}

	err = validate.Struct(&result)
	if err != nil {
		return result, fmt.Errorf("validate: %w", err)
	}

	return result, nil
}
