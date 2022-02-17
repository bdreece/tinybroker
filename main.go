package main

import (
    "log"
	"net/http"
    "time"
	"github.com/gorilla/mux"
    rest "github.com/bdreece/tinybroker/resthandlers"
)

func methodRouter(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "POST":
      rest.CreateHandler{}.ServeHTTP(w, r)
    case "GET":
      rest.ReadHandler{}.ServeHTTP(w, r)
    case "PUT":
      rest.UpdateHandler{}.ServeHTTP(w, r)
    case "DELETE":
      rest.DeleteHandler{}.ServeHTTP(w, r)
    default: log.Printf("[ERR] Invalid request method: %s\n", r.Method)
  }
}

func main() {
  router := mux.NewRouter()

  router.HandleFunc("/tb/{topic}", methodRouter).
         Methods("POST", "GET", "PUT", "DELETE")

  srv := &http.Server{
    Handler:    router,
    Addr:       "127.0.0.1:8080",
    WriteTimeout: 15 * time.Second,
    ReadTimeout: 15 * time.Second,
  }

  log.Fatal(srv.ListenAndServe())
}
