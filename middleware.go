/* tinybroker - A simple message broker, written in Go
Copyright (C) 2022 Brian Reece

This program is free software; you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation; either version 2 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License along
with this program; if not, write to the Free Software Foundation, Inc.,
51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/

package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bdreece/tattle"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

const (
	AUTH_USERNAME_FIELD string = "username"
	AUTH_PASSWORD_FIELD string = "password"
	AUTH_TOPICS_FIELD string = "topics"
)

type Middleware struct {
	User       *string
	Pass       string
	Secret     *string
	JWTTimeout *time.Duration
	Verbose    *int
	Logger 	   *tattle.Logger
}

type AuthPacket struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Topics []string `json:"topics"`
}

func NewMiddleware(username, password, secret *string, jwtTimeout *time.Duration, verbose *int, logger *tattle.Logger) Middleware {
	h := sha256.New()
	h.Write([]byte(*password))
	hashPW := h.Sum(nil)

	return Middleware{
		User:       username,
		Pass:       fmt.Sprintf("%x", hashPW),
		Secret:     secret,
		JWTTimeout: jwtTimeout,
		Verbose:    verbose,
		Logger: 	logger,
	}
}

func (m *Middleware) verifyAuthRequest(body *AuthPacket) bool {
	// Validate username and password against server
	if body.Username != *m.User || body.Password != m.Pass {
		if *m.Verbose > 1 {
			m.Logger.Logln("Username or password incorrect!")
		}
		return false
	}
	return true
}

func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body AuthPacket

	if *m.Verbose > 0 {
		m.Logger.Logln("User attempting to authenticate")
	}

	// Read and unmarshal body to AuthPacket
	data, n := ReadBody(r, m.Logger)
	if err := json.Unmarshal(data[:n], &body); err != nil {
		m.Logger.Errf("Error unmarshaling request body: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Verify authentication info
	if !m.verifyAuthRequest(&body) {
		if *m.Verbose > 0 {
			m.Logger.Logln("User failed to log in")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if *m.Verbose > 0 {
		m.Logger.Logf("User %s logged in!\n", *m.User)
	}

	// Create JWT claims
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(*m.JWTTimeout)),
		Issuer:    *m.User,
		Audience:  body.Topics,
	}

	// Create JWT and sign
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(*m.Secret))
	if err != nil {
		m.Logger.Errf("Error creating JWT: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte(ss))
}

func (m Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			tokenString string
			token *jwt.Token
		)

		// Parse authorization header
		_, err := fmt.Sscanf(r.Header.Get("Authorization"), "Bearer %v", &tokenString)
		if err != nil {
			if *m.Verbose > 1 {
				m.Logger.Errf("Error parsing authorization header: %s\n", err.Error())
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Get requested topic
		topic := mux.Vars(r)["topic"]

		// Parse JWT
		token, err = jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(*m.Secret), nil
		})
		if err != nil {
			if *m.Verbose > 0 {
				m.Logger.Logf("Error parsing token: %s\n", err.Error())
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*jwt.RegisteredClaims)

		// Check token validity
		if !(ok && token.Valid) {
			if *m.Verbose > 0 {
				m.Logger.Logln("Invalid token!")
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Check authorized topics against requested topic
		if !claims.VerifyAudience(topic, true) {
			if *m.Verbose > 0 {
				m.Logger.Logf("User not authorized on topic: %s\n", topic)
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if *m.Verbose > 0 {
			m.Logger.Logln("Successfully validated token!")
		}
		next.ServeHTTP(w, r)
	})
}
