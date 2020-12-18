package app

import (
	cfgService "github.com/utmhikari/repomaster/internal/service/cfg"
	repoService "github.com/utmhikari/repomaster/internal/service/repo"
	"log"
	"net/http"
	"strconv"
)


// Start is the entry to start the web app
func Start(cfgPath string) error {
	// init config
	err := cfgService.InitGlobalConfig(cfgPath)
	if err != nil{
		return err
	}
	log.Printf("Start repomaster app with config: %+v\n", cfgService.GlobalCfg)
	// init web handler
	webHandler := getWebHandler()
	// init server
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(cfgService.GlobalCfg.Port),
		Handler: webHandler,
	}
	// refresh repos
	repoService.Refresh()
	// launch server
	log.Println("Start repomaster server...")
	return server.ListenAndServe()
}
