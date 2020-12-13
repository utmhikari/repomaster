package app

import (
	"log"
	"net/http"
	"strconv"
)

// Config is the cfg template
type Config struct {
	Port int
}

// Start is the entry to start the web app
func Start(config Config) error {
	log.Printf("Start web app with config: %v\n", config)
	log.Println("Register routers...")
	router := router()
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(config.Port),
		Handler: router,
	}
	log.Println("Start server...")
	return server.ListenAndServe()
}
