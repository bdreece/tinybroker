package main

import (
	"io"
	"net/http"

	"github.com/bdreece/tattle"
)

func ReadBody(r *http.Request, l *tattle.Logger) ([]byte, int) {
	// Read request body
	var (
		data []byte
		total int
	)	
	buf := make([]byte, 1024)
	for {
		n, err := r.Body.Read(buf)

		if err != nil && err != io.EOF {
			l.Errf("Error reading request body: %s\n", err.Error())
			continue
		}
		
		if n > 0 {
			data = append(data, buf[:n]...)
			total += n
		}
		
		if err == io.EOF {
			break
		}
	}

	return data, total
}