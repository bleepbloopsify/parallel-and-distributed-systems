package main

const (
	// HeartbeatTimeout defines how long it is between heartbeats.
	HeartbeatTimeout = 300
	// ElectionMinTimeout is the bottom range of the randomized election timeout
	ElectionMinTimeout = 350
	// ElectionMaxTimeout describes the maximum range of the randomized election timeout
	ElectionMaxTimeout = 700
)
