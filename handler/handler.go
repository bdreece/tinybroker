package handler

import (
  "log"
  "net/http"
  "github.com/gorilla/mux"
)

type Handler struct {
  Topics    map[string] chan string
  Capacity  int
  Verbose bool
}

func NewHandler(capacity int, verbose bool) Handler {
  return Handler {
    Topics: make(map[string]chan string),
    Capacity: capacity,
    Verbose: verbose,
  }
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  topic := mux.Vars(r)["topic"]

  if h.Verbose {
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

  if h.Verbose {
    log.Println("[LOG] Sent response!")
  }
}

func (h Handler) CreateResponse(w http.ResponseWriter, r *http.Request, topic string) {
  if h.Verbose {
    log.Printf("[LOG] Creating topic: %s\n", topic)
  }

  data := r.PostFormValue("TB_DATA")
  if h.Topics[topic] == nil {
    h.Topics[topic] = make(chan string, h.Capacity)
  }

  if len(data) > 0 {
    if h.Verbose {
      log.Printf("[LOG] Create request contains data: %s\n", data)
    }

    h.Topics[topic]<- data
  }

  w.WriteHeader(http.StatusOK)
}

func (h Handler) ReadResponse(w http.ResponseWriter, r *http.Request, topic string) {
  if h.Verbose {
    log.Printf("[LOG] Reading from topic: %s\n", topic)
  }

  if h.Topics[topic] == nil {
    if h.Verbose {
      log.Printf("[LOG] Topic %s not found!", topic)
    }
    w.WriteHeader(http.StatusNotFound)
    return
  }

  data := <-h.Topics[topic]

  if h.Verbose {
    log.Printf("[LOG] Read data: %s\n", data)
  }

  _, err := w.Write([]byte(data))
  if err != nil {
    log.Printf("[ERR] %v\n", err.Error())
  }
}

func (h Handler) UpdateResponse(w http.ResponseWriter, r *http.Request, topic string) {
  if h.Verbose {
    log.Printf("[LOG] Updating topic: %s\n", topic)
  }

  if h.Topics[topic] == nil {
    if h.Verbose {
      log.Printf("[LOG] Topic %s not found!\n", topic)
    }
    w.WriteHeader(http.StatusNotFound)
    return
  }
  
  data := r.PostFormValue("TB_DATA")
  if len(data) < 1 {
    if h.Verbose {
      log.Println("[LOG] Data not found!")
    }
    w.WriteHeader(http.StatusBadRequest)
  }

  if h.Verbose {
    log.Printf("[LOG] Updating with data: %s\n", data)
  }

  h.Topics[topic]<- data
  w.WriteHeader(http.StatusOK)
}

func (h Handler) DeleteResponse(w http.ResponseWriter, r *http.Request, topic string) {
  if h.Verbose {
    log.Printf("[LOG] Deleting topic: %s\n", topic)
  }

  if h.Topics[topic] == nil {
    if h.Verbose {
      log.Println("[LOG] Topic not found!")
    }
    w.WriteHeader(http.StatusNotFound)
    return
  }

  delete(h.Topics, topic)
  w.WriteHeader(http.StatusOK)
}
