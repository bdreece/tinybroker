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

const VERSION_MESSAGE string = "tinybroker v0.2-alpha"

func validateEndpoint(args []string) error {
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
		version         *bool
		addr            *string
		loginEndpoint   *string
		endpointPrefix  *string
		keyFile         *string
		certFile        *string
		verbose         *int
		topicCapacity   *int
		jwtTimeout      *int
		writeTimeout    *int
		readTimeout     *int
		shutdownTimeout *int
	)

	// Parse command-line flags
	parser := argparse.NewParser("tinybroker", "A simple message broker, written in Go")

	version = parser.Flag("v", "version", &argparse.Options{
		Required: false,
		Help:     "Display version information and exit",
		Default:  false,
	})

	verbose = parser.FlagCounter("V", "verbose", &argparse.Options{
		Required: false,
		Help:     "Enable verbose output",
		Default:  0,
	})

	addr = parser.String("a", "address", &argparse.Options{
		Required: false,
		Help:     "Address over which broker is served",
		Default:  ":8080",
	})

	endpointPrefix = parser.String("p", "endpoint-prefix", &argparse.Options{
		Required: false,
		Validate: validateEndpoint,
		Help:     "Prefix for login and topic endpoints",
		Default:  "/tb",
	})

	loginEndpoint = parser.String("l", "login-endpoint", &argparse.Options{
		Required: false,
		Validate: validateEndpoint,
		Help:     "API endpoint for JWT authentication",
		Default:  "/login",
	})

	topicCapacity = parser.Int("t", "topic-capacity", &argparse.Options{
		Required: false,
		Help:     "Topic backlog capacity",
		Default:  32,
	})

	keyFile = parser.String("k", "key-file", &argparse.Options{
		Required: false,
		Help:     "TLS key file",
		Default:  "",
	})

	certFile = parser.String("c", "cert-file", &argparse.Options{
		Required: false,
		Help:     "TLS cert file",
		Default:  "",
	})

	jwtTimeout = parser.Int("j", "jwt-timeout", &argparse.Options{
		Required: false,
		Help:     "JWT lifetime duration (seconds)",
		Default:  3600,
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

	shutdownTimeout = parser.Int("s", "shutdown-timeout", &argparse.Options{
		Required: false,
		Help:     "HTTP server kill signal timeout (seconds)",
		Default:  5,
	})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	if *version {
		fmt.Println(VERSION_MESSAGE)
		os.Exit(0)
	}

	if *verbose > 0 {
		log.Println("[LOG] Starting tinybroker")
		log.Println("[LOG] Configuring router URL handler")
	}

	srv := configureServer(addr, loginEndpoint, endpointPrefix,
		time.Duration(int64(*jwtTimeout))*time.Second,
		time.Duration(int64(*writeTimeout))*time.Second,
		time.Duration(int64(*readTimeout))*time.Second,
		topicCapacity, verbose)

	if *verbose > 0 {
		log.Println("[LOG] Starting server")
	}

	if *certFile == "" || *keyFile == "" {
		if *verbose > 1 {
			log.Println("[LOG] Missing key file or cert file")
		}
		if *verbose > 0 {
			log.Println("[LOG] Launching HTTP server")
		}
		launchHTTPServer(&srv)
	} else {
		if *verbose > 1 {
			log.Println("[LOG] Found key file and cert file")
		}
		if *verbose > 0 {
			log.Println("[LOG] Launching HTTPS server")
		}
		launchHTTPSServer(&srv, certFile, keyFile)
	}

	if *verbose > 0 {
		log.Println("[LOG] Shutdown signal received")
	}

	// Shutdown timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*shutdownTimeout)*time.Second)
	defer cancel()

	if *verbose > 0 {
		log.Println("[LOG] Starting shutdown procedure")
	}

	shutdownProcedure(&srv, ctx)

	log.Println("[LOG] Goodbye!")
	os.Exit(0)
}
