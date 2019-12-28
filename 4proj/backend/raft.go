package main

import (
	"log"
	"math/rand"
	"strings"
	"time"
)

// Entry describes entries into the server's state machine log
type Entry struct {
	Term    uint64
	Index   uint64
	Command Command

	done  chan bool
	reply interface{}
	error error
}

// Server describes the entire backend
type Server struct {
	Self string

	nodes  map[string]*Node
	Term   uint64
	Leader string

	Ready bool

	State string

	votedFor string
	votes    uint64

	commitIndex uint64
	lastApplied uint64

	log []*Entry

	logLock chan bool

	lock        chan bool
	heartbeat   chan bool
	stopLeading chan bool

	commit func(Command) (interface{}, error)
}

// CreateServer initializes a server
func CreateServer(listen string, backends string, commit func(Command) (interface{}, error)) *Server {
	server := &Server{
		Self:      listen,
		Term:      0,
		Leader:    "",
		nodes:     map[string]*Node{},
		lock:      make(chan bool, 1),
		heartbeat: make(chan bool, 1),

		Ready: false,

		State: "follower",

		commitIndex: 0,
		lastApplied: 0,

		log:         []*Entry{&Entry{Index: 0, Term: 0, Command: Command{}}},
		logLock:     make(chan bool, 1),
		stopLeading: make(chan bool, 1),

		commit: commit,
	}

	server.parseBackends(backends)

	return server
}

func (server *Server) isLeader() bool {
	return server.Leader == server.Self
}

func (server *Server) getLeader() *Node {
	return server.nodes[server.Leader]
}

func (server *Server) parseBackends(backends string) {
	for _, addr := range strings.FieldsFunc(backends, func(c rune) bool {
		return c == ','
	}) {
		server.nodes[addr] = CreateNode(server, addr)
	}
}

// Timeouts keeps track of the election timeout
func (server *Server) timeouts() {
	for {
		timeout := rand.Intn(ElectionMaxTimeout-ElectionMinTimeout) + ElectionMinTimeout

		go server.applyLogs()

		select {
		case <-server.heartbeat:
			if !server.Ready {
				server.lock <- true
				server.Ready = true
				<-server.lock
			}
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			log.Println("Heartbeat timeout passed, election starting")
			go server.startCandidacy()
		}
	}
}

func (server *Server) startCandidacy() {
	if server.State == "candidate" {
		log.Println("Am a candidate, continuing")
		return
	}

	server.lock <- true
	defer func() {
		<-server.lock
	}()

	log.Println("Starting Candidacy")
	server.Term++
	server.votes = 1
	server.votedFor = server.Self
	server.State = "candidate"

	votes := make(chan bool, len(server.nodes))

	for _, node := range server.nodes {
		go node.RequestVote(votes)
	}

	for range server.nodes {
		vote := <-votes
		if vote {
			server.votes++
		}
	}

	close(votes)

	log.Printf("Acquired %v votes\n", server.votes)

	if server.votes > uint64(len(server.nodes))/2 {
		server.State = "leader"
		server.Leader = server.Self
		go server.Lead()
		return
	}

	server.State = "follower"
}

func (server *Server) appendEntry(data Command) *Entry {
	server.logLock <- true
	defer func() {
		<-server.logLock
	}()

	entry := &Entry{
		Index:   uint64(len(server.log)),
		Term:    server.Term,
		Command: data,

		done: make(chan bool, 1),
	}

	server.log = append(server.log, entry)

	return entry
}

// Lead is for the server that will start leadering
func (server *Server) Lead() {
	log.Printf("Attempting to lead term %v\n", server.Term)

	server.lock <- true
	server.Ready = true // I am the leader so I am always ready
	<-server.lock

	for {
		select {
		case <-server.stopLeading:
			log.Println("Stopping leading!")
			server.lock <- true
			defer func() {
				<-server.lock
			}()
			server.State = "follower"
			return
		case <-time.After(HeartbeatTimeout * time.Millisecond):
			for _, node := range server.nodes {
				go node.AppendEntries()
			}

			go server.commitMajority()

			server.heartbeat <- true
		}
	}
}

// Start will register server under RPC
func (server *Server) Start() error {

	go server.timeouts()

	return nil
}

func (server *Server) applyLogs() {
	server.logLock <- true
	defer func() {
		<-server.logLock
	}()

	if server.lastApplied < server.commitIndex {
		server.apply(server.log[server.lastApplied+1])
		server.lastApplied++
	}
}

func (server *Server) apply(entry *Entry) {
	log.Println("Applying entry!")
	reply, err := server.commit(entry.Command)
	if err != nil {
		entry.error = err
		log.Println(err)
	}

	log.Printf("reply, err: %v, %v", reply, err)

	entry.reply = reply

	if entry.done != nil {
		entry.done <- true
		close(entry.done)
	}
}

