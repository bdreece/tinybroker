package tb

import (
	"google.golang.org/protobuf/proto"
	"log"
	"net"
)

type Client struct {
	InBuf  chan *Packet
	OutBuf chan *Packet
	Conn   net.Conn
}

func (c *Client) read(verbose *bool) {
	// Thread-local variables
	var (
		bytes []byte
		err   error
		n     int
	)

	// Forever loop
	for {
		// Read from socket
		n, err = c.Conn.Read(bytes)
		if err != nil {
			log.Println("[ERR] (Client/service/read):", err)
		}

		if *verbose {
			log.Printf("[LOG] (Client/service/read): Received %d bytes\n", n)
		}

		// Parse packet
		packet := new(Packet)
		err = proto.Unmarshal(bytes, packet)
		if err != nil {
			log.Println("[ERR] (Client/service/read):", err)
		}

		if *verbose {
			log.Println("[LOG] (Client/service/read): Parsed bytes")
		}
		// Send packet over inBuf channel
		c.InBuf <- packet
	}
}

func (c *Client) write(verbose *bool) {
	// Thread-local variables
	var (
		bytes []byte
		err   error
		sent  int
	)

	// Forever loop (until channel closes)
	for packet := range c.OutBuf {
		bytes, err = proto.Marshal(packet)
		if err != nil {
			log.Println("[ERR] (Client/service/write):", err)
		}

		if *verbose {
			log.Println("[LOG] (Client/service/write): Serialized packet")
		}

		// Send all bytes
		for n := 0; n < len(bytes); {
			sent, err = c.Conn.Write(bytes)
			if err != nil {
				log.Println("[ERR] (Client/service/write):", err)
			}
			n += sent
		}

		if *verbose {
			log.Printf("[LOG] (Client/service/write): Sent %d bytes\n", sent)
		}
	}
}

func (c *Client) Service(verbose *bool) {
	// Read into inBuf channel
	go Recoverer(5, "Client/service/read", func() { c.read(verbose) })

	// Write from outBuf channel
	go Recoverer(5, "Client/service/write", func() { c.write(verbose) })
}

func NewClient(conn net.Conn, verbose *bool, queueLen int) *Client {
	c := new(Client)
	c.InBuf = make(chan *Packet, queueLen)
	c.OutBuf = make(chan *Packet, queueLen)
	c.Conn = conn
	c.Service(verbose)

	return c
}
