package http

import (
	"net/http"

	app "github.com/umran/new.crm/backend/internal/application/greeting"
	"github.com/umran/new.crm/backend/internal/infrastructure/http/handler"
)

type Server struct {
	addr string
	mux  *http.ServeMux
}

func NewServer(addr string) *Server {
	greetingService := app.NewService()
	helloHandler := handler.NewHelloHandler(greetingService)

	mux := http.NewServeMux()
	mux.Handle("/", helloHandler)

	return &Server{
		addr: addr,
		mux:  mux,
	}
}

func (s *Server) Run() error {
	return http.ListenAndServe(s.addr, s.mux)
}
