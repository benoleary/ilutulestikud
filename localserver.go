package main

import (
	"fmt"
	"github.com/benoleary/ilutulestikud/server"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("This program needs exactly one argument, which is the directory to serve for the Angular code")
		os.Exit(1)
	}

	angularDirectory := os.Args[1]
	server.Serve(angularDirectory)
}
