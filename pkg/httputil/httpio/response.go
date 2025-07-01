package httpio

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/xsqrty/notes/pkg/httputil/httpio/errx"
	"github.com/xsqrty/notes/pkg/httputil/httpio/mapping"
	"net/http"
	"strconv"
)

type ResponseDecorator interface {
	http.ResponseWriter
	StatusCode() int
	BytesWritten() int
}

type resDec struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
	wroteHeader  bool
}

type ErrorResponse struct {
	Error *errx.CodeError `json:"error"`
}

func Json(w http.ResponseWriter, statusCode int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(v)
}

func Error(w http.ResponseWriter, statusCode int, err error) {
	var codeOptionsError = errx.NewUnknown(err.Error())
	var maxBytesErr *http.MaxBytesError
	var ve validator.ValidationErrors

	if errors.As(err, &ve) {
		message, options := mapping.MapValidatorErrors(ve)
		codeOptionsError = errx.NewOptional(errx.CodeValidation, message, options)
	} else if errors.As(err, &codeOptionsError) {
		codeOptionsError = errx.NewOptional(codeOptionsError.Code, err.Error(), codeOptionsError.Options)
	} else if errors.As(err, &maxBytesErr) {
		statusCode = http.StatusRequestEntityTooLarge
		codeOptionsError = errx.NewOptional(errx.CodeBodyTooLarge, fmt.Sprintf("request limit %d bytes is exceeded", maxBytesErr.Limit), map[string]string{
			"max_bytes": strconv.Itoa(int(maxBytesErr.Limit)),
		})
	} else if errors.Is(err, ErrJsonParse) {
		codeOptionsError = errx.New(errx.CodeJsonParse, err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{codeOptionsError})
}

func NewResponseDecorator(w http.ResponseWriter) ResponseDecorator {
	return &resDec{
		ResponseWriter: w,
	}
}

func (w *resDec) WriteHeader(statusCode int) {
	if statusCode >= 100 && statusCode <= 199 && statusCode != http.StatusSwitchingProtocols {
		w.ResponseWriter.WriteHeader(statusCode)
		return
	}

	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *resDec) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += n
	return n, err
}

func (w *resDec) BytesWritten() int {
	return w.bytesWritten
}

func (w *resDec) StatusCode() int {
	return w.statusCode
}
