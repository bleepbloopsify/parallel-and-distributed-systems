package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	listenAddr := flag.String("listen", ":8081", "the address to listen on")

	flag.Parse()

	fmt.Println("Listening on port 8081...")
	server := InitServer(*listenAddr)
	err := server.Listen()
	if err != nil {
		log.Fatal(err)
	}
}
