package main

import (
	"fmt"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":8083")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue // ignore this connection if accepting errors out
		}
		defer conn.Close()

		fmt.Printf("Accepted connection from %s\n", conn.RemoteAddr())

		conn.Write([]byte("Hello"))
	}
}
