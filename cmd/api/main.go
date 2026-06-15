package main

import (
	"log"

	httpserver "github.com/umran/new.crm/backend/internal/infrastructure/http"
)

func main() {
	server := httpserver.NewServer(":8080")
	log.Println("server listening on :8080")
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
