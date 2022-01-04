package main

import (
	"flag"
	"github.com/bdreece/tinybroker/tb"
	"log"
	"os"
)

const (
	VERSION string = "tinybroker v0.1-alpha-20220103"
	LICENSE string = ""
	HELP    string = VERSION + `
  
  Usage: tinybroker [OPTIONS] <PORT>
  
  Options:
  --addr ADDR
  `
)

func main() {
	var (
		addr                        string
		connLen, packetLen, dataLen int
		version, verbose            bool
	)

	flag.StringVar(&addr, "a", "127.0.0.1:8000", "IP address to listen from")
	flag.IntVar(&connLen, "c", 1, "Maximum connection backlog")
	flag.IntVar(&packetLen, "p", 8, "Maximum packet backlog")
	flag.IntVar(&dataLen, "d", 32, "Maximum topic data backlog")
	flag.BoolVar(&version, "v", false, "Show version message and exit")
	flag.BoolVar(&verbose, "V", false, "Show verbose output")

	flag.Parse()

	if version {
		log.Println(VERSION)
		log.Println(LICENSE)
		os.Exit(0)
	}

	broker := new(tb.Broker)
	broker.Start(addr, &verbose, connLen, packetLen, dataLen)
	for {
	}
}
