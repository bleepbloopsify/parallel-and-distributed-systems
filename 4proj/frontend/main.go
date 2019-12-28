package main

import (
	"flag"
	"fmt"
)

func main() {
	fmt.Println("Hello World!")

	addr := flag.String("listen", ":8080", "The address for this server to listen on")

	backends := flag.String("backend", ":8081,:8082", "The other backends available")

	flag.Parse()

	server := CreateWebserver(*backends)
	server.ListenAndServe(*addr)
}
