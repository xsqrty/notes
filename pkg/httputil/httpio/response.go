package httpio

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/xsqrty/notes/pkg/httputil/httpio/errx"
	"github.com/xsqrty/notes/pkg/httputil/httpio/mapping"
)

// ResponseDecorator is an interface that extends http.ResponseWriter with methods to retrieve status code and bytes written.
type ResponseDecorator interface {
	http.ResponseWriter
	// StatusCode retrieves the HTTP status code of the response.
	StatusCode() int
	// BytesWritten returns the number of bytes written to the response body.
	BytesWritten() int
}

// resDec is a custom implementation of http.ResponseWriter, tracking status code and number of bytes written.
type resDec struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

// ErrorResponse represents the structure for error responses, containing detailed error information.
type ErrorResponse struct {
	Error *errx.CodeError `json:"error"`
}

// Json sends a JSON response with the specified status code and payload to the given http.ResponseWriter.
func Json(w http.ResponseWriter, statusCode int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(v) // nolint: gosec, errcheck
}

// Error writes an error message and status code as a JSON response to the provided http.ResponseWriter.
func Error(w http.ResponseWriter, statusCode int, err error) {
	codeOptionsError := errx.NewUnknown(err.Error())
	var maxBytesErr *http.MaxBytesError
	var ve validator.ValidationErrors

	switch {
	case errors.As(err, &ve):
		message, options := mapping.MapValidatorErrors(ve)
		codeOptionsError = errx.NewOptional(errx.CodeValidation, message, options)
	case errors.As(err, &codeOptionsError):
		codeOptionsError = errx.NewOptional(codeOptionsError.Code, err.Error(), codeOptionsError.Options)
	case errors.As(err, &maxBytesErr):
		statusCode = http.StatusRequestEntityTooLarge
		codeOptionsError = errx.NewOptional(
			errx.CodeBodyTooLarge,
			fmt.Sprintf("request limit %d bytes is exceeded", maxBytesErr.Limit),
			map[string]string{
				"max_bytes": strconv.Itoa(int(maxBytesErr.Limit)),
			},
		)
	case errors.Is(err, ErrJsonParse):
		codeOptionsError = errx.New(errx.CodeJsonParse, err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{codeOptionsError}) // nolint: gosec, errcheck
}

// NewResponseDecorator wraps an http.ResponseWriter with additional functionality to track status code and bytes written.
func NewResponseDecorator(w http.ResponseWriter) ResponseDecorator {
	return &resDec{
		ResponseWriter: w,
	}
}

// WriteHeader writes the HTTP status code to the underlying ResponseWriter and updates the internal statusCode field.
func (w *resDec) WriteHeader(statusCode int) {
	if statusCode >= 100 && statusCode <= 199 && statusCode != http.StatusSwitchingProtocols {
		w.ResponseWriter.WriteHeader(statusCode)
		return
	}

	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write writes data to the underlying ResponseWriter and tracks the number of bytes written.
func (w *resDec) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += n
	return n, err
}

// BytesWritten returns the total number of bytes written to the response by the ResponseWriter.
func (w *resDec) BytesWritten() int {
	return w.bytesWritten
}

// StatusCode returns the HTTP status code that has been set for the response.
func (w *resDec) StatusCode() int {
	return w.statusCode
}
