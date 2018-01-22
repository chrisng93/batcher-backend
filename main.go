package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/chrisng93/batcher-backend/api"
	flags "github.com/jessevdk/go-flags"
)

type flagOptions struct {
	Port string `long:"port" description:"The port for the server to run on." default:"8080" required:"false"`
}

var options flagOptions

func main() {
	options = flagOptions{}
	_, err := flags.Parse(&options)
	if err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	router := api.Init()
	err = http.ListenAndServe(fmt.Sprintf(":%s", options.Port), corsMiddleware(router))
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers:", "Origin, Content-Type, X-Auth-Token, Authorization")
			w.Header().Set("Content-Type", "application/json")
			return
		}

		next.ServeHTTP(w, r)
	})
}
