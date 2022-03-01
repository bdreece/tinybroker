package main

import (
	"fmt"
	"github.com/bdreece/go-structs/ringbuf"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type Handler struct {
	Topics   map[string]*ringbuf.RingBuf
	Capacity *int
	Verbose  *int
}

func NewHandler(capacity *int, verbose *int) Handler {
	return Handler{
		Topics:   make(map[string]*ringbuf.RingBuf),
		Capacity: capacity,
		Verbose:  verbose,
	}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	topic := mux.Vars(r)["topic"]

	if *h.Verbose > 0 {
		log.Printf("[LOG] Serving request on topic: %s\n", topic)
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
		log.Printf("[ERR] Invalid request method: %s\n", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if *h.Verbose > 0 {
		log.Println("[LOG] Sent response!")
	}
}

func (h Handler) CreateResponse(w http.ResponseWriter, r *http.Request, topic string) {
	if *h.Verbose > 1 {
		log.Printf("[LOG] Creating topic: %s\n", topic)
	}

	data := r.PostFormValue("TB_DATA")
	if _, ok := h.Topics[topic]; !ok {
		newTopic := ringbuf.New(*h.Capacity)
		h.Topics[topic] = &newTopic
	}

	if len(data) > 0 {
		if *h.Verbose > 1 {
			log.Printf("[LOG] Create request contains data: %s\n", data)
		}

		h.Topics[topic].Write(data)
	}

	w.WriteHeader(http.StatusOK)
}

func (h Handler) ReadResponse(w http.ResponseWriter, r *http.Request, topic string) {
	if *h.Verbose > 1 {
		log.Printf("[LOG] Reading from topic: %s\n", topic)
	}

	if h.Topics[topic] == nil {
		if *h.Verbose > 1 {
			log.Printf("[LOG] Topic %s not found!", topic)
		}
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data := h.Topics[topic].Read()

	if data == nil {
		if *h.Verbose > 1 {
			log.Println("[LOG] Topic contains no data")
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if *h.Verbose > 1 {
		log.Printf("[LOG] Read data: %v\n", data)
	}

	_, err := w.Write([]byte(fmt.Sprint(data)))
	if err != nil {
		log.Printf("[ERR] %v\n", err.Error())
	}
}

func (h Handler) UpdateResponse(w http.ResponseWriter, r *http.Request, topic string) {
	if *h.Verbose > 1 {
		log.Printf("[LOG] Updating topic: %s\n", topic)
	}

	if h.Topics[topic] == nil {
		if *h.Verbose > 1 {
			log.Printf("[LOG] Topic %s not found!\n", topic)
		}
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data := r.PostFormValue("TB_DATA")
	if len(data) < 1 {
		if *h.Verbose > 1 {
			log.Println("[LOG] Data not found!")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if *h.Verbose > 1 {
		log.Printf("[LOG] Updating with data: %s\n", data)
	}

	h.Topics[topic].Write(data)
	w.WriteHeader(http.StatusOK)
}

func (h Handler) DeleteResponse(w http.ResponseWriter, r *http.Request, topic string) {
	if *h.Verbose > 1 {
		log.Printf("[LOG] Deleting topic: %s\n", topic)
	}

	if h.Topics[topic] == nil {
		if *h.Verbose > 1 {
			log.Println("[LOG] Topic not found!")
		}
		w.WriteHeader(http.StatusNotFound)
		return
	}

	delete(h.Topics, topic)
	w.WriteHeader(http.StatusOK)
}
