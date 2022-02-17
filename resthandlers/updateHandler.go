package resthandlers

import (
  "log"
  "net/http"
)

type UpdateHandler struct {}

func (c UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  log.Println("[LOG] In UpdateHandler")
}
