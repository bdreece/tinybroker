package main

import (
  "context"
  "errors"
  "flag"
  "log"
  "net/http"
  "os"
  "os/signal"
  "time"

  "github.com/gorilla/mux"
  "github.com/bdreece/tinybroker/handler"
  "github.com/bdreece/tinybroker/middleware"
)

func getEnvironmentVars(verbose bool) (string, string, string, error) {
  var (
    username string = os.Getenv("TB_USER")
    password string = os.Getenv("TB_PASS")
    secret string = os.Getenv("TB_SECRET")
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

func main() {
  var (
    addr string
    verbose bool
    capacity int
    writeTimeout int64
    readTimeout int64
    killTimeout int64
  )

  // Parse command-line flags
  flag.BoolVar(&verbose, "v", false, "Enable verbose output")
  flag.IntVar(&capacity, "c", 32, "Topic queue capacity")
  flag.Int64Var(&writeTimeout, "w", 5, "HTTP write timeout (in seconds)")
  flag.Int64Var(&readTimeout, "r", 5, "HTTP read timeout (in seconds)")
  flag.Int64Var(&killTimeout, "k", 5, "Server kill timeout (in seconds)")
  flag.StringVar(&addr, "a", "127.0.0.1:8080", "Listening address and port")
  flag.Parse()

  if verbose {
    log.Println("[LOG] Starting tinybroker")
    log.Println("[LOG] Configuring router URL handler")
  }

  srv := configureServer(addr, verbose, time.Duration(writeTimeout) * time.Second, time.Duration(readTimeout) * time.Second, capacity)

  if verbose {
    log.Println("[LOG] Starting server")
  }

  launchServer(&srv)

  if verbose {
    log.Println("[LOG] Shutdown signal received")
  }

  // Shutdown timeout
  ctx, cancel := context.WithTimeout(context.Background(), time.Duration(killTimeout) * time.Second)
  defer cancel()

  if verbose {
    log.Println("[LOG] Starting shutdown procedure")
  }

  shutdownProcedure(&srv, ctx)

  log.Println("[LOG] Goodbye!")
  os.Exit(0)
}
