package app

import (
	"log"
	"net/http"
)

func StartApp() {
	log.Println("Register routers...")
	router := Router()
	server := &http.Server{
		Addr: ":8000",
		Handler: router,
	}
	log.Println("Start server...")
	err := server.ListenAndServe()
	if err != nil{
		panic(err)
	}
}
