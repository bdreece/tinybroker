package middleware

import "net/http"

type Middleware struct {
  Secret    string
}

func New(secret string) Middleware {
  return Middleware{
    Secret: secret,
  }
}

func (m Middleware) authMiddleware(next http.Handler) http.Handler {
  return nil 
}
