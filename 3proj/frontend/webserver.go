package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kataras/iris"
)

// Server collects the webserver for you
type Server struct {
	app     *iris.Application
	backend *Backend
}

// InitWebserver gives back a setup Iris frontend
func InitWebserver(backendAddr string) (*Server, error) {
	app := iris.Default()

	server := &Server{
		app:     app,
		backend: InitBackend(backendAddr),
	}

	app.RegisterView(iris.HTML("./views", ".html"))
	app.Handle("GET", "/health", health)
	app.Handle("GET", "/", server.index)
	app.Handle("POST", "/", server.createGoose)
	app.Handle("GET", "/{id:uint64}", server.individualGoose)
	app.Handle("POST", "/{id:uint64}", server.editGoose)
	app.Handle("POST", "/{id:uint64}/honk", server.honk)
	app.Handle("POST", "/{id:uint64}/delete", server.deleteGoose)

	return server, nil
}

// Run exposes the internal iris Run method
func (server *Server) Run(listenAddr string) {
	stop := make(chan bool)
	healthCheckTicker := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-healthCheckTicker.C:
				err := server.backend.HealthCheck()
				if err != nil {
					log.Printf("Detected failure on %s at %s", server.backend.addr, time.Now().String())
				}
			case <-stop:
				break
			}
		}
	}()

	server.app.Run(iris.Addr(listenAddr))
	stop <- true
	close(stop)
}

func health(ctx iris.Context) {
	ctx.JSON(iris.Map{"healthy": true})
}

func (server *Server) index(ctx iris.Context) {
	geese, err := server.backend.GetGeese()
	if err != nil {
		ctx.StatusCode(500)
		ctx.View("error.html")
		return
	}

	ctx.ViewData("geese", geese)
	ctx.View("index.html")
}

func (server *Server) individualGoose(ctx iris.Context) {
	id, err := ctx.Params().GetUint64("id")
	if err != nil {
		ctx.StatusCode(404)
		ctx.View("404.html")
		return
	}

	goose, err := server.backend.GetGoose(id)
	if err != nil {
		ctx.StatusCode(500)
		ctx.View("error.html")
		return
	}

	ctx.ViewData("goose", goose)
	ctx.View("goose.html")
}

func (server *Server) editGoose(ctx iris.Context) {
	id, err := ctx.Params().GetUint64("id")
	if err != nil {
		ctx.StatusCode(404)
		ctx.View("404.html")
		return
	}

	name := ctx.FormValue("name")

	err = server.backend.EditGoose(id, name)
	if err != nil {
		ctx.StatusCode(500)
		ctx.View("error.html")
		return
	}

	next := ctx.URLParamDefault("next", fmt.Sprintf("/%d", id))
	ctx.Redirect(next)
}

func (server *Server) honk(ctx iris.Context) {
	id, err := ctx.Params().GetUint64("id")
	if err != nil {
		ctx.StatusCode(404)
		ctx.View("404.html")
		return
	}

	err = server.backend.Honk(id)
	if err != nil {
		ctx.StatusCode(500)
		ctx.View("error.html")
		return
	}

	next := ctx.URLParamDefault("next", "/")
	ctx.Redirect(next)
}

func (server *Server) createGoose(ctx iris.Context) {
	name := ctx.FormValue("name")
	_, err := server.backend.CreateGoose(name)

	if err != nil {
		ctx.StatusCode(500)
		ctx.View("error.html")
		return
	}

	next := ctx.URLParamDefault("next", "/")
	ctx.Redirect(next)
}

func (server *Server) deleteGoose(ctx iris.Context) {
	id, err := ctx.Params().GetUint64("id")
	if err != nil {
		ctx.StatusCode(404)
		ctx.View("404.html")
		return
	}

	err = server.backend.DeleteGoose(id)
	if err != nil {
		ctx.StatusCode(500)
		ctx.View("error.html")
		return
	}

	next := ctx.URLParamDefault("next", "/")
	ctx.Redirect(next)
}
