package errx

// CodeError represents a custom error type with a code, message, and optional metadata. Used for standardized error handling.
type CodeError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Options map[string]string `json:"options,omitempty"`
}

// New creates a new instance of CodeError with the specified code and message, and no additional options.
func New(code, message string) *CodeError {
	return NewOptional(code, message, nil)
}

// NewOptional creates a new CodeError with the given code, message, and optional key-value pairs.
func NewOptional(code, message string, options map[string]string) *CodeError {
	return &CodeError{
		Code:    code,
		Message: message,
		Options: options,
	}
}

// NewUnknown creates a new CodeError instance with CodeUnknown and the provided message.
func NewUnknown(message string) *CodeError {
	return New(CodeUnknown, message)
}

// Error returns the error message stored in the CodeError instance.
func (e *CodeError) Error() string {
	return e.Message
}
