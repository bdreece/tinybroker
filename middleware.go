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
	"errors"
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

func (m *Middleware) verifyAuthRequest(body *map[string]interface{}) bool {
	var (
		user string
		pass string
		ok   bool
	)

	// Check if body JSON contains username field, pun to string
	if user, ok = (*body)[AUTH_USERNAME_FIELD].(string); !ok {
		if *m.Verbose > 1 {
			m.Logger.Logln("Request body missing 'username' field!")
		}
		return false
	}

	// Check if body JSON contains password field, pun to string
	if pass, ok = (*body)[AUTH_PASSWORD_FIELD].(string); !ok {
		if *m.Verbose > 1 {
			m.Logger.Logln("Request body missing 'password' field!")
		}
		return false
	}

	// Validate username and password against server
	if user != *m.User || pass != m.Pass {
		if *m.Verbose > 1 {
			m.Logger.Logln("Username or password incorrect!")
		}
		return false
	}
	return true
}

func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		topics []string
		ok     bool
	)

	body := ReadJSON(r, m.Logger)

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

	// Check for topics in request body
	if topics, ok = body[AUTH_TOPICS_FIELD].([]string); !ok {
		if *m.Verbose > 0 {
			m.Logger.Logln("No topics listed!")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Create JWT claims
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(*m.JWTTimeout)),
		Issuer:    *m.User,
		Audience:  topics,
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
		var tokenString string
		var token *jwt.Token

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
		token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			var (
				claims jwt.RegisteredClaims
				ok     bool
			)

			if _, ok = token.Method.(*jwt.SigningMethodHMAC); !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return nil, errors.New("invalid signing method")
			}

			if claims, ok = token.Claims.(jwt.RegisteredClaims); !ok {
				return nil, errors.New("invalid claims type")
			}

			// Validate authorized topics against requested topic
			if !claims.VerifyAudience(topic, true) {
				return nil, errors.New("not authorized for topic")
			}

			return []byte(*m.Secret), nil
		})

		if err != nil {
			if *m.Verbose > 0 {
				m.Logger.Logf("Error parsing token: %s\n", err.Error())
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Check token validity
		if !token.Valid {
			if *m.Verbose > 0 {
				m.Logger.Logln("Invalid token!")
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
