package main

import (
	"fmt"
	"integration/handlers"
	"log"
	"net/http"
)

func main() {
	// Creates a simple HTTP server used for the demo purpouses
	server := &http.Server{
		Addr:    fmt.Sprintf(":8000"),
		Handler: handlers.New(),
	}

	log.Printf("Starting the HTTP Server at port %q", server.Addr)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("%v", err)
	} else {
		log.Println("Server closed.")
	}
}
