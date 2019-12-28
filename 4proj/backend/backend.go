package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"

	"./protos"
)

// Command details our state machine operations
type Command struct {
	Action string
	Data   interface{}
}

// Backend wraps raft and a data store
type Backend struct {
	listen string
	raft   *Server

	idx       uint64
	store     map[uint64]*protos.Giraffe
	storelock chan bool
}

// CreateBackend is a constructor for backend
func CreateBackend(listen string, backends string) *Backend {
	backend := &Backend{
		listen: listen,
		idx:    3,
		store: map[uint64]*protos.Giraffe{
			0: &protos.Giraffe{
				Idx:        0,
				Name:       "leon",
				NeckLength: 15,
			},
			1: &protos.Giraffe{
				Idx:        1,
				Name:       "Giraffe3",
				NeckLength: 257,
			},
			2: &protos.Giraffe{
				Idx:        2,
				Name:       "Bob",
				NeckLength: 12,
			},
		},
		storelock: make(chan bool, 1),
	}

	backend.raft = CreateServer(listen, backends, backend.CommitEntry)

	return backend
}

// CommitEntry is a callback by raft to commit entries to this state machine
func (backend *Backend) CommitEntry(command Command) (interface{}, error) {
	switch command.Action {
	case "CreateGiraffe":
		return backend.createGiraffe(command.Data)
	case "EditGiraffe":
		return backend.editGiraffe(command.Data)
	case "DeleteGiraffe":
		return backend.deleteGiraffe(command.Data)
	}

	return nil, fmt.Errorf("unrecognized command %v", command.Action)
}

// Run will start the backend
func (backend *Backend) Run() error {

	rpc.Register(backend)
	rpc.Register(backend.raft)

	go backend.raft.timeouts()

	rpc.HandleHTTP()

	gob.Register(Command{})
	gob.Register(LogCreateGiraffeArgs{})
	gob.Register(LogEditGiraffeArgs{})

	l, e := net.Listen("tcp", backend.listen)
	if e != nil {
		return e
	}

	return http.Serve(l, nil)
}

// Healthcheck provides a way to check if a server is ready
func (backend *Backend) Healthcheck(args int, reply *bool) error {

	*reply = backend.raft.Ready

	return nil
}

// ListEntries gives back the status on all giraffes
func (backend *Backend) ListEntries(args int, reply *[]protos.Giraffe) error {
	*reply = []protos.Giraffe{}

	for _, giraffe := range backend.store {
		*reply = append(*reply, *giraffe)
	}

	return nil
}

// LogCreateGiraffeArgs passes index and name
type LogCreateGiraffeArgs struct {
	Name string
	Idx  uint64
}

func (backend *Backend) createGiraffe(data interface{}) (*protos.Giraffe, error) {
	backend.storelock <- true
	defer func() {
		<-backend.storelock
	}()

	args := data.(LogCreateGiraffeArgs)

	giraffe := &protos.Giraffe{
		Idx:        args.Idx,
		Name:       args.Name,
		NeckLength: 0,
	}

	backend.store[giraffe.Idx] = giraffe

	log.Printf("Create giraffe %v\n", *giraffe)

	return giraffe, nil
}

// CreateGiraffe exposes RPC to client to request a creation of a giraffe
func (backend *Backend) CreateGiraffe(args *string, reply *protos.Giraffe) error {
	if backend.raft.isLeader() {
		backend.storelock <- true
		command := Command{
			Action: "CreateGiraffe",
			Data: LogCreateGiraffeArgs{
				Name: *args,
				Idx:  backend.idx,
			},
		}
		backend.idx++
		<-backend.storelock

		entry := backend.raft.appendEntry(command)

		<-entry.done
		if entry.error != nil {
			return entry.error
		}

		*reply = *entry.reply.(*protos.Giraffe)

		return nil
	}

	leader := backend.raft.getLeader()
	if leader == nil {
		return errors.New("no leader")
	}
	err := leader.Connect()
	if err != nil {
		return err
	}

	err = leader.client.Call("Backend.CreateGiraffe", args, reply)

	return err
}

// ReadGiraffe expoes RPC to fetch a giraffe. This adds nothing to the log
func (backend *Backend) ReadGiraffe(args *uint64, reply *protos.Giraffe) error {
	if giraffe, found := backend.store[*args]; found {
		*reply = *giraffe
		return nil
	}

	return errors.New("Giraffe not found")
}

func (backend *Backend) editGiraffe(data interface{}) (interface{}, error) {
	args, ok := data.(LogEditGiraffeArgs)
	if !ok {
		return nil, errors.New("Incorrect type passed")
	}

	backend.storelock <- true
	defer func() {
		<-backend.storelock
	}()

	if g, found := backend.store[args.Idx]; found {
		g.Name = args.Name
		g.NeckLength = args.NeckLength
		return *g, nil
	}

	return nil, errors.New("Giraffe not found")
}

// LogEditGiraffeArgs will keep all information necessary to edit a giraffe later
type LogEditGiraffeArgs struct {
	Idx        uint64
	Name       string
	NeckLength uint64
}

// EditGiraffe RPC to create a log entry and edit a giraffe
func (backend *Backend) EditGiraffe(args *LogEditGiraffeArgs, reply *protos.Giraffe) error {

	if backend.raft.isLeader() {
		command := Command{
			Action: "EditGiraffe",
			Data:   *args,
		}

		entry := backend.raft.appendEntry(command)

		<-entry.done
		if entry.error != nil {
			return entry.error
		}

		*reply = entry.reply.(protos.Giraffe)

		return nil
	}

	leader := backend.raft.getLeader()
	if leader == nil {
		return errors.New("no leader")
	}
	err := leader.Connect()
	if err != nil {
		return err
	}

	err = leader.client.Call("Backend.EditGiraffe", args, reply)

	return err
}

func (backend *Backend) deleteGiraffe(data interface{}) (interface{}, error) {
	idx, ok := data.(uint64)
	if !ok {
		return false, errors.New("incorrect type passed")
	}

	backend.storelock <- true
	defer func() {
		<-backend.storelock
	}()

	if _, found := backend.store[idx]; !found {
		return false, errors.New("giraffe not found")
	}

	delete(backend.store, idx)

	return true, nil
}

// DeleteGiraffe is rpc to add a log entry to delete giraffes
func (backend *Backend) DeleteGiraffe(args *uint64, reply *bool) error {
	if backend.raft.isLeader() {
		command := Command{
			Action: "DeleteGiraffe",
			Data:   *args,
		}

		entry := backend.raft.appendEntry(command)

		<-entry.done
		if entry.error != nil {
			return entry.error
		}

		*reply = entry.reply.(bool)

		return nil
	}

	leader := backend.raft.getLeader()
	if leader == nil {
		return errors.New("no leader")
	}

	err := leader.Connect()
	if err != nil {
		return err
	}

	err = leader.client.Call("Backend.DeleteGiraffe", args, reply)

	return err
}
