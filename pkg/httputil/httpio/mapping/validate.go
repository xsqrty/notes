package mapping

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// mapping is a predefined map that associates validation tags with corresponding error message templates.
var mapping = map[string]string{
	"required":   `Field "{key}" is required`,
	"email":      `Field "{key}" is incorrect email`,
	"min:string": `Field "{key}" must be at least {param} characters long`,
	"min":        `Field "{key}" must be at least {param}`,
	"max:string": `Field "{key}" must be no longer than {param} characters.`,
	"max":        `Field "{key}" must be no more than {param}`,
	"uuid":       `Field "{key}" is not a valid UUID`,
}

// MapValidatorErrors converts validation errors into a readable message and a map of field-specific error messages.
func MapValidatorErrors(ve validator.ValidationErrors) (string, map[string]string) {
	resultMessage := ""
	resultOptions := make(map[string]string)

	for _, fe := range ve {
		message := getMessage(fe)
		resultOptions[fe.Field()] = message
		resultMessage = message
	}

	if resultMessage == "" {
		resultMessage = ve.Error()
	}

	return resultMessage, resultOptions
}

// getMessage generates a user-friendly validation error message based on the provided FieldError.
func getMessage(fe validator.FieldError) string {
	message := ""
	kind := fe.Kind()

	if kind == reflect.String {
		message = mapping[fe.Tag()+":int"]
	} else if kind >= reflect.Int && kind <= reflect.Int64 {
		message = mapping[fe.Tag()+":int"]
	}

	if message == "" {
		message = mapping[fe.Tag()]
	}

	if message != "" {
		return strings.Replace(strings.Replace(message, "{key}", fe.Field(), 1), "{param}", fe.Param(), 1)
	}

	return fe.Error()
}
