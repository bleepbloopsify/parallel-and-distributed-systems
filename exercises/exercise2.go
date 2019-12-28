package main

import (
	"fmt"
	"net"
)

func main() {
	service := "localhost:8083"

	conn, err := net.Dial("tcp", service)
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	buf := make([]byte, 0xff)
	nRead, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Conn could not read")
	}

	fmt.Printf("read %d as %s\n", nRead, buf)
}
