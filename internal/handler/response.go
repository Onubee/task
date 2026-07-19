package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *HTTPError  `json:"error,omitempty"`
}

func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("⚠️  Failed to encode JSON: %v", err)
	}
}

func RespondSuccess(w http.ResponseWriter, data interface{}) {
	RespondJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

func respondWithError(w http.ResponseWriter, status int, err error) {
	RespondJSON(w, status, Response{
		Success: false,
		Error: &HTTPError{
			Status:  status,
			Code:    http.StatusText(status),
			Message: err.Error(),
		},
	})
}
