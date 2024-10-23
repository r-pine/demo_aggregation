package server

import (
	"errors"
	"net/http"
	"os"
	"os/signal"

	"github.com/r-pine/demo_aggregation/app/pkg/logging"
)

func RunServer(
	log logging.Logger, handler http.Handler, httpAddr string,
) {
	server := &http.Server{
		Addr:    httpAddr,
		Handler: handler,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		log.Infoln("receive interrupt signal")
		if err := server.Close(); err != nil {
			log.Fatal("Server Close:", err)
		}
	}()
	log.Infoln("Start server", httpAddr)
	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Infoln("Server closed under request")
		} else {
			log.Fatal("Server closed unexpect")
		}
	}

	log.Infoln("Server exiting")
}