// commitMajority will attempt to figure out an appropriate commitIndex
func (server *Server) commitMajority() {
	// calculate an N such that N > commitIndex, a majority of matchIndex[i] ≥ N, and log[N].term == currentTerm

	if uint64(len(server.log)-1) == server.commitIndex {
		return
	}

	for n := server.commitIndex + 1; n < uint64(len(server.log)); n++ {
		count := 0
		for _, node := range server.nodes {
			if node.matchIndex >= n {
				count++
			}
		}
		if count > len(server.nodes)/2 && server.log[n].Term == server.Term {
			server.commitIndex = n
		} else {
			break
		}
	}
}

// AppendEntriesArgs contains heartbeat and log update information
type AppendEntriesArgs struct {
	Term         uint64
	Leader       string
	PrevLogIndex uint64
	PrevLogTerm  uint64
	Entries      []Entry
	LeaderCommit uint64
}

// AppendEntriesReply gives details on success of the heartbeat
type AppendEntriesReply struct {
	Term    uint64
	Success bool
}

// AppendEntries is both heartbeat and update in a single RPC
func (server *Server) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) error {
	server.lock <- true
	defer func() {
		<-server.lock
	}()

	reply.Term = server.Term

	if args.Term < server.Term {
		log.Println("Stale leader")
		reply.Success = false
		return nil
	}

	if args.Term > server.Term {
		if server.Leader == server.Self {
			log.Println("No longer leader")
			server.stopLeading <- true
		}

		server.Term = args.Term
		server.Leader = args.Leader
		server.votedFor = ""
		log.Printf("Changing leader to %v\n", args.Leader)
	}

	if args.Term == server.Term && server.Leader != args.Leader {
		if server.State == "leader" {
			log.Println("Competing leader in the group!")
			go server.startCandidacy()
		} else {
			server.Leader = args.Leader
			log.Printf("Changing leader to %v\n", args.Leader)
		}
	}

	if args.PrevLogIndex > uint64(len(server.log)) {
		log.Println("They are starting way after we are")
		reply.Success = false
		return nil
	}

	if args.PrevLogIndex >= uint64(len(server.log)) {
		reply.Success = false
		return nil
	}

	if args.PrevLogIndex != 0 {
		entry := server.log[args.PrevLogIndex]
		if entry.Term != args.PrevLogTerm {
			log.Println("Log history does not match, go back more")
			reply.Success = false
			return nil
		}
	}

	reply.Success = true
	server.heartbeat <- true

	if len(args.Entries) == 0 {
		if args.LeaderCommit > server.commitIndex {
			server.commitIndex = args.LeaderCommit
		} // We move the commit index back to the leader (in case)
		return nil
	}

	// Now we know that the prev index matches, we can update the rest of the log with new entries

	log.Printf("Received entries %v\n", args.Entries)

	var entry Entry
	for _, entry := range args.Entries {
		if entry.Index >= uint64(len(server.log)) {
			server.log = append(server.log, &entry)
		} else {
			existing := server.log[entry.Index]
			if existing.Term != entry.Term {
				server.log = server.log[:entry.Index]
				server.log = append(server.log, &entry)
			}
		}
	}

	if args.LeaderCommit > server.commitIndex {
		if args.LeaderCommit < entry.Index {
			server.commitIndex = args.LeaderCommit
		} else {
			server.commitIndex = entry.Index
		}
	} // We move the commit index back to the leader (in case)

	return nil
}

// RequestVoteArgs defines the params required to request a vote from this node
type RequestVoteArgs struct {
	Candidate string
	Term      uint64

	LastLogIndex uint64
	LastLogTerm  uint64
}

// RequestVoteReply holds the response to candidate requests
type RequestVoteReply struct {
	Term        uint64
	VoteGranted bool
}

// RequestVote allows candidates to request votes
func (server *Server) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) error {
	server.lock <- true
	defer func() {
		<-server.lock
	}()

	reply.Term = server.Term

	if args.Term < server.Term {
		reply.VoteGranted = false
		return nil
	}

	if args.Term > server.Term && args.LastLogIndex >= server.lastApplied {
		server.votedFor = args.Candidate
		server.Term = args.Term
		server.heartbeat <- true
		reply.VoteGranted = true
		return nil
	}

	server.Term = args.Term

	if (server.votedFor == "" || server.votedFor == args.Candidate) &&
		args.LastLogIndex >= server.lastApplied {
		server.votedFor = args.Candidate
		server.heartbeat <- true
		reply.VoteGranted = true
		return nil
	}

	reply.VoteGranted = false
	return nil
}
