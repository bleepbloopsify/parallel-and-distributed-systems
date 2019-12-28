package main

import (
	"flag"
	"log"

	"github.com/kataras/iris"
)

func main() {
	listenAddr := flag.String("listen", ":8080", "The address to bind the webserver to")
	backendAddr := flag.String("backend", ":8090", "The address and port of the backend service")
	flag.Parse()

	backend, err := InitBackend(*backendAddr)
	if err != nil {
		log.Fatal(err)
	}

	app := iris.New()

	app.RegisterView(iris.HTML("./views", ".html"))

	app.Get("/", getIndex)
	app.Get("/dogs", backend.getDogs)
	app.Get("/dogs/{id:int32}", backend.getDog)
	app.Post("/dogs/{id:int32}", backend.postDog)
	app.Post("/dogs", backend.postDogs)
	app.Post("/dogs/{id:int32}/delete", backend.postDeleteDog)

	app.Run(iris.Addr(*listenAddr))
}
