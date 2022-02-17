package resthandlers

import (
  "log"
  "net/http"
)

type DeleteHandler struct {}

func (c DeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  log.Println("[LOG] In DeleteHandler")
}

