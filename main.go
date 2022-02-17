package main

import (
	"context"
    "flag"
	"log"
    "github.com/gorilla/mux"
	"net/http"
	"os"
	"os/signal"
	redis "github.com/go-redis/redis/v8"
	rest "github.com/bdreece/tinybroker/handler"
	"time"
)

const (
  writeTimeout time.Duration = 15 * time.Second
  readTimeout = 15 * time.Second
  killTimeout = 5 * time.Second
)

func configureServer(addr string, writeTimeout, readTimeout time.Duration) http.Server {
  // Configure router and server
  router := mux.NewRouter()
  
  handler := rest.Handler{
    rdb: &redis.NewClient()
  }

  router.Handle("/tb/{topic}", handler).
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
      log.Println(err)
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

func main() {
  var (
    addr string
    verbose bool
  )

  // Parse command-line flags
  flag.BoolVar(&verbose, "v", false, "Enable verbose output")
  flag.StringVar(&addr, "a", "127.0.0.1:8080", "Listening address and port")
  flag.Parse()

  if verbose {
    log.Println("[LOG] Starting tinybroker")
    log.Println("[LOG] Configuring router URL handler")
  }

  srv := configureServer(addr, writeTimeout, readTimeout)

  if verbose {
    log.Println("[LOG] Starting server")
  }

  launchServer(&srv)

  if verbose {
    log.Println("[LOG] Shutdown signal received")
  }

  // Shutdown timeout
  ctx, cancel := context.WithTimeout(context.Background(), killTimeout)
  defer cancel()

  if verbose {
    log.Println("[LOG] Starting shutdown procedure")
  }

  shutdownProcedure(&srv, ctx)

  log.Println("[LOG] Goodbye!")
  os.Exit(0)
}
