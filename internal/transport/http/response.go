package http

import (
	"encoding/json"
	"net/http"
	"pr-review-service/internal/domain"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

func handleDomainError(w http.ResponseWriter, err error) {
	if domainErr, ok := err.(*domain.DomainError); ok {
		status := http.StatusBadRequest

		switch domainErr.Code {
		case domain.ErrCodeTeamExists, domain.ErrCodePRExists:
			status = http.StatusConflict
		case domain.ErrCodePRMerged, domain.ErrCodeNotAssigned, domain.ErrCodeNoCandidate:
			status = http.StatusConflict
		case domain.ErrCodeNotFound:
			status = http.StatusNotFound
		}

		respondError(w, status, domainErr.Code, domainErr.Message)
		return
	}

	respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
}
