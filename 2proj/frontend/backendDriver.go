package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
)

// ClientMagic is the greeting we send to the server
const ClientMagic = "DOGE"

// ServerMagic is the response we expect from the server
const ServerMagic = "EGOD"

// Conn holds the backendDrivers stuff
type Conn struct {
	conn net.Conn
	body []byte // We use byte array json data to hold our bodies
}

// Dial replicates TCP dial, but ensures the protocol.
func Dial(addr string) (*Conn, error) {
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		return nil, err
	}

	// We only close the connection if there are issues reading the protocol
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()
	conn.Write([]byte(ClientMagic))

	magic := make([]byte, len(ServerMagic))
	nread, err := conn.Read(magic)
	if err != nil {
		return nil, err
	}

	if nread != len(ServerMagic) || string(magic) != ServerMagic {
		return nil, errors.New("Server is not using the correct protocol")
	}

	return &Conn{
		conn: conn,
	}, nil
}

// Close closes the connection and cleans up
func (conn *Conn) Close() {
	conn.conn.Close()
}

func (conn *Conn) writeFuncID(funcID int64) error {
	buf := make([]byte, 8) // int64 is 8 bytes long
	binary.PutUvarint(buf, uint64(funcID))
	_, err := conn.conn.Write(buf)
	if err != nil {
		return err
	}

	return nil
}

func (conn *Conn) addBody(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	conn.body = data

	return nil
}

// SendReq wraps the connection using json interfaces
func (conn *Conn) SendReq(obj interface{}) error {
	buf := make([]byte, 8)
	binary.PutUvarint(buf, uint64(len(conn.body)))

	conn.conn.Write(buf)
	conn.conn.Write(conn.body)

	n, err := conn.conn.Read(buf)
	if err != nil {
		return err
	}

	if n != 8 {
		return errors.New("Bad size field")
	}

	sizeField, nbytesread := binary.Uvarint(buf)
	if nbytesread == 0 {
		return errors.New("Parsing size field failed")
	}

	buf = make([]byte, sizeField)
	nRead, err := io.ReadFull(conn.conn, buf)
	if err != nil {
		return err
	}

	if uint64(nRead) != sizeField {
		return errors.New("Miswritten size")
	}

	err = json.Unmarshal(buf, obj)
	if err != nil {
		return err
	}

	return err
}

// Dog is what we are CRUD'ing
type Dog struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	HasToy      bool   `json:"has_toy"`
	TimesPet    int64  `json:"times_pet"`
	WantsToPlay bool   `json:"wants_to_play"`
}

type createDogPayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateDog will negotiate dog creation on the backend
func (conn *Conn) CreateDog(name string, description string) (*Dog, error) {

	const funcID = 1 // This is function 1

	err := conn.writeFuncID(funcID)
	if err != nil {
		return nil, err
	}

	err = conn.addBody(createDogPayload{Name: name, Description: description})
	if err != nil {
		return nil, err
	}

	var dog Dog
	err = conn.SendReq(&dog)
	if err != nil {
		return nil, err
	}

	return &dog, nil
}

type idPayload struct {
	ID int32 `json:"id"`
}

// ReadDog returns a dog from the backend
func (conn *Conn) ReadDog(ID int32) (*Dog, error) {
	const funcID = 2
	err := conn.writeFuncID(funcID)
	if err != nil {
		return nil, err
	}
	err = conn.addBody(idPayload{ID: ID})
	if err != nil {
		return nil, err
	}

	var dog Dog
	err = conn.SendReq(&dog)
	if err != nil {
		return nil, err
	}

	return &dog, nil
}

// UpdateDog takes a dog object and updates the server using the id
// inside the struct
func (conn *Conn) UpdateDog(dog *Dog) (*Dog, error) {
	const funcID = 3
	err := conn.writeFuncID(funcID)
	if err != nil {
		return nil, err
	}

	err = conn.addBody(dog)
	if err != nil {
		return nil, err
	}

	var updatedDog Dog
	err = conn.SendReq(&updatedDog)
	if err != nil {
		return nil, err
	}

	return &updatedDog, nil
}

// DeleteDog deletes itself from the server
func (conn *Conn) DeleteDog(ID int32) (interface{}, error) {
	const funcID = 4
	err := conn.writeFuncID(funcID)
	if err != nil {
		return nil, err
	}

	err = conn.addBody(idPayload{ID: ID})
	if err != nil {
		return nil, err
	}

	var obj interface{}
	err = conn.SendReq(&obj)
	if err != nil {
		return nil, err
	}

	return &obj, nil
}

// ListDogs retrieves the list of dogs from the server
func (conn *Conn) ListDogs() ([]*Dog, error) {
	const funcID = 5
	err := conn.writeFuncID(funcID)
	if err != nil {
		return nil, err
	}

	err = conn.addBody(struct{}{}) // empty body
	if err != nil {
		return nil, err
	}

	var dogs []*Dog
	err = conn.SendReq(&dogs)
	if err != nil {
		return nil, err
	}

	return dogs, nil
}
