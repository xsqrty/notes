package httpio

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"io"
	"reflect"
	"strings"
)

var validate = validator.New()
var (
	ErrJsonParse = errors.New("json parse error")
)

func init() {
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		json := fld.Tag.Get("json")
		if json == "-" {
			return ""
		}

		return strings.Split(json, ",")[0]
	})
}

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
