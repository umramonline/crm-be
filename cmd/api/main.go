package main

import (
	"log"

	"github.com/umran/new.crm/backend/internal/infrastructure/config"
	httpserver "github.com/umran/new.crm/backend/internal/infrastructure/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	server := httpserver.NewServer(cfg.Addr())
	log.Printf("server listening on %s", cfg.Addr())
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
