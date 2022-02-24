package main

import (
  "errors"
  "fmt"
  "log"
  "net/http"
  "os"
  "time"

  jwt "github.com/golang-jwt/jwt/v4"
)


type Middleware struct {
  Secret    string
  Verbose   bool
}

func NewMiddleware(secret string, verbose bool) Middleware {
  return Middleware{
    Secret: secret,
    Verbose: verbose,
  }
}

func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  var (
    TB_USER string = os.Getenv("TB_USER")
    TB_PASS        = os.Getenv("TB_PASS")
  )

  if m.Verbose {
    log.Println("Attempting to login")
  }

  user := r.PostFormValue("TB_USER")
  pass := r.PostFormValue("TB_PASS")

  if user != TB_USER || pass != TB_PASS {
    if m.Verbose {
      log.Println("Username or password incorrect!")
    }
    w.WriteHeader(http.StatusUnauthorized)
    return
  }

  if m.Verbose {
    log.Printf("User %s logged in!\n", user)
  }

  // Create JWT
  claims := & jwt.RegisteredClaims{
    ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
    Issuer:    user,
  }

  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  ss, err := token.SignedString([]byte(m.Secret))
  if err != nil {
    if m.Verbose {
      log.Printf("Error creating JWT: %s\n", err.Error())
    }
    w.WriteHeader(http.StatusInternalServerError)
  }

  w.Write([]byte(ss))
}

func (m Middleware) AuthMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    var tokenString string
    var token *jwt.Token

    // Parse authorization header
    _, err := fmt.Sscanf(r.Header.Get("Authorization"), "Bearer %v", &tokenString)
    if err != nil {
      if m.Verbose {
        log.Printf("Error parsing authorization header: %s\n", err.Error())
      }
      w.WriteHeader(http.StatusInternalServerError)
      return
    }

    // Parse token
    token, err = jwt.Parse(tokenString, func (token *jwt.Token) (interface{}, error) {
      if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        if m.Verbose {
          log.Println("Invalid signing method!")
        }
        w.WriteHeader(http.StatusUnauthorized)
        return nil, errors.New("Invalid signing method!")
      }

      return []byte(m.Secret), nil
    })

    if err != nil {
      if m.Verbose {
        log.Printf("Error parsing token: %s\n", err.Error())
      }

      w.WriteHeader(http.StatusInternalServerError)
      return
    }

    // Check token validity
    if !token.Valid {
      if m.Verbose {
        log.Println("Invalid token!")
      }
      w.WriteHeader(http.StatusUnauthorized)
      return
    }

    if m.Verbose {
      log.Println("Successfully validated token!")
    }

    next.ServeHTTP(w, r)
  })
}
