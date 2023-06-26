// servemux (also known as a router) stores a mapping between the predefined URL paths for your application and the corresponding handlers. Usually you have one servemux for your application containing all your routes.

// Handlers are responsible for carrying out your application logic and writing response headers and bodies.

// http.ServeMux is An HTTP request multiplexer, often called a router, is responsible for routing incoming HTTP requests to the appropriate handler functions based on the request's URL or other criteria.

// http.DefaultServeMux is the default instance of http.ServeMux created by Go's net/http package. When you register your handler functions using functions like http.HandleFunc without explicitly specifying a custom http.ServeMux, they are automatically registered with http.DefaultServeMux.

package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()

	// specify who is allowed to connect
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	mux.Use(middleware.Heartbeat("/ping"))

	mux.Post("/log", app.WriteLog)

	return mux
}
