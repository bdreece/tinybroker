package resthandlers

import (
  "log"
  "net/http"
  redis "github.com/go-redis/redis/v8"
)

type DeleteHandler struct {
  rdb *redis.Client
}

func (c DeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  log.Println("[LOG] In DeleteHandler")
}

