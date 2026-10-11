package main

import (
	"log"

	"app_go/internal/searchmedia"
)

func main() {
	config := searchmedia.ConfigFromEnv()
	server := searchmedia.NewServer(config)

	log.Printf("searchmedia listening on %s", config.HTTPAddr)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("searchmedia: %v", err)
	}
}
