package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/bdreece/tinybroker/handler"
	"github.com/bdreece/tinybroker/middleware"
	"github.com/gorilla/mux"
)

func getEnvironmentVars(verbose bool) (string, string, string, error) {
	var (
		username string = os.Getenv("TB_USER")
		password string = os.Getenv("TB_PASS")
		secret   string = os.Getenv("TB_SECRET")
	)

	if username == "" || password == "" || secret == "" {
		if verbose {
			log.Println("[ERR] Empty environment variables!")
		}
		return username, password, secret, errors.New("Failed to retrieve some environment variables")
	}

	return username, password, secret, nil
}

func configureServer(addr string, verbose bool, writeTimeout, readTimeout time.Duration, capacity int) http.Server {
	// Get environment variables
	username, password, secret, err := getEnvironmentVars(verbose)
	if err != nil {
		log.Fatalf("[ERR] %v\n", err.Error())
	}

	// Configure router and server
	router := mux.NewRouter()

	routeHandler := handler.NewHandler(capacity, verbose)

	authMw := middleware.NewMiddleware(username, password, secret, verbose)

	router.Handle("/login", authMw).
		Methods("POST")

	router.Handle("/{topic}", authMw.AuthMiddleware(routeHandler)).
		Methods("POST", "GET", "PUT", "DELETE")

	return http.Server{
		Handler:      router,
		Addr:         addr,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
	}
}

func launchServer(srv *http.Server) {
	// Launch server asynchronously
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("[ERR] %v\n", err.Error())
		}
	}()

	// Setup channel for signal handling
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until signal
	<-c
}

func shutdownProcedure(srv *http.Server, ctx context.Context) {
	// Shutdown procedure
	go func() {
		srv.Shutdown(ctx)
	}()

	<-ctx.Done()
}
