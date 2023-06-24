package main

import (
	"fmt"
	"log"
	"net/http"
)

const webPort = "8084"

// Receiver that we use for the application
type Config struct{}

func main() {
	app := Config{}

	log.Printf("Starting broker service on port %s \n", webPort)

	// define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// start the server
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}

}
