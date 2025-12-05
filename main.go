package main

import (
	"log"

	"github.com/JpUnique/TinyGo/cmd"
)

func main() {
	srv, err := cmd.New()
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}
