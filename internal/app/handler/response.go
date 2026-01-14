package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Thorlik/marketplace/internal/app/dto"
	"go.uber.org/zap"
)

func RespondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}, logger *zap.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

func RespondWithError(w http.ResponseWriter, statusCode int, message string, logger *zap.Logger) {
	RespondWithJSON(w, statusCode, dto.ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}, logger)
}
