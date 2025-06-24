package httpx

import (
	"encoding/json"
	"net/http"
)

func WriteValidationErrors(w http.ResponseWriter, errors map[string]string) {
	response := map[string]any{
		"errors": errors,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}
