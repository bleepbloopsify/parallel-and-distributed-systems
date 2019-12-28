package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"net"
)

// ClientMagic is the greeting we expect from the client
const ClientMagic = "DOGE"

// ServerMagic is the response we send to the client
const ServerMagic = "EGOD"

func main() {
	addr := flag.String("listen", ":8090", "the address for the backend service to listen on")
	flag.Parse()

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Listening...")

	store := InitStorage()
	store.SeedStorage()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("Handling connection from %v\n", conn.RemoteAddr())

		err = store.connHandler(conn)
		if err != nil {
			log.Print(err)
		}
	}
}

func readUint(conn net.Conn) (uint64, error) {
	buf := make([]byte, 8)
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		return 0, err
	}

	n, nerr := binary.Uvarint(buf)
	if nerr <= 0 {
		return 0, errors.New("Error parsing number")
	}

	return n, nil
}

func (store *DogStorage) connHandler(conn net.Conn) error {
	defer func() {
		conn.Close() // we handle closing here
	}()

	magic := make([]byte, len(ClientMagic))

	nread, err := conn.Read(magic)
	if err != nil {
		return err
	}

	if nread != len(ClientMagic) || string(magic) != ClientMagic {
		return errors.New("Client is on the wrong protocol")
	}

	nwrit, err := conn.Write([]byte(ServerMagic))
	if err != nil {
		return errors.New("Could not finish writing serverMagic")
	}

	if nwrit != len(ServerMagic) {
		return errors.New("Could not write entire serverMagic")
	}
	// Connection initialized. Read type of message now

	choice, err := readUint(conn)
	if err != nil {
		return err
	}

	bodySize, err := readUint(conn)
	if err != nil {
		return err
	}

	buf := make([]byte, bodySize)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return err
	}

	var recvBody interface{}
	err = json.Unmarshal(buf, &recvBody)
	if err != nil {
		return err
	}

	var response interface{}
	switch choice {
	case 1:
		response, err = store.CreateDog(recvBody)
	case 2:
		response, err = store.ReadDog(recvBody)
	case 3:
		response, err = store.UpdateDog(recvBody)
	case 4:
		response, err = store.DeleteDog(recvBody)
	case 5:
		response, err = store.ListDogs(recvBody)
	default:
		response = struct {
			Health string `json:"health"`
		}{Health: "healthy"}
	}

	if err != nil {
		return err
	}

	body, err := json.Marshal(response)
	if err != nil {
		return err
	}

	buf = make([]byte, 8)
	binary.PutUvarint(buf, uint64(len(body)))

	nsent, err := conn.Write(buf)
	if err != nil {
		return err
	}

	if nsent != 8 {
		return errors.New("Could not write entire size")
	}

	_, err = conn.Write(body)
	if err != nil {
		return err
	}

	return nil
}

// TODO: func read size
// TODO: func read json body
// TODO: func write json body
