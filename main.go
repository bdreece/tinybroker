package main

import (
	"github.com/bdreece/tinybroker/tb"
)

func main() {
	broker := new(tb.Broker)
	broker.Start(8000, 1, 32)
	for {
	}
}
