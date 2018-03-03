package main

import (
	"fmt"
	"net/http"

	"github.com/benoleary/ilutulestikud/backend/server"
)

func main() {
	fmt.Printf("Local server started.\n")

	// We could load the allowed origin from a file, but this app is very specific to a set of fixed addresses.
	serverState := server.NewWithDefaultHandlers("http://localhost:4233")
	http.HandleFunc("/backend/", serverState.HandleBackend)
	http.ListenAndServe(":8080", nil)
}
