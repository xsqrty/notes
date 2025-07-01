package errx

type CodeError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Options map[string]string `json:"options,omitempty"`
}

func (e *CodeError) Error() string {
	return e.Message
}

func New(code, message string) *CodeError {
	return NewOptional(code, message, nil)
}

func NewOptional(code, message string, options map[string]string) *CodeError {
	return &CodeError{
		Code:    code,
		Message: message,
		Options: options,
	}
}

func NewUnknown(message string) *CodeError {
	return New(CodeUnknown, message)
}
