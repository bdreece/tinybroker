package main

import (
	"github.com/golang-collections/go-datastructures/queue"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var queueLen uint64 = 1024
var topicQueues map[string]*queue.RingBuffer

func readTopic(w http.ResponseWriter, r *http.Request) {
	log.Println("In readTopic()")

	topic := mux.Vars(r)["topic"]
	buffer, exists := topicQueues[topic]

	if !exists {
		log.Println("Topic not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	bytes, err := buffer.Get()
	if err != nil {
		log.Println("Error in fetching buffer")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(bytes.([]byte))
}

func createTopic(w http.ResponseWriter, r *http.Request) {
	log.Println("In createTopic()")

	topic := mux.Vars(r)["topic"]
	_, exists := topicQueues[topic]

	if exists {
		log.Println("Topic found")
		w.WriteHeader(http.StatusConflict)
		return
	}

	topicQueues[topic] = queue.NewRingBuffer(queueLen)

	if data, exists := r.PostForm["TB_DATA"]; exists {
		topicQueues[topic].Put([]byte(data[0]))
	} else {
		log.Println("No post data")
	}

	w.WriteHeader(http.StatusCreated)
}

func updateTopic(w http.ResponseWriter, r *http.Request) {
	log.Println("In updateTopic()")

	topic := mux.Vars(r)["topic"]
	buffer, exists := topicQueues[topic]
	if !exists {
		log.Println("Topic not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data, exists := r.PostForm["TB_DATA"]
	if !exists {
		log.Println("Error in fetching post form")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := buffer.Put([]byte(data[0]))
	if err != nil {
		log.Println("Error in pushing to buffer")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteTopic(w http.ResponseWriter, r *http.Request) {
	log.Println("In deleteTopic()")

	topic := mux.Vars(r)["topic"]
	buffer, exists := topicQueues[topic]

	if !exists {
		log.Println("Topic not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	buffer.Dispose()

	if !buffer.IsDisposed() {
		log.Println("Error in disposing of buffer")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	topicQueues = make(map[string]*queue.RingBuffer)

	router := mux.NewRouter()
	router.HandleFunc("/tb/{topic}", readTopic).Methods("GET")
	router.HandleFunc("/tb/{topic}", createTopic).Methods("POST")
	router.HandleFunc("/tb/{topic}", updateTopic).Methods("PUT")
	router.HandleFunc("/tb/{topic}", deleteTopic).Methods("DELETE")

	log.Println("Starting listening socket on port 8080")
	http.ListenAndServe(":8080", handlers.CompressHandler(router))
}
