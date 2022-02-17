
package resthandlers

import (
  "log"
  "net/http"
)

type ReadHandler struct {}

func (c ReadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  log.Println("[LOG] In ReadHandler")
}
