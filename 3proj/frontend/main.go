package main

import (
	"flag"
	"log"
)

func main() {
	listenAddr := flag.String("listen", ":8080", "The server for the frontend to listen on")
	backendAddr := flag.String("backend", ":8081", "The address for the backend")

	flag.Parse()

	server, err := InitWebserver(*backendAddr)
	if err != nil {
		log.Fatal("Could not initialize webserver")
	}

	server.Run(*listenAddr)
}
