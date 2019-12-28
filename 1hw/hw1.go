package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Actor holds its own name and references to all its movies
type Actor struct {
	Name   string
	Movies []string
}

// Movie holds its own name and references to its actors
type Movie struct {
	Name   string
	Actors []string
}

// Cast holds actors and movies
type Cast struct {
	Actors map[string]Actor
	Movies map[string]Movie
}

// LoadCast loads the cast.txt file format
func LoadCast(filename string) Cast {
	actors := map[string]Actor{}
	movies := map[string]Movie{}

	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	defer func(scanner *bufio.Scanner) {
		if err := scanner.Err(); err != nil {
			panic(err)
		}
	}(scanner) // err checking for scanner

	for scanner.Scan() {
		movieName := scanner.Text()

		movie := Movie{Name: movieName}

		for scanner.Scan() {
			actorName := scanner.Text()
			if actorName == "" {
				break
			}

			movie.Actors = append(movie.Actors, actorName)

			if actor, found := actors[actorName]; found {
				actor.Movies = append(actor.Movies, movieName)
				actors[actorName] = actor
			} else {
				actors[actorName] = Actor{
					Name:   actorName,
					Movies: []string{movieName},
				}
			}
		}

		movies[movieName] = movie
	}

	return Cast{Movies: movies, Actors: actors}
}

// DiscoverPathFromRoot is a method that operates on a cast of movies and actors, and given a target and a destination, will output the shortest path
func (cast Cast) DiscoverPathFromRoot(target string, baconName string) {
	if target == baconName {
		fmt.Printf("%s is %s\nFound with a KBN of 0\n", target, baconName)
		return
	}

	// We use maps as a set (so we know what we've already traversed)
	traversedActors := map[string]struct{}{}
	traversedMovies := map[string]struct{}{}

	// Setup a queue per level
	queue := []Node{}
	nextLevel := []Node{}

	// Seed the queue
	if actor, found := cast.Actors[target]; !found {
		fmt.Println("Unknown actor")
		return
	} else {
		for _, movie := range actor.Movies {
			queue = append(queue, InitQueue(target, movie))
		}
	}

	traversedActors[target] = struct{}{} // we've traversed the target's movies

	// While we still have things to look at
	for len(queue) > 0 {

		for _, item := range queue {

			movieName := item.FetchCurrentMovie()
			if _, found := traversedMovies[movieName]; found {
				continue
			} // we've already processed this movie, SKIP

			traversedMovies[movieName] = struct{}{}

			movie := cast.Movies[movieName]
			// We want to add each actor and all of their movies to this predecessor
			for _, actorName := range movie.Actors {
				if _, found := traversedActors[actorName]; found {
					continue
				}

				if actorName == baconName {
					item.PrintNode(baconName)
					return
				}

				actor := cast.Actors[actorName]
				// We hop by actor-movie pairs
				for _, nextMovie := range actor.Movies {
					if _, found := traversedMovies[nextMovie]; found {
						continue
					}
					nextLevel = append(nextLevel, item.AddToQueue(actorName, nextMovie))
				}

				// And we have finished processing this actor
				traversedActors[actorName] = struct{}{}
			}
		}

		queue = nextLevel
		nextLevel = []Node{}
	}

	fmt.Println("Infinite KBN")
}

func main() {

	cast := LoadCast("cast.txt")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Enter actor name: ")
		text, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		text = strings.TrimRight(text, "\n")

		if text == "" {
			break
		}

		/** Commands for Debugging use
		if text[:len("actor")] == "actor" {
			fmt.Println(cast.Actors[text[len("actor")+1:]])
			continue
		}

		if text[:len("movie")] == "movie" {
			fmt.Println(cast.Movies[text[len("movie")+1:]])
			continue
		}
		**/
		cast.DiscoverPathFromRoot(text, "Kevin Bacon")
	}

	fmt.Println("Bye!")
}
