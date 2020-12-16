package app

import (
	"github.com/utmhikari/repomaster/internal/service/cfg"
	"log"
	"net/http"
	"strconv"
)


// Start is the entry to start the web app
func Start(cfgPath string) error {
	// init config
	err := cfg.InitGlobalConfig(cfgPath)
	if err != nil{
		return err
	}
	log.Printf("Start repomaster app with config: %+v\n", cfg.GlobalCfg)
	// init router
	router := router()
	// init server
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.GlobalCfg.Port),
		Handler: router,
	}
	// launch repo root
	log.Println("Start repomaster server...")
	return server.ListenAndServe()
}
