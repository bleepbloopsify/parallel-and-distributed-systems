# Name

Leon Chou

# Building

`$ cd ./backend && go build .`
`$ cd ./frontend && go build .`

# Running

`$ cd ./backend && ./backend`
`$ cd ./frontend && ./frontend`

# Building and running with docker-compose

`$ docker-compose up`


# State of work

Assignment is completed

# Design decisions

I chose to use a simple TCP connection with length-prefixed JSON bodies as a means of data transfer. There are no concepts of sessions, and each connection is individual.

There are theoretically issues with hanging connections on the backend.

# To note

I used docker to test the multiple frontend parts of this project.

I used go-iris for the frontend, and `go mod` to manage the dependencies in the projects.