package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

type Middleware struct {
	User    *string
	Pass    *string
	Secret  *string
	Verbose *int
}

func NewMiddleware(username, password, secret *string, verbose *int) Middleware {
	return Middleware{
		User:    username,
		Pass:    password,
		Secret:  secret,
		Verbose: verbose,
	}
}

func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if *m.Verbose > 1 {
		log.Printf("[LOG] User from %v attempting to login\n", r.UserAgent())
	}

	user := r.PostFormValue("TB_USER")
	pass := r.PostFormValue("TB_PASS")

	if user != *m.User || pass != *m.Pass {
		if *m.Verbose > 1 {
			log.Println("[LOG] Username or password incorrect!")
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if *m.Verbose > 1 {
		log.Printf("[LOG] User %s logged in!\n", user)
	}

	// Create JWT
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		Issuer:    user,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(*m.Secret))
	if err != nil {
		log.Printf("[ERR] Error creating JWT: %s\n", err.Error())
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
			if *m.Verbose > 1 {
				log.Printf("[ERR] Error parsing authorization header: %s\n", err.Error())
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Parse token
		token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return nil, errors.New("Invalid signing method!")
			}

			return []byte(*m.Secret), nil
		})

		if err != nil {
			if *m.Verbose > 0 {
				log.Printf("[ERR] %v\n", err.Error())
			}

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Check token validity
		if !token.Valid {
			if *m.Verbose > 0 {
				log.Println("[LOG] Invalid token!")
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if *m.Verbose > 0 {
			log.Println("[LOG] Successfully validated token!")
		}

		next.ServeHTTP(w, r)
	})
}
