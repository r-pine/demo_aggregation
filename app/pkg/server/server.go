package server

import (
	"errors"
	"github.com/r-pine/demo_aggregation/app/pkg/logging"
	"net/http"
	"os"
	"os/signal"
)

func RunServer(
	log logging.Logger, handler http.Handler, httpAddr string,
) {
	server := &http.Server{
		Addr:    httpAddr,
		Handler: handler,
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		log.Infoln("receive interrupt signal")
		if err := server.Close(); err != nil {
			log.Fatal("Server Close:", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Infoln("Server closed under request")
		} else {
			log.Fatal("Server closed unexpect")
		}
	}

	log.Infoln("Server exiting")
}
