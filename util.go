package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/bdreece/tattle"
)

func ReadBody(r *http.Request, l *tattle.Logger) []byte {
	// Read request body
	var data []byte
	buf := make([]byte, 1024)
	for {
		n, err := r.Body.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			l.Errf("Error reading request body: %s\n", err.Error())
			continue
		}
		if n > 0 {
			data = append(data, buf...)
		}
	}

	return data
}

func ReadJSON(r *http.Request, l *tattle.Logger) map[string]any {
	body := make(map[string]any)
	// Unmarshal data
	data := ReadBody(r, l)
	if err := json.Unmarshal(data, &body); err != nil {
		l.Errf("Error unmarshaling request body: %s\n", err.Error())
		return nil
	}

	return body
}