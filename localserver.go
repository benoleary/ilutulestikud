package main

import (
	"fmt"
	"github.com/benoleary/ilutulestikud/server"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("This program needs exactly one argument, which is the directory to serve for the Angular code")
		os.Exit(1)
	}

	angularDirectory := os.Args[1]
	httpFileServer := http.FileServer(http.Dir(angularDirectory))
	http.Handle("/client/", http.StripPrefix("/client", httpFileServer))

	// We could load the allowed origin from a file, but this app is very specific to a set of fixed addresses.
	serverState := server.CreateNew("http://localhost:4233")
	http.HandleFunc("/backend/", serverState.HandleBackend)
	http.ListenAndServe(":8080", nil)
}
