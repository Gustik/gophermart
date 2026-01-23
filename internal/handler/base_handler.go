package handler

import (
	"errors"
	"net/http"

	"go.uber.org/zap"
)

type BaseHandler struct {
	logger         *zap.Logger
	errorStatusMap map[error]int
}

func NewBaseHandler(logger *zap.Logger) *BaseHandler {
	return &BaseHandler{
		logger: logger,
	}
}

// HandleError обрабатывает ошибку
func (h BaseHandler) HandleError(w http.ResponseWriter, err error) {
	for knownErr, status := range h.errorStatusMap {
		if errors.Is(err, knownErr) {
			if status == http.StatusOK {
				w.WriteHeader(status)
			} else {
				http.Error(w, err.Error(), status)
			}
			return
		}
	}

	// Неизвестная ошибка
	h.logger.Error("Внутренняя ошибка", zap.Error(err))
	http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
}
