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
	"fmt"
	"net/http"

	"github.com/bdreece/go-structs/ringbuf"
	"github.com/bdreece/tattle"
	"github.com/gorilla/mux"
)

type Handler struct {
	Topics   map[string]*ringbuf.RingBuf
	Capacity *int
	Verbose  *int
	Logger   *tattle.Logger
}

func NewHandler(capacity *int, verbose *int, logger *tattle.Logger) Handler {
	return Handler{
		Topics:   make(map[string]*ringbuf.RingBuf),
		Capacity: capacity,
		Verbose:  verbose,
		Logger:   logger,
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
		h.Logger.Errf("Invalid request method: %s\n", r.Method)
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

	data := r.PostFormValue("TB_DATA")
	if _, ok := h.Topics[topic]; !ok {
		newTopic := ringbuf.New(*h.Capacity)
		h.Topics[topic] = &newTopic
	}

	if len(data) > 0 {
		if *h.Verbose > 1 {
			h.Logger.Logf("Create request contains data: %s\n", data)
		}

		h.Topics[topic].Write(data)
	}

	w.WriteHeader(http.StatusOK)
}

func (h Handler) ReadResponse(w http.ResponseWriter, r *http.Request, topic string) {
	if *h.Verbose > 1 {
		h.Logger.Logf("Reading from topic: %s\n", topic)
	}

	if h.Topics[topic] == nil {
		if *h.Verbose > 1 {
			h.Logger.Logf("Topic %s not found!", topic)
		}
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data := h.Topics[topic].Read()

	if data == nil {
		if *h.Verbose > 1 {
			h.Logger.Logln("Topic contains no data")
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if *h.Verbose > 1 {
		h.Logger.Logf("Read data: %v\n", data)
	}

	_, err := w.Write([]byte(fmt.Sprint(data)))
	if err != nil {
		h.Logger.Errln(err.Error())
	}
}

func (h Handler) UpdateResponse(w http.ResponseWriter, r *http.Request, topic string) {
	if *h.Verbose > 1 {
		h.Logger.Logf("Updating topic: %s\n", topic)
	}

	if h.Topics[topic] == nil {
		if *h.Verbose > 1 {
			h.Logger.Logf("Topic %s not found!\n", topic)
		}
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data := r.PostFormValue("TB_DATA")
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

	h.Topics[topic].Write(data)
	w.WriteHeader(http.StatusOK)
}

func (h Handler) DeleteResponse(w http.ResponseWriter, r *http.Request, topic string) {
	if *h.Verbose > 1 {
		h.Logger.Logf("Deleting topic: %s\n", topic)
	}

	if h.Topics[topic] == nil {
		if *h.Verbose > 1 {
			h.Logger.Logf("Topic not found!")
		}
		w.WriteHeader(http.StatusNotFound)
		return
	}

	delete(h.Topics, topic)
	w.WriteHeader(http.StatusOK)
}
