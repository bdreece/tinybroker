/* tinybroker - A simple message broker, written in Go
Copyright (C) 2022 Brian Reece

This program is free software; you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation; either version 2 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License along
with this program; if not, write to the Free Software Foundation, Inc.,
51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/akamensky/argparse"
	"github.com/bdreece/tattle"
)

type Args struct {
	Version         *bool
	Addr            *string
	LoginEndpoint   *string
	EndpointPrefix  *string
	KeyFile         *string
	CertFile        *string
	Verbose         *int
	TopicCapacity   *int
	JwtTimeout      *int
	WriteTimeout    *int
	ReadTimeout     *int
	ShutdownTimeout *int
}

const VERSION_MESSAGE string = "tinybroker v0.2-alpha"

func validateEndpoint(args []string) error {
	if !strings.HasPrefix(args[0], "/") {
		return errors.New("invalid endpoint string, must begin with '/'")
	}
	if strings.Contains(args[0], "?") {
		return errors.New("invalid endpoint string, cannot contain '?'")
	}
	return nil
}

func parseArgs(args *Args) {
	// Parse command-line flags
	parser := argparse.NewParser("tinybroker", "A simple message broker, written in Go")

	args.Version = parser.Flag("v", "version", &argparse.Options{
		Required: false,
		Help:     "Display version information and exit",
		Default:  false,
	})

	args.Verbose = parser.FlagCounter("V", "verbose", &argparse.Options{
		Required: false,
		Help:     "Enable verbose output",
		Default:  0,
	})

	args.Addr = parser.String("a", "address", &argparse.Options{
		Required: false,
		Help:     "Address over which broker is served",
		Default:  ":8080",
	})

	args.EndpointPrefix = parser.String("p", "endpoint-prefix", &argparse.Options{
		Required: false,
		Validate: validateEndpoint,
		Help:     "Prefix for login and topic endpoints",
		Default:  "/tb",
	})

	args.LoginEndpoint = parser.String("l", "login-endpoint", &argparse.Options{
		Required: false,
		Validate: validateEndpoint,
		Help:     "API endpoint for JWT authentication",
		Default:  "/login",
	})

	args.TopicCapacity = parser.Int("t", "topic-capacity", &argparse.Options{
		Required: false,
		Help:     "Topic backlog capacity",
		Default:  32,
	})

	args.KeyFile = parser.String("k", "key-file", &argparse.Options{
		Required: false,
		Help:     "TLS key file",
		Default:  "",
	})

	args.CertFile = parser.String("c", "cert-file", &argparse.Options{
		Required: false,
		Help:     "TLS cert file",
		Default:  "",
	})

	args.JwtTimeout = parser.Int("j", "jwt-timeout", &argparse.Options{
		Required: false,
		Help:     "JWT lifetime duration (seconds)",
		Default:  3600,
	})

	args.WriteTimeout = parser.Int("w", "write-timeout", &argparse.Options{
		Required: false,
		Help:     "HTTP server write timeout (seconds)",
		Default:  5,
	})

	args.ReadTimeout = parser.Int("r", "read-timeout", &argparse.Options{
		Required: false,
		Help:     "HTTP server read timeout (seconds)",
		Default:  5,
	})

	args.ShutdownTimeout = parser.Int("s", "shutdown-timeout", &argparse.Options{
		Required: false,
		Help:     "HTTP server kill signal timeout (seconds)",
		Default:  5,
	})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}
}

func main() {
	args := new(Args)
	logger := tattle.NewLogger("LOG", "WARN", "ERR", nil)
	parseArgs(args)

	if *args.Version {
		fmt.Println(VERSION_MESSAGE)
		os.Exit(0)
	}

	if *args.Verbose > 0 {
		logger.Logln("Starting tinybroker")
		logger.Logln("Configuring router URL handler")
	}

	srv := configureServer(args.Addr, args.LoginEndpoint, args.EndpointPrefix,
		time.Duration(int64(*args.JwtTimeout))*time.Second,
		time.Duration(int64(*args.WriteTimeout))*time.Second,
		time.Duration(int64(*args.ReadTimeout))*time.Second,
		args.TopicCapacity, args.Verbose, &logger)

	if *args.Verbose > 0 {
		logger.Logln("Starting server")
	}

	if *args.CertFile == "" || *args.KeyFile == "" {
		if *args.Verbose > 1 {
			logger.Logln("Missing key file or cert file")
		}
		if *args.Verbose > 0 {
			logger.Logln("Launching HTTP server")
		}
		launchHTTPServer(&srv, &logger)
	} else {
		if *args.Verbose > 1 {
			logger.Logln("Found key file and cert file")
		}
		if *args.Verbose > 0 {
			logger.Logln("Launching HTTPS server")
		}
		launchHTTPSServer(&srv, args.CertFile, args.KeyFile, &logger)
	}

	if *args.Verbose > 0 {
		logger.Logln("Shutdown signal received")
	}

	// Shutdown timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*args.ShutdownTimeout)*time.Second)
	defer cancel()

	if *args.Verbose > 0 {
		logger.Logln("Starting shutdown procedure")
	}

	shutdownProcedure(&srv, ctx)

	logger.Logln("Goodbye!")
	os.Exit(0)
}
