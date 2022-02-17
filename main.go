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
	rest "github.com/bdreece/tinybroker/resthandlers"
	"time"
)

const (
  writeTimeout time.Duration = 15 * time.Second
  readTimeout = 15 * time.Second
  killTimeout = 5 * time.Second
)

var rdb redis.Client

func methodRouter(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "POST":
      rest.CreateHandler{&rdb}.ServeHTTP(w, r)
    case "GET":
      rest.ReadHandler{&rdb}.ServeHTTP(w, r)
    case "PUT":
      rest.UpdateHandler{&rdb}.ServeHTTP(w, r)
    case "DELETE":
      rest.DeleteHandler{&rdb}.ServeHTTP(w, r)
    default: log.Printf("[ERR] Invalid request method: %s\n", r.Method)
  }
}

func configureServer(addr string, writeTimeout, readTimeout time.Duration) http.Server {
  // Configure router and server
  router := mux.NewRouter()
  router.HandleFunc("/tb/{topic}", methodRouter).
         Methods("POST", "GET", "PUT", "DELETE")

  return http.Server{
    Handler:    router,
    Addr:       addr,
    WriteTimeout: writeTimeout,
    ReadTimeout: readTimeout,
  }
}

func serveHTTP(srv *http.Server) {
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

  serveHTTP(&srv)

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
