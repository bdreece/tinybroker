package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

func getEnvironmentVars(verbose *int) (*string, *string, *string, error) {
	var (
		username string = os.Getenv("TB_USER")
		password string = os.Getenv("TB_PASS")
		secret   string = os.Getenv("TB_SECRET")
	)

	if username == "" || password == "" || secret == "" {
		if *verbose > 0 {
			log.Println("[ERR] Empty environment variables!")
		}
		return &username, &password, &secret, errors.New("Failed to retrieve some environment variables")
	}

	return &username, &password, &secret, nil
}

func configureServer(addr, authEndpoint, endpointPrefix *string,
	jwtTimeout, writeTimeout, readTimeout time.Duration,
	capacity, verbose *int) http.Server {

	// Get environment variables
	username, password, secret, err := getEnvironmentVars(verbose)
	if err != nil {
		if *verbose > 1 {
			log.Printf("[ERR] TB_USER: %s\n", *username)
			log.Printf("[ERR] TB_PASS: %s\n", *password)
			log.Printf("[ERR] TB_SECRET: %s\n", *secret)
		}
		log.Fatalf("[ERR] %v\n", err.Error())
	}

	// Configure router and server
	router := mux.NewRouter()
	routeHandler := NewHandler(capacity, verbose)
	mw := NewMiddleware(username, password, secret, &jwtTimeout, verbose)

	router.Handle(*endpointPrefix+*authEndpoint, mw).
		Methods("POST")

	router.Handle(*endpointPrefix+"/{topic}", mw.AuthMiddleware(routeHandler)).
		Methods("POST", "GET", "PUT", "DELETE")

	if *verbose > 1 {
		log.Printf("[LOG] Configured server for address: %s\n", *addr)
	}

	return http.Server{
		Handler:      router,
		Addr:         *addr,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
	}
}

func launchHTTPServer(srv *http.Server) {
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

func launchHTTPSServer(srv *http.Server, certFile, keyFile *string) {
	// Launch server asynchronously
	go func() {
		if err := srv.ListenAndServeTLS(*certFile, *keyFile); err != nil {
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
