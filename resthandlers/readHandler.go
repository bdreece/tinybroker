
package resthandlers

import (
  "log"
  "net/http"
  redis "github.com/go-redis/redis/v8"
)

type ReadHandler struct {
  rdb *redis.Client
}

func (c ReadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  log.Println("[LOG] In ReadHandler")
}
