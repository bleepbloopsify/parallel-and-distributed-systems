package main

import (
	"log"
	"net"
)

// Server holds the information and state of the server
type Server struct {
	addr         string // address to listen on
	gooseStorage *GooseStorage
}

// InitServer will initialize everything for the server
func InitServer(addr string) *Server {
	server := &Server{
		addr:         addr,
		gooseStorage: InitGooseStorage(),
	}

	return server
}

// Listen wrap's http/net's Listen method
// It also initializes goroutines for each connection
func (server *Server) Listen() error {
	ln, err := net.Listen("tcp", server.addr)
	if err != nil {
		return err
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting connection %v\n", err)
			continue
		}

		go server.handle(conn)
	}
}

func (server *Server) handle(conn net.Conn) {
	c := server.InitConnection(conn)
	err := c.Handle()
	if err != nil {
		log.Printf("[%v] error: %v", conn.RemoteAddr(), err)
	}
}
