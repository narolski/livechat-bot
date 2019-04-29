package main

import (
	"fmt"
	"integration/handlers"
	"log"
	"net/http"
)

func main() {
	// We create a simple server using http.Server and run.
	server := &http.Server{
		Addr:    fmt.Sprintf(":8000"),
		Handler: handlers.New(),
	}

	log.Printf("Starting HTTP Server. Listening at %q", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("%v", err)
	} else {
		log.Println("Server closed!")
	}
}
