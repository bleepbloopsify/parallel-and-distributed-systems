package main

import (
	"errors"
	"log"
	"net/rpc"
)

// Node describes the state of a member of the cluster
type Node struct {
	Addr   string
	client *rpc.Client

	nextIndex  uint64
	matchIndex uint64

	server *Server // We need a reference here
	lock   chan bool
}

// CreateNode is a consstructor for node
func CreateNode(server *Server, addr string) *Node {
	return &Node{
		Addr: addr,

		client: nil,
		server: server,

		nextIndex:  1,
		matchIndex: 0,

		lock: make(chan bool, 1),
	}
}

// Connect will attempt to negotiate a connection with a node
func (node *Node) Connect() error {
	node.lock <- true
	defer func() {
		<-node.lock
	}()

	if node.client != nil {
		return nil
	}

	client, err := rpc.DialHTTP("tcp", node.Addr)
	if err != nil {
		return err
	}

	node.client = client

	return nil
}

// Close wraps node.client.Close()
func (node *Node) Close() {
	node.lock <- true
	defer func() {
		<-node.lock
	}()

	if node.client == nil {
		return
	}

	node.client.Close()
	node.client = nil
}

// RequestVote will send a request to this node for a vote for this term
func (node *Node) RequestVote(votes chan bool) error {
	err := node.Connect()
	if err != nil {
		log.Printf("Node %v is dead, no votes granted\n", node.Addr)
		votes <- false
		return err
	}

	lastLog := node.server.log[len(node.server.log)-1]

	args := &RequestVoteArgs{
		Candidate: node.server.Self,
		Term:      node.server.Term,

		LastLogIndex: lastLog.Index,
		LastLogTerm:  lastLog.Term,
	}

	log.Printf("Requested votes from %v as candidate %v for term %v\n", node.Addr, args.Candidate, args.Term)

	node.lock <- true
	defer func() {
		<-node.lock
	}()

	if node.client == nil {
		return errors.New("Client died mid connection")
	}

	var reply RequestVoteReply
	err = node.client.Call("Server.RequestVote", args, &reply)

	if err != nil {
		log.Printf("RequestVote error %v\n", err)
		votes <- false
		<-node.lock
		node.Close()
		node.lock <- true
		return err
	}

	votes <- reply.VoteGranted

	return nil
}

// AppendEntries is a goroutine for the client to synchronize the target node's log
func (node *Node) AppendEntries() {
	err := node.Connect()
	if err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panicked with node %v\n", node)
			log.Fatal(r)
		}
	}()

	entries := []Entry{}
	if len(node.server.log) > 0 {
		for _, entry := range node.server.log[node.nextIndex:] {
			entries = append(entries, *entry)
		}
	}

	args := &AppendEntriesArgs{
		Term:         node.server.Term,
		Leader:       node.server.Self,
		Entries:      entries,
		PrevLogIndex: 0,
		PrevLogTerm:  node.server.Term,
		LeaderCommit: node.server.commitIndex,
	}

	if node.nextIndex != 0 {
		args.PrevLogIndex = node.nextIndex - 1
		args.PrevLogTerm = node.server.log[node.nextIndex-1].Term
	}

	node.lock <- true
	defer func() {
		<-node.lock
	}()

	if node.client == nil {
		return
	}

	var reply AppendEntriesReply
	err = node.client.Call("Server.AppendEntries", args, &reply)
	if err != nil {
		<-node.lock
		node.Close()
		node.lock <- true
		return
	}

	if reply.Success == false {
		node.server.lock <- true
		if reply.Term > node.server.Term {
			node.server.Term = reply.Term
		}
		<-node.server.lock

		if node.nextIndex > 0 {
			node.nextIndex--
		}
		return
	}

	if len(entries) != 0 {
		top := entries[len(entries)-1]
		node.nextIndex = top.Index + 1
		node.matchIndex = top.Index
	}
}
