package resthandlers

import (
  "log"
  "net/http"
  redis "github.com/go-redis/redis/v8"
)

type UpdateHandler struct {
  rdb *redis.Client
}

func (c UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  log.Println("[LOG] In UpdateHandler")
}
