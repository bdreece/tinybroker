package main

import (
  "context"
  "flag"
  "log"
  "os"
  "time"
)

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
