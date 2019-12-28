package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
	"sort"
)

// Backend defines the connection
type Backend struct {
	addr string
}

// BackendConnection will hold the state of each request
type BackendConnection struct {
	conn net.Conn
}

// InitBackend will attempt to connect to the backend
func InitBackend(addr string) *Backend {
	backend := &Backend{
		addr: addr,
	}

	return backend
}

const clientMagic = "HONK"
const serverMagic = "KNOH"

// Dial wraps net.Dial
func (b *Backend) Dial() (*BackendConnection, error) {
	conn, err := net.Dial("tcp", b.addr)
	if err != nil {
		return nil, err
	}

	bc := &BackendConnection{
		conn: conn,
	}

	err = bc.handshake()
	if err != nil {
		return nil, err
	}

	return bc, nil
}

func (bc *BackendConnection) handshake() error {
	_, err := bc.conn.Write([]byte(clientMagic))
	if err != nil {
		return err
	}

	magic := make([]byte, len(serverMagic))
	nread, err := bc.conn.Read(magic)
	if err != nil {
		return err
	}

	if nread != len(serverMagic) || serverMagic != string(magic) {
		return errors.New("Server and client version mismatch")
	}

	return nil
}

// Close wraps net.Close
func (bc *BackendConnection) Close() {
	bc.conn.Close()
}

// HealthCheck will print a log line on health check failure
func (b *Backend) HealthCheck() error {
	var healthCheckResponse interface{}
	err := b.RPC(0, []byte("{}"), &healthCheckResponse)
	if err != nil {
		return err
	}

	return nil
}

// RPC will rpc out to the backend
func (b *Backend) RPC(choice uint64, body []byte, v interface{}) error {
	conn, err := b.Dial()
	if err != nil {
		return err
	}

	defer conn.Close()

	choiceBuf := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(choiceBuf, choice)
	_, err = conn.conn.Write(choiceBuf)
	if err != nil {
		return err
	}

	bodyLenBuf := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(bodyLenBuf, uint64(len(body)))

	_, err = conn.conn.Write(bodyLenBuf)
	if err != nil {
		return err
	}

	_, err = conn.conn.Write(body)
	if err != nil {
		return err
	}

	respLenBuf := make([]byte, binary.MaxVarintLen64)
	_, err = conn.conn.Read(respLenBuf)
	if err != nil {
		return err
	}
	respLen, _ := binary.Uvarint(respLenBuf)
	respBuf := make([]byte, respLen)
	_, err = conn.conn.Read(respBuf)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBuf, v)

	return err
}

// Goose describes the state of each goose
type Goose struct {
	ID uint64 `json:"id"`

	Name  string `json:"name"`
	Honks uint64 `json:"honks"`
}

// ClientBody is every field we need to send the server
type ClientBody struct {
	ID uint64 `json:"id"`

	Name string `json:"name"`
}

// CreateGoose wraps backend CreateGoose
func (b *Backend) CreateGoose(name string) (*Goose, error) {
	body, err := json.Marshal(ClientBody{Name: name})
	if err != nil {
		return nil, err
	}

	var goose Goose
	err = b.RPC(1, body, &goose)

	return &goose, err
}

// EditGoose wraps backend EditGoose
func (b *Backend) EditGoose(ID uint64, name string) error {
	body, err := json.Marshal(ClientBody{ID: ID, Name: name})
	if err != nil {
		return err
	}

	var unused interface{}
	return b.RPC(3, body, &unused)
}

// GetGoose wraps backend GetGoose
func (b *Backend) GetGoose(ID uint64) (*Goose, error) {
	body, err := json.Marshal(ClientBody{ID: ID})
	if err != nil {
		return nil, err
	}

	var goose Goose
	err = b.RPC(2, body, &goose)

	return &goose, err
}

// Honk wraps backend Honk
func (b *Backend) Honk(ID uint64) error {
	body, err := json.Marshal(ClientBody{ID: ID})
	if err != nil {
		return err
	}

	var unused interface{}
	return b.RPC(5, body, &unused)
}

// DeleteGoose wraps DeleteGoose
func (b *Backend) DeleteGoose(ID uint64) error {
	body, err := json.Marshal(ClientBody{ID: ID})
	if err != nil {
		return err
	}

	var unused interface{}
	return b.RPC(4, body, &unused)
}

// GetGeese wraps GetGeese
func (b *Backend) GetGeese() ([]Goose, error) {
	body, err := json.Marshal(struct{}{})
	if err != nil {
		return nil, err
	}

	var geese []Goose
	b.RPC(6, body, &geese)

	sort.SliceStable(geese, func(i, j int) bool {
		return geese[i].ID < geese[j].ID
	})

	return geese, nil
}
