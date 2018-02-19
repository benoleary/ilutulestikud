package main

import (
	"github.com/benoleary/ilutulestikud/server"
	"net/http"
)

func main() {
	// We could load the allowed origin from a file, but this app is very specific to a set of fixed addresses.
	serverState := server.New("http://localhost:4233")
	http.HandleFunc("/backend/", serverState.HandleBackend)
	http.ListenAndServe(":8080", nil)
}
