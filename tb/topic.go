package tb

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
)

type Topic struct {
	Name string
	Buf  chan []byte
	Sub  chan *Client
	subs []*Client
	next int
}

func NewTopic(name string, verbose *bool, queueLen int) *Topic {
	topic := &Topic{}

	topic.Name = name
	topic.Buf = make(chan []byte, queueLen)
	topic.Sub = make(chan *Client, queueLen)
	topic.subs = make([]*Client, queueLen)
	topic.next = 0

	go Recoverer(5, "Topic/service-"+name, func() { topic.service(verbose) })

	return topic
}

func (t *Topic) service(verbose *bool) {
	for {
		select {
		case data := <-t.Buf:
			if *verbose {
				log.Printf("[LOG] (Topic/service-%s): Received %d bytes\n", t.Name, len(data))
			}

			// Create response packet
			packet := new(Packet)
			packet.PacketType = Packet_RESPONSE
			packet.Topic = t.Name
			packet.TimeStamp = timestamppb.Now()
			packet.Data = data

			// Send to next sub
			t.subs[t.next].OutBuf <- packet
			t.next += 1

			// Reset next to zero after reaching end
			if t.next >= len(t.subs) {
				t.next = 0
			}

		case client := <-t.Sub:
			// Add client to subs
			t.subs = append(t.subs, client)

		default:
			if *verbose {
				log.Printf("[LOG] (Topic/service-%s): No activity", t.Name)
			}
		}
	}
}
