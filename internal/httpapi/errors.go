package httpapi

import (
	"errors"
	"net/http"

	"github.com/sheridiany/loomi/internal/diagnostics"
	"github.com/sheridiany/loomi/internal/productdata"
)

type apiErrorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func writeAPIError(w http.ResponseWriter, err error) {
	requestID := diagnostics.NewRequestID()
	code := productdata.ErrorCode(err)
	message := "Internal server error."
	var productErr productdata.ProductError
	if errors.As(err, &productErr) {
		message = productErr.Message
	}
	writeJSON(w, statusForError(err), apiErrorResponse{Error: apiError{Code: string(code), Message: message, RequestID: requestID}})
}

func statusForError(err error) int {
	switch productdata.ErrorCode(err) {
	case productdata.CodeInvalidRequest:
		return http.StatusBadRequest
	case productdata.CodeThreadNotFound:
		return http.StatusNotFound
	case productdata.CodeMethodNotAllowed:
		return http.StatusMethodNotAllowed
	default:
		return http.StatusInternalServerError
	}
}
