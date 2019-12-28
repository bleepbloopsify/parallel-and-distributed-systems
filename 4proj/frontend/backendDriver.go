package main

import (
	"errors"
	"log"
	"net/rpc"
	"strings"

	"4proj/frontend/protos"
)

// Node represents a single node and the client we use to communicate with them
type Node struct {
	Addr  string
	Alive bool

	Client *rpc.Client

	lock chan bool
}

// Backend describes the whole cluster and our primary point of contact
type Backend struct {
	nodes   map[string]*Node
	primary *Node // We cache the primary point of contact

	lock chan bool
}

// CreateNode is the constructor for node
func CreateNode(addr string) *Node {
	return &Node{
		Addr:   addr,
		Client: nil,

		lock: make(chan bool, 1),
	}
}

func (node *Node) connect() error {
	node.lock <- true
	defer func() {
		<-node.lock
	}()

	if node.Client != nil { // we don't need to connect if we've already connected
		return nil
	}

	client, err := rpc.DialHTTP("tcp", node.Addr)
	if err != nil {
		return err
	}

	var fit bool
	err = client.Call("Backend.Healthcheck", 0, &fit)
	if err != nil || !fit {
		client.Close()
		return errors.New("Node is not okay")
	}

	node.Client = client
	return nil
}

func (node *Node) close() {
	node.lock <- true
	defer func() {
		<-node.lock
	}()

	if node.Client == nil {
		return
	}

	node.Client.Close()
	node.Client = nil
}

func parseBackends(backends string) map[string]*Node {
	nodes := map[string]*Node{}

	for _, addr := range strings.FieldsFunc(backends, func(c rune) bool {
		return c == ','
	}) {
		nodes[addr] = CreateNode(addr)
	}

	return nodes
}

// CreateBackend is a constructor for the backend
func CreateBackend(backends string) *Backend {
	return &Backend{
		nodes: parseBackends(backends),
		lock:  make(chan bool, 1),
	}
}

func (backend *Backend) connectNode(node *Node) {
	err := node.connect()
	if err != nil {
		log.Print(err)
		return
	}

	backend.lock <- true
	defer func() {
		<-backend.lock
	}()

	if backend.primary == nil {
		backend.primary = node
	}
}

func (backend *Backend) selectPrimary() error {
	backend.lock <- true
	defer func() {
		<-backend.lock
	}()

	if backend.primary != nil {
		return nil
	}
	for _, node := range backend.nodes {
		if node.Client != nil { // We just grab the first node we see
			backend.primary = node
			return nil
		}

		go backend.connectNode(node)
	}

	return errors.New("No node is fit to be primary yet")
}

// CreateGiraffe is an rpc exposed method to create a giraffe
func (backend *Backend) CreateGiraffe(name string) (*protos.Giraffe, error) {
	err := backend.selectPrimary() // just to make sure
	if err != nil {
		return nil, err
	}

	err = backend.primary.connect()
	if err != nil {
		backend.primary = nil
		return nil, err
	}

	backend.primary.lock <- true
	defer func() {
		<-backend.primary.lock
	}()

	if backend.primary.Client == nil {
		return nil, errors.New("Client died mid-connection")
	}

	log.Printf("Rpc to %v\n", backend.primary.Addr)

	var reply protos.Giraffe
	err = backend.primary.Client.Call("Backend.CreateGiraffe", &name, &reply)
	if err != nil {
		<-backend.primary.lock
		backend.primary.close()
		backend.primary.lock <- true
		backend.primary = nil
		return nil, err
	}

	return &reply, nil
}

// ReadGiraffe is an RPC exposed method to read a giraffe
func (backend *Backend) ReadGiraffe(idx uint64) (*protos.Giraffe, error) {
	err := backend.selectPrimary() // just to make sure
	if err != nil {
		backend.primary = nil
		return nil, err
	}

	var giraffe protos.Giraffe
	err = backend.primary.Client.Call("Backend.ReadGiraffe", &idx, &giraffe)
	if err != nil {
		backend.primary = nil
		return nil, err
	}

	return &giraffe, nil
}

// LogEditGiraffeArgs will keep all information necessary to edit a giraffe later
type LogEditGiraffeArgs struct {
	Idx        uint64
	Name       string
	NeckLength uint64
}

// UpdateGiraffe is an RPC exposed method to update an entry
func (backend *Backend) UpdateGiraffe(args *LogEditGiraffeArgs) error {
	err := backend.selectPrimary() // just to make sure
	if err != nil {
		backend.primary = nil
		return err
	}

	var reply protos.Giraffe
	err = backend.primary.Client.Call("Backend.EditGiraffe", args, &reply)
	if err != nil {
		return err
	}

	return nil
}

// DeleteGiraffe is an RPC exposed method to delete an entry
func (backend *Backend) DeleteGiraffe(idx uint64) error {
	err := backend.selectPrimary() // just to make sure
	if err != nil {
		backend.primary = nil
		return err
	}

	var success bool
	err = backend.primary.Client.Call("Backend.DeleteGiraffe", &idx, &success)
	if err != nil {
		return err
	}

	if !success {
		return errors.New("error deleting giraffe")
	}

	return nil
}

// ListEntries will grab all of the entries for a particular store
func (backend *Backend) ListEntries() ([]protos.Giraffe, error) {
	err := backend.selectPrimary()
	if err != nil {
		backend.primary = nil
		return nil, err
	}

	err = backend.primary.connect()
	if err != nil {
		backend.primary.close()
		backend.primary = nil
		return nil, errors.New("Need to select a new primary")
	}

	backend.primary.lock <- true
	defer func() {
		<-backend.primary.lock
	}()

	if backend.primary.Client == nil {
		return nil, errors.New("Primary client died mid-connection")
	}

	var entries []protos.Giraffe
	err = backend.primary.Client.Call("Backend.ListEntries", 0, &entries)

	if err != nil {
		<-backend.primary.lock
		backend.primary.close()
		backend.primary.lock <- true
		backend.primary = nil
		return nil, err
	}

	return entries, nil
}
