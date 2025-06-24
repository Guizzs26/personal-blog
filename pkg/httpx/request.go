package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Guizzs26/personal-blog/pkg/validatorx"
)

func Bind[T any](r *http.Request) (*T, error) {
	var data T

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse the request body: %w", err)
	}

	if err := validatorx.ValidateStruct(&data); err != nil {
		return nil, err
	}

	return &data, nil
}
