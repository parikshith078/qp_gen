package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var webPort = os.Getenv("WEB_PORT")

type Config struct {
}

func main() {
	app := Config{}

	log.Printf("Starting broker service on port %s\n", webPort)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
