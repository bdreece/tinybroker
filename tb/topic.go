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

func NewTopic(name string, queueLen int) *Topic {
	topic := &Topic{}

	topic.Name = name
	topic.Buf = make(chan []byte, queueLen)
	topic.Sub = make(chan *Client, queueLen)
	topic.subs = make([]*Client, queueLen)
	topic.next = 0

	go Recoverer(5, "Topic/service-"+name, func() { topic.service() })

	return topic
}

func (t *Topic) service() {
	for {
		select {
		case data := <-t.Buf:
			// Create response packet
			packet := Packet{}
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
			log.Println("[LOG] (Topic/service-" + t.Name + "): No activity")
		}
	}
}
