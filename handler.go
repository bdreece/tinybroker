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
	"encoding/json"
	"net/http"
	"time"

	"github.com/bdreece/go-structs/ringbuf"
	"github.com/bdreece/tattle"
	"github.com/gorilla/mux"
)

type Handler struct {
	Topics   map[string]*ringbuf.RingBuf[Message]
	Capacity *int
	Verbose  *int
	Logger   *tattle.Logger
}

type Message struct {
	Time	 	time.Time	`json:"time"`
	Data 		string		`json:"data"`
}

const (
	MESSAGE_TIME_FIELD string = "time"
	MESSAGE_DATA_FIELD string = "data"
)

func NewHandler(capacity *int, verbose *int, logger *tattle.Logger) Handler {
	return Handler{
		Topics:   make(map[string]*ringbuf.RingBuf[Message]),
		Capacity: capacity,
		Verbose:  verbose,
		Logger:   logger,
	}
}

func newMessage(data []byte) Message {
	return Message {
		Time: time.Now(),
		Data: string(data),
	}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	topic := mux.Vars(r)["topic"]

	if *h.Verbose > 0 {
		h.Logger.Logf("Serving request on topic: %s\n", topic)
	}

	switch r.Method {
	case "POST":
		h.CreateResponse(w, r, topic)
	case "GET":
		h.ReadResponse(w, r, topic)
	case "PUT":
		h.UpdateResponse(w, r, topic)
	case "DELETE":
		h.DeleteResponse(w, r, topic)
	default:
		if *h.Verbose > 0 {
			h.Logger.Logf("Invalid request method: %s\n", r.Method)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if *h.Verbose > 0 {
		h.Logger.Logln("Sent response!")
	}
}

func (h Handler) CreateResponse(w http.ResponseWriter, r *http.Request, topic string) {
	if *h.Verbose > 1 {
		h.Logger.Logf("Creating topic: %s\n", topic)
	}

	// Create topic if it doesn't exist
	if _, ok := h.Topics[topic]; !ok {
		newTopic := ringbuf.New[Message](*h.Capacity)
		h.Topics[topic] = &newTopic
	}

	// Read raw body
	data, n := ReadBody(r, h.Logger)

	if len(data) > 0 {
		if *h.Verbose > 1 {
			h.Logger.Logf("Create request contains data: %s\n", data)
		}

		// Create message on topic queue
		h.Topics[topic].Write(newMessage(data[:n]))
	}

	// No error on empty body
	w.WriteHeader(http.StatusOK)
}

func (h Handler) ReadResponse(w http.ResponseWriter, r *http.Request, topic string) {
	if *h.Verbose > 1 {
		h.Logger.Logf("Reading from topic: %s\n", topic)
	}

	// Check if topic queue exists
	if h.Topics[topic] == nil {
		if *h.Verbose > 1 {
			h.Logger.Logf("Topic %s not found!\n", topic)
		}
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Get message from topic queue
	data, err := h.Topics[topic].Read()

	// Empty queue
	if err != nil {
		if *h.Verbose > 1 {
			h.Logger.Logln("Topic contains no data")
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	msg := data.(Message)

	if *h.Verbose > 1 {
		h.Logger.Logf("Read data: %v\n", msg.Data)
	}

	// Marshal to JSON
	msgJson, err := json.Marshal(msg)
	if err != nil {
		// Message should be serializable, failure not on behalf of user
		h.Logger.Errf("Error marshalling JSON: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(msgJson)
	if err != nil {
		h.Logger.Errf("Error writing response: %s\n", err.Error())
	}
}

func (h Handler) UpdateResponse(w http.ResponseWriter, r *http.Request, topic string) {
	if *h.Verbose > 1 {
		h.Logger.Logf("Updating topic: %s\n", topic)
	}

	// Check if topic queue exists
	if h.Topics[topic] == nil {
		if *h.Verbose > 1 {
			h.Logger.Logf("Topic %s not found!\n", topic)
		}
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Read raw body
	data, n := ReadBody(r, h.Logger)

	// Error on empty body
	if len(data) < 1 {
		if *h.Verbose > 1 {
			h.Logger.Logln("Data not found!")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if *h.Verbose > 1 {
		h.Logger.Logf("Updating with data: %s\n", data)
	}

	// Create message on topic queue
	h.Topics[topic].Write(newMessage(data[:n]))
	w.WriteHeader(http.StatusOK)
}

func (h Handler) DeleteResponse(w http.ResponseWriter, r *http.Request, topic string) {
	if *h.Verbose > 1 {
		h.Logger.Logf("Deleting topic: %s\n", topic)
	}

	// Check if topic exists
	if h.Topics[topic] == nil {
		if *h.Verbose > 1 {
			h.Logger.Logf("Topic %s not found!\n")
		}
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Remove topic queue from map
	delete(h.Topics, topic)
	w.WriteHeader(http.StatusOK)
}
