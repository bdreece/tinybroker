package tb

import (
	"log"
	"net"
)

type Broker struct {
	listener net.Listener
	topics   map[string]*Topic
	clients  []*Client
}

func Recoverer(maxPanics int, loc string, f func()) {
	defer func() {
		if v := recover(); v != nil {
			log.Println("[PANIC] (", loc, "):", v)

			if maxPanics == 0 {
				panic("[ABORT] (connection/service): Too many panics!")
			} else {
				go Recoverer(maxPanics-1, loc, f)
			}
		}
	}()
	f()
}

func (b *Broker) accept(c chan<- *Client, verbose *bool, packetLen int) {
	for {
		// Accept a new connection
		conn, err := b.listener.Accept()
		if err != nil {
			log.Println("[ERR] (Broker/accept):", err)
		}

		if *verbose {
			log.Println("[LOG] (Broker/accept): Accepted new connection!")
		}

		// Initialize client
		client := NewClient(conn, verbose, packetLen)

		// Append to list of clients
		b.clients = append(b.clients, client)

		// Send client to broker service routine
		c <- client
	}
}

func (b *Broker) service(c <-chan *Client, verbose *bool, dataLen int) {
	if *verbose {
		log.Println("[LOG] (Broker/service) Started service routine")
	}

	for {
		// Append accepted connections
		if len(c) > 0 {
			client := <-c
			b.clients = append(b.clients, client)
		}

		// Iteratively receive packets from all clients
		for _, client := range b.clients {
			if len(client.InBuf) > 0 {
				// Retrieve packet from client
				packet := <-client.InBuf

				if *verbose {
					log.Println("[LOG] (Broker/service) Received packet")
				}

				switch packet.PacketType {
				// Pub
				case Packet_PUBLISH:
					topic := b.topics[packet.Topic]
					if topic == nil {
						// Topic doesn't exist, create it
						b.topics[packet.Topic] = NewTopic(packet.Topic, verbose, dataLen)
					}
					// Send data to topic service routine
					topic.Buf <- packet.Data
				// Sub
				case Packet_SUBSCRIBE:
					topic := b.topics[packet.Topic]
					if topic == nil {
						// Topic doesn't exist, create it
						topic = NewTopic(packet.Topic, verbose, dataLen)
					}
					topic.Sub <- client
				// Other
				default:
					log.Println("[ERR] (Broker/service) Invalid packet type")
				}
			}
		}
	}
}

func (b *Broker) Start(addr string, verbose *bool, connLen, packetLen, dataLen int) {
	var err error

	// Open listening socket
	b.listener, err = net.Listen("tcp4", addr)
	if err != nil {
		log.Fatalln("[ABORT] (Broker/Start):", err)
	}

	if *verbose {
		log.Println("[LOG] (Broker/Start): Opened listening socket")
	}
	// Launch accept and service goroutines through panic handler
	c := make(chan *Client, connLen)
	go Recoverer(5, "(Broker/accept)", func() { b.accept(c, verbose, packetLen) })
	go Recoverer(5, "(Broker/service)", func() { b.service(c, verbose, dataLen) })
}
