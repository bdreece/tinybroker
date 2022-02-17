package resthandlers

import (
  "log"
  "net/http"
  redis "github.com/go-redis/redis/v8"
)

type CreateHandler struct {
  rdb *redis.Client
}

func (c CreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  log.Println("[LOG] In CreateHandler")
}
