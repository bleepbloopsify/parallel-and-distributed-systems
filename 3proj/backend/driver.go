package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"log"
	"net"
)

// Connection will hold all of the state of the current connection
type Connection struct {
	conn         net.Conn
	gooseStorage *GooseStorage // reference to data store
}

// InitConnection will store the state of communications (not thread-safe)
func (server *Server) InitConnection(conn net.Conn) *Connection {
	return &Connection{conn: conn, gooseStorage: server.gooseStorage}
}

const clientMagic = "HONK"
const serverMagic = "KNOH"

// Handshake will initialize the conversation and check protocol versions
func (conn *Connection) handshake() error {
	magic := make([]byte, len(clientMagic))

	nread, err := conn.conn.Read(magic)
	if err != nil {
		return err
	}

	if nread != len(clientMagic) || string(magic) != clientMagic {
		return errors.New("Client is not a gooseStorage client")
	}

	conn.conn.Write([]byte(serverMagic))

	return nil
}

func (conn *Connection) parseRequest() (uint64, []byte, error) {
	choiceBuf := make([]byte, binary.MaxVarintLen64)
	_, err := conn.conn.Read(choiceBuf)
	if err != nil {
		return 0, nil, err
	}

	choice, nread := binary.Uvarint(choiceBuf) // This is the choice field
	if nread <= 0 {
		return 0, nil, errors.New("Bad parsing integer")
	}

	bodyLengthBuf := make([]byte, binary.MaxVarintLen64)
	_, err = conn.conn.Read(bodyLengthBuf)
	if err != nil {
		return 0, nil, err
	}

	bodyLength, _ := binary.Uvarint(bodyLengthBuf) // This is the size of the body

	body := make([]byte, bodyLength)
	_, err = conn.conn.Read(body)
	if err != nil {
		return 0, nil, err
	}

	return choice, body, nil
}

// Close wraps net.Close
func (conn *Connection) Close() {
	conn.conn.Close()
}

func (conn *Connection) sendResponse(msg []byte) error {
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(buf, uint64(len(msg)))
	_, err := conn.conn.Write(buf)
	if err != nil {
		return err
	}
	_, err = conn.conn.Write(msg)

	return err
}

// ClientBody is the json format we expect client requests to come in.
type clientBody struct {
	ID uint64 `json:"id"`

	Name string `json:"name"`
}

// Handle will take care of RPC across the connection
func (conn *Connection) Handle() error {
	defer conn.Close()

	err := conn.handshake()
	if err != nil {
		return err
	}

	choice, body, err := conn.parseRequest()
	if err != nil {
		return err
	}
	var response []byte
	var clientBody clientBody
	err = json.Unmarshal(body, &clientBody)
	if err != nil {
		log.Printf("%v", body)
		return err
	}

	switch choice {
	case 0: // HealthCheck
		log.Printf("[%v] HealthCheck", conn.conn.RemoteAddr())
		response = []byte("{ \"healthy\": true }")
	case 1: // Create
		log.Printf("[%v] Create", conn.conn.RemoteAddr())
		goose := conn.gooseStorage.CreateGoose(clientBody.Name)
		response, err = json.Marshal(goose)
		if err != nil {
			return err
		}
	case 2: // Read
		log.Printf("[%v] Read", conn.conn.RemoteAddr())
		goose, err := conn.gooseStorage.GetGoose(clientBody.ID)
		if err != nil {
			return err
		}
		response, err = json.Marshal(goose)
		if err != nil {
			return err
		}
	case 3: // Update
		log.Printf("[%v] Update", conn.conn.RemoteAddr())
		err := conn.gooseStorage.EditGoose(clientBody.ID, clientBody.Name)
		if err != nil {
			return err
		}
		response = []byte("{}")
	case 4: // Delete
		log.Printf("[%v] Delete", conn.conn.RemoteAddr())
		err := conn.gooseStorage.DeleteGoose(clientBody.ID)
		if err != nil {
			return err
		}
		response = []byte("{}")
	case 5: // Honk
		log.Printf("[%v] Honk", conn.conn.RemoteAddr())
		err = conn.gooseStorage.Honk(clientBody.ID)
		if err != nil {
			return err
		}
		response = []byte("{}")
	case 6: // Listing
		log.Printf("[%v] Listing", conn.conn.RemoteAddr())
		geese := conn.gooseStorage.GetGeese()
		response, err = json.Marshal(geese)
	default:
		log.Printf("[%v] Bad Choice", conn.conn.RemoteAddr())
		return errors.New("Nonexistent choice made by the client")
	}

	return conn.sendResponse(response)
}
