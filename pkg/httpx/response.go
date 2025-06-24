package httpx

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, code ErrorCode, message string, details ...APIErrorDetail) {
	err := APIError{}
	err.Error.Code = code
	err.Error.Message = message
	err.Error.Status = status

	if len(details) > 0 {
		err.Error.Details = details
	}

	WriteJSON(w, status, err)
}

func WriteValidationErrors(w http.ResponseWriter, errors map[string]string) {
	err := APIError{}
	err.Error.Code = ErrorCodeUnprocessable
	err.Error.Message = "Validation Failed"
	err.Error.Status = http.StatusUnprocessableEntity

	details := make([]APIErrorDetail, 0, len(errors))
	for f, d := range errors {
		details = append(details, APIErrorDetail{
			Field:       f,
			Description: d,
		})
	}
	err.Error.Details = details

	WriteError(w, http.StatusUnprocessableEntity, ErrorCodeUnprocessable, "Validation Failed", details...)
}
