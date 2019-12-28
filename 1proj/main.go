package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"
)

// App holds information about the server thats running
type App struct {
	Storage *KittenStorage
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/kittens", 302)
}

func (app *App) handleListKittens(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		log.Printf("GET %s", r.URL)
		t, err := template.ParseFiles("kittens.html")
		if err != nil {
			log.Print(err)
			errorHandler(w, r, 500, errors.New("Error parsing template"))
			return
		}

		kittens := app.Storage.ListKittens()
		sort.Sort(ByID(kittens))

		t.Execute(w, struct{ Kittens []*Kitten }{
			Kittens: kittens,
		})
		return
	}
	http.Error(w, "Error not allowed", 405)
}

func (app *App) handleKitten(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		log.Printf("GET %s", r.URL)
		id, found := r.URL.Query()["id"]
		if !found || len(id) == 0 {
			errorHandler(w, r, 400, errors.New("Missing id query param"))
			return
		}
		i, err := strconv.ParseInt(id[0], 10, 64)
		if err != nil {
			errorHandler(w, r, 400, errors.New("id query param must be integer"))
			return
		}
		kitten := app.Storage.ReadKitten(i)
		t, err := template.ParseFiles("kitten.html")
		if err != nil {
			errorHandler(w, r, 500, errors.New("Error parsing template"))
			return
		}
		t.Execute(w, kitten)
		return
	} else if r.Method == "POST" {
		log.Printf("POST %s", r.URL)
		name := r.FormValue("name")
		if name == "" {
			errorHandler(w, r, 404, errors.New("Missing name param"))
			return
		}

		kitten := app.Storage.CreateKitten(name)

		if next := r.FormValue("next"); next != "" {
			http.Redirect(w, r, next, 302)

			return
		}

		http.Redirect(w, r, fmt.Sprintf("/kitten?id=%d", kitten.ID), 302)
		return
	} else if r.Method == "DELETE" {
		log.Printf("DELETE %s", r.URL)
		id, found := r.URL.Query()["id"]
		if !found || len(id) == 0 {
			errorHandler(w, r, 400, errors.New("Missing id query param"))
			return
		}
		i, err := strconv.ParseInt(id[0], 10, 64)
		if err != nil {
			errorHandler(w, r, 400, errors.New("id query param must be integer"))
			return
		}
		deleted := app.Storage.DeleteKitten(i)
		if deleted == nil {
			errorHandler(w, r, 404, errors.New("Kitten doesn't exist"))
			return
		}
	}

	http.Error(w, "Error not allowed", 405)
}

func (app *App) handleUpdateKitten(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		log.Printf("POST %s", r.URL)
		id, found := r.URL.Query()["id"]
		if !found || len(id) == 0 {
			errorHandler(w, r, 400, errors.New("Missing id query param"))
			return
		}
		i, err := strconv.ParseInt(id[0], 10, 64)
		if err != nil {
			errorHandler(w, r, 400, errors.New("id query param must be integer"))
			return
		}

		kitten := app.Storage.EditKitten(i, r.FormValue("name"))
		if kitten == nil {
			errorHandler(w, r, 404, errors.New("Could not find kitten to edit"))
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/kitten?id=%d", kitten.ID), 302)
		return
	}
}

func (app *App) handleDeleteKitten(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		log.Printf("POST %s", r.URL)
		id, found := r.URL.Query()["id"]
		if !found || len(id) == 0 {
			errorHandler(w, r, 400, errors.New("Missing id query param"))
			return
		}
		i, err := strconv.ParseInt(id[0], 10, 64)
		if err != nil {
			errorHandler(w, r, 400, errors.New("id query param must be integer"))
			return
		}

		kitten := app.Storage.DeleteKitten(i)
		if kitten == nil {
			errorHandler(w, r, 404, errors.New("Could not find kitten to delete"))
			return
		}

		http.Redirect(w, r, "/kittens", 302)
		return
	}
}

func (app *App) handleFeedKitten(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		log.Printf("POST %s", r.URL)

		id, found := r.URL.Query()["id"]
		if !found || len(id) == 0 {
			errorHandler(w, r, 400, errors.New("Missing id query param"))
			return
		}
		i, err := strconv.ParseInt(id[0], 10, 64)
		if err != nil {
			errorHandler(w, r, 400, errors.New("id query param must be integer"))
			return
		}

		kitten := app.Storage.FeedKitten(i)
		if kitten == nil {
			errorHandler(w, r, 404, errors.New("Could not find kitten to feed"))
			return
		}

		if next := r.FormValue("next"); next != "" {
			http.Redirect(w, r, next, 302)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/kitten?id=%d", kitten.ID), 302)
		return
	}
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int, err error) {
	w.WriteHeader(status)
	fmt.Fprintf(w, "%v", err)
}

// PrefillData will add some dummy kittens
func (app *App) PrefillData() {
	app.Storage.CreateKitten("meow")
	app.Storage.CreateKitten("meow 2")
	app.Storage.CreateKitten("meow 3")
}

func main() {
	port := flag.Int("listen", 8080, "the port to listen on")
	flag.Parse()

	app := &App{
		Storage: InitStorage(),
	}
	app.PrefillData()

	// Ticker to play with kittens
	ticker := time.NewTicker(10 * time.Second)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				app.Storage.PlayingKittens()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	// Fin play loop

	// We write to DefaultThreadingMux because we are only hosting one web server
	// with this binary
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/kittens", app.handleListKittens)
	http.HandleFunc("/kitten", app.handleKitten)
	http.HandleFunc("/kitten/update", app.handleUpdateKitten)
	http.HandleFunc("/kitten/delete", app.handleDeleteKitten)
	http.HandleFunc("/kitten/feed", app.handleFeedKitten)

	log.Printf("Listening on port %d..", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	close(quit) // stop ticking the playingkittens
	if err != nil {
		log.Fatal(err)
	}
}
