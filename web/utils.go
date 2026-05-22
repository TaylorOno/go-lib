package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("failed to request body: %w", err)
	}

	_ = r.Body.Close()
	return v, nil
}
