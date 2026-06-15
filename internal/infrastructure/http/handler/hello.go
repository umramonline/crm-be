package handler

import (
	"net/http"

	app "github.com/umran/new.crm/backend/internal/application/greeting"
)

type HelloHandler struct {
	service *app.Service
}

func NewHelloHandler(service *app.Service) *HelloHandler {
	return &HelloHandler{service: service}
}

func (h *HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	greeting := h.service.GetHello()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(greeting.Message))
}
