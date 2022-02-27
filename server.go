package main

import (
  "context"
  "log"
  "net/http"
  "time"
  "os"
  "os/signal"

  "github.com/bdreece/tinybroker/handler"
  "github.com/bdreece/tinybroker/middleware"
  "github.com/gorilla/mux"
)

func configureServer(addr string, verbose bool, writeTimeout, readTimeout time.Duration, capacity int) http.Server {
  // Get environment variables
  username, password, secret, err := getEnvironmentVars(verbose)
  if err != nil {
    log.Fatalf("[ERR] %v\n", err.Error())
  }

  // Configure router and server
  router := mux.NewRouter()
  
  handler := handler.NewHandler(capacity, verbose)

  authMw := middleware.NewMiddleware(username, password, secret, verbose) 

  router.Handle("/login", authMw).
         Methods("POST")

  router.Handle("/{topic}", authMw.AuthMiddleware(handler)).
         Methods("POST", "GET", "PUT", "DELETE")

  return http.Server{
    Handler:    router,
    Addr:       addr,
    WriteTimeout: writeTimeout,
    ReadTimeout: readTimeout,
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
