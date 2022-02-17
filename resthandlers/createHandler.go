package resthandlers

import (
  "log"
  "net/http"
)

type CreateHandler struct {}

func (c CreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  log.Println("[LOG] In CreateHandler")
}
