package main

import (
	"errors"
	"log"

	"github.com/kataras/iris"
)

// Backend holds the address and configuration of the backend service
type Backend struct {
	Addr string
}

// InitBackend pull configuration and verifies that the backend is working
func InitBackend(addr string) (*Backend, error) {
	log.Printf("Attempting to healthcheck backend at %s\n", addr)
	conn, err := Dial(addr)
	if err != nil {
		return nil, errors.New("Backend is unresponsive")
	}

	err = conn.writeFuncID(999)
	if err != nil {
		return nil, err
	}

	err = conn.addBody(struct{}{})
	if err != nil {
		return nil, err
	}

	var obj interface{}
	err = conn.SendReq(&obj)
	if err != nil {
		return nil, err
	}

	// TODO: maybe verify contents of actual service
	// TODO: maybe configure the client with data from the service

	return &Backend{
		Addr: addr,
	}, nil // we've verified that the backend works
}

// Dial wraps connection dialing
func (backend *Backend) Dial() (*Conn, error) {
	return Dial(backend.Addr)
}

func getIndex(ctx iris.Context) {
	ctx.Redirect("/dogs", 301)
}

func (backend *Backend) getDogs(ctx iris.Context) {

	conn, err := backend.Dial()
	if err != nil {
		return
	}

	dogs, err := conn.ListDogs()
	if err != nil {
		return
	}

	ctx.ViewData("dogs", dogs)

	ctx.View("dogs.html")
}

func (backend *Backend) getDog(ctx iris.Context) {
	id, err := ctx.Params().GetInt32("id")
	if err != nil {
		ctx.StatusCode(404)
		return
	}

	conn, err := backend.Dial()
	if err != nil {
		ctx.StatusCode(500)
		return
	}

	dog, err := conn.ReadDog(id)
	if err != nil {
		ctx.StatusCode(404)
		return
	}

	ctx.ViewData("dog", dog)
	ctx.View("dog.html")
}

func (backend *Backend) postDogs(ctx iris.Context) {
	name := ctx.PostValue("name")
	description := ctx.PostValue("description")

	conn, err := backend.Dial()
	if err != nil {
		ctx.StatusCode(500)
		return
	}

	_, err = conn.CreateDog(name, description)
	if err != nil {
		ctx.StatusCode(500)
		return
	}

	ctx.Redirect("/dogs", 302)
}

func (backend *Backend) postDeleteDog(ctx iris.Context) {
	id, err := ctx.Params().GetInt32("id")
	if err != nil {
		ctx.StatusCode(404)
		return
	}

	conn, err := backend.Dial()
	if err != nil {
		ctx.StatusCode(500)
		return
	}

	_, err = conn.DeleteDog(id)
	if err != nil {
		ctx.StatusCode(404)
		return
	}

	ctx.Redirect("/dogs", 302)
}

func (backend *Backend) postDog(ctx iris.Context) {
	id, err := ctx.Params().GetInt32("id")
	if err != nil {
		ctx.StatusCode(404)
		return
	}

	name := ctx.PostValue("name")
	description := ctx.PostValue("description")

	conn, err := backend.Dial()
	if err != nil {
		ctx.StatusCode(500)
		return
	}

	_, err = conn.UpdateDog(&Dog{
		ID:          id,
		Name:        name,
		Description: description,
	})
	if err != nil {
		ctx.StatusCode(404)
		return
	}

	ctx.Redirect("/dogs", 302)
}
