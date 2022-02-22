package main

import "net/http"

type Middleware struct {
  Secret    string
}

func NewMiddleware(secret string) Middleware {
  return Middleware{
    Secret: secret,
  }
}

func (m Middleware) authMiddleware(next http.Handler) http.Handler {
  // TODO
  return nil 
}
