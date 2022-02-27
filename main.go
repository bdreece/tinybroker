package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/akamensky/argparse"
)

func validateAuthEndpoint(args []string) error {
	if !strings.HasPrefix(args[0], "/") {
		return errors.New("Invalid endpoint string, must begin with '/'")
	}

	if strings.Contains(args[0], "?") {
		return errors.New("Invalid endpoint string, cannot contain '?'")
	}

	return nil
}

func main() {
	var (
		addr          *string
		authEndpoint  *string
		verbose       *int
		topicCapacity *int
		writeTimeout  *int
		readTimeout   *int
		killTimeout   *int
	)

	// Parse command-line flags
	parser := argparse.NewParser("tinybroker", "A simple message broker, written in Go")

	addr = parser.String("a", "address", &argparse.Options{
		Required: false,
		Help:     "Address to serve broker on (address:port)",
		Default:  ":8080",
	})

	authEndpoint = parser.String("e", "auth-endpoint", &argparse.Options{
		Required: false,
		Validate: validateAuthEndpoint,
		Help:     "API endpoint for JWT authentication",
		Default:  "/login",
	})

	verbose = parser.FlagCounter("v", "verbose", &argparse.Options{
		Required: false,
		Help:     "Enable verbose output",
		Default:  0,
	})

	topicCapacity = parser.Int("c", "topic-capacity", &argparse.Options{
		Required: false,
		Help:     "Topic backlog capacity",
		Default:  32,
	})

	writeTimeout = parser.Int("w", "write-timeout", &argparse.Options{
		Required: false,
		Help:     "HTTP server write timeout (seconds)",
		Default:  5,
	})

	readTimeout = parser.Int("r", "read-timeout", &argparse.Options{
		Required: false,
		Help:     "HTTP server read timeout (seconds)",
		Default:  5,
	})

	killTimeout = parser.Int("k", "kill-timeout", &argparse.Options{
		Required: false,
		Help:     "HTTP server kill signal timeout (seconds)",
		Default:  5,
	})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	if *verbose > 0 {
		log.Println("[LOG] Starting tinybroker")
		log.Println("[LOG] Configuring router URL handler")
	}

	srv := configureServer(addr, authEndpoint,
		time.Duration(int64(*writeTimeout))*time.Second,
		time.Duration(int64(*readTimeout))*time.Second,
		topicCapacity, verbose)

	if *verbose > 0 {
		log.Println("[LOG] Starting server")
	}

	launchServer(&srv)

	if *verbose > 0 {
		log.Println("[LOG] Shutdown signal received")
	}

	// Shutdown timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*killTimeout)*time.Second)
	defer cancel()

	if *verbose > 0 {
		log.Println("[LOG] Starting shutdown procedure")
	}

	shutdownProcedure(&srv, ctx)

	log.Println("[LOG] Goodbye!")
	os.Exit(0)
}
