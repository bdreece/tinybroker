package handler

import (
  "log"
  "net/http"
  redis "github.com/go-redis/redis/v8"
)

type Handler struct {
  Rdb *redis.Client
}

func New(rdb *redis.Client) Handler {
  return Handler {
    Rdb: rdb
  }
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "POST":
      h.ServeCreateResponse(w, r)
    case "GET":
      h.ServeReadResponse(w, r)
    case "PUT":
      h.ServeUpdateResponse(w, r)
    case "DELETE":
      h.ServeDeleteResponse(w, r)
    default: log.Printf("[ERR] Invalid request method: %s\n", r.Method)
  }
}

func (h Handler) ServeCreateResponse(w http.ResponseWriter, r *http.Request) {

}

func (h Handler) ServeReadResponse(w http.ResponseWriter, r *http.Request) {

}

func (h Handler) ServeUpdateResponse(w http.ResponseWriter, r *http.Request) {

}

func (h Handler) ServeDeleteResponse(w http.ResponseWriter, r *http.Request) {

}
