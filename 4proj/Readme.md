Leon Chou

# Project 4

Completion Status: Needs Work


# Building

The frontend uses Iris as a webserver.
The backend uses nothing special

# Running

to run the frontend:

from within the frontend directory

`$ go run . --listen <listenaddr> --backend <comma separated list of backends>`

example:

`$ go run . --listen :9090 --backend :8080,:8081,:8082,:8083,:8084`

to run the backend

`$ go run . --listen <listenaddr> --backend <comma separated list of other backends>`

example: 

`$ go run . --listen :8080 --backend :8081,:8082,:8083,:8084`
`$ go run . --listen :8081 --backend :8080,:8082,:8083,:8084`
`$ go run . --listen :8082 --backend :8080,:8081,:8083,:8084`
`$ go run . --listen :8084 --backend :8080,:8081,:8082,:8083`
`$ go run . --listen :8083 --backend :8080,:8081,:8082,:8084`

# State of work

This raft implementation is imperfect.

The voting gets confused, and multiple leaders get elected - i'm not entirely finished with investigating why

Sometimes some writes get caught, or get sent to multiple people - or don't get replicated properly

The client has issues if the server dies in the middle of servicing a request

I used go's RPC library because I wanted to focus on actually implementing the raft protocol

sometimes voting never gets resolved, or there are deadlocks on the leader, and this causes the clients to endlessly deadlock as well - also unsure why


# External help

I used this to help me with raft: https://raft.github.io/raft.pdf

# Resilience

This is theoretically resilient to every node except the leader dying for reads, but for writes will not service if any nodes > n-1/2 nodes die

Test case 1: It passes
Test case 2: this is not good for my cluster
Test case 3: this is mostly fine
Test case 4: this is fine
Test case 5: This is problematic as well

I implemented RAFT

When nodes are terminated, the leader will continuously poll them with heartbeats and log entries until they come back up

Without n-1/2 replicas, the commit will never reach a majority, and so will fail to write/edit/delete entries (however, reading stil works)

Leaders get reelected once other nodes join the cluster, and logs get replicated again, therefore preserving data

I chose Raft, because it was Paxos but easier - spoiler it was still quite difficult to implement.

It allows me to have single-leader and a tough cluster

We use a log to avoid blocking, so we add everything to a log that gets replicated to other nodes, and then committed