package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/chrisng93/batcher-backend/api"
	"github.com/gorilla/handlers"
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
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", options.Port), handlers.CORS()(router)))
}
