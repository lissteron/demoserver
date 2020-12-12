package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/lissteron/demoserver/config"
	"github.com/lissteron/demoserver/internal/app/controllers"
)

func main() {
	var (
		logger  = log.New(os.Stdout, "", log.Lshortfile|log.Ltime)
		handler = controllers.NewRequestHandler(
			logger,
			config.MaxInputRequests,
			config.MaxBodySize,
			config.URLsLimit,
		)

		server = &http.Server{
			Addr:     fmt.Sprintf(":%d", config.HTTPPort),
			Handler:  handler,
			ErrorLog: logger,
		}
	)

	go func() {
		logger.Printf("[info] server listen on addr: '%s'\n", server.Addr)

		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Printf("[error] ListenAndServe: %v\n", err)
			os.Exit(1)
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	<-exit

	logger.Println("[info] go to shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), config.ServerShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Printf("[error] server Shutdown: %v\n", err)
	}
}
