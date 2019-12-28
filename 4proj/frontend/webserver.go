package main

import (
	"log"

	"github.com/kataras/iris"
)

// Webserver describes the frontend webserver
type Webserver struct {
	app     *iris.Application
	backend *Backend
}

// CreateWebserver serves as a constructor for the webserver
func CreateWebserver(backends string) *Webserver {
	app := iris.Default()

	server := &Webserver{
		app:     app,
		backend: CreateBackend(backends),
	}

	app.RegisterView(iris.HTML("./views", ".html"))
	app.Handle("GET", "/", server.index)
	app.Handle("POST", "/", server.postIndex)
	app.Handle("GET", "/:id", server.getSpecific)
	app.Handle("POST", "/:id", server.postEdit)
	app.Handle("POST", "/:id/delete", server.postDelete)

	server.backend.selectPrimary()

	return server
}

// ListenAndServe wraps app.Run(iris.Addr(addr))
func (server *Webserver) ListenAndServe(addr string) error {
	return server.app.Run(iris.Addr(addr))
}

func (server *Webserver) index(ctx iris.Context) {
	entries, err := server.backend.ListEntries()
	if err != nil {
		ctx.StatusCode(500)
		ctx.ViewData("Error", err)
		ctx.View("error.html")
		return
	}

	ctx.ViewData("Giraffes", entries)
	ctx.View("index.html")
}

func (server *Webserver) postIndex(ctx iris.Context) {
	log.Print("Hello from postIndex")

	_, err := server.backend.CreateGiraffe(ctx.FormValue("name"))
	if err != nil {
		ctx.StatusCode(500)
		ctx.ViewData("Error", err)
		ctx.View("error.html")
		return
	}

	ctx.Redirect("/", 302)
}

func (server *Webserver) getSpecific(ctx iris.Context) {
	id, err := ctx.Params().GetUint64("id")
	if err != nil {
		ctx.StatusCode(400)
		ctx.ViewData("Error", err)
		ctx.View("error.html")
	}

	giraffe, err := server.backend.ReadGiraffe(id)
	if err != nil {
		ctx.StatusCode(500)
		ctx.ViewData("Error", err)
		ctx.View("error.html")
	}

	ctx.ViewData("Giraffe", giraffe)
	ctx.View("specific.html")
}

func (server *Webserver) postDelete(ctx iris.Context) {
	id, err := ctx.Params().GetUint64("id")
	if err != nil {
		ctx.StatusCode(400)
		ctx.ViewData("Error", err)
		ctx.View("error.html")
	}

	err = server.backend.DeleteGiraffe(id)
	if err != nil {
		ctx.StatusCode(500)
		ctx.ViewData("Error", err)
		ctx.View("error.html")
	}

	ctx.Redirect("/", 302)
}

func (server *Webserver) postEdit(ctx iris.Context) {
	id, err := ctx.Params().GetUint64("id")
	if err != nil {
		ctx.StatusCode(400)
		ctx.ViewData("Error", err)
		ctx.View("error.html")
	}

	err = server.backend.UpdateGiraffe(&LogEditGiraffeArgs{
		Idx:        id,
		Name:       ctx.FormValue("name"),
		NeckLength: uint64(ctx.PostValueInt64Default("necklength", 0)),
	})

	if err != nil {
		ctx.StatusCode(500)
		ctx.ViewData("Error", err)
		ctx.View("error.html")
	}

	ctx.Redirect("/", 302)
}
