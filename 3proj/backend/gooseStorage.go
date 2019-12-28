package main

import (
	"errors"
	"fmt"
)

// Goose describes the state of each goose
type Goose struct {
	ID uint64 `json:"id"`

	Name  string `json:"name"`
	Honks uint64 `json:"honks"`

	lock chan bool // each goose has it's own lock
}

// GooseStorage will take care of multithreaded accesses
type GooseStorage struct {
	geese       map[uint64]*Goose
	transientID uint64

	lock chan bool // the map has a write lock
}

// InitGooseStorage will take care of channels and mutexes
func InitGooseStorage() *GooseStorage {
	store := GooseStorage{
		geese:       make(map[uint64]*Goose),
		transientID: 0,
		lock:        make(chan bool, 1),
	}

	for i := 0; i < 15; i++ {
		store.CreateGoose(fmt.Sprintf("Goose %d", i))
	}

	return &store
}

// CreateGoose takes care of creating each individual geese
func (store *GooseStorage) CreateGoose(name string) *Goose {
	store.lock <- true
	defer func() {
		<-store.lock
	}()

	goose := Goose{
		ID: store.transientID,

		Name:  name,
		Honks: 0,

		lock: make(chan bool, 1),
	}

	store.geese[store.transientID] = &goose
	store.transientID++

	return &goose
}

// EditGoose will override any field in the specified goose
func (store *GooseStorage) EditGoose(ID uint64, name string) error {

	if goose, ok := store.geese[ID]; ok {
		goose.lock <- true
		defer func() {
			<-goose.lock
		}()

		goose.Name = name

		return nil
	}

	return errors.New("Could not find goose")
}

// Honk will make a goose honk
func (store *GooseStorage) Honk(ID uint64) error {
	if goose, ok := store.geese[ID]; ok {
		goose.lock <- true
		defer func() {
			<-goose.lock
		}()

		goose.Honks++

		return nil
	}

	return errors.New("Could not find goose")
}

// DeleteGoose will remove the goose from storage
func (store *GooseStorage) DeleteGoose(ID uint64) error {
	if goose, ok := store.geese[ID]; ok {
		close(goose.lock)
		delete(store.geese, ID)
		return nil
	}

	return errors.New("Could not find goose")
}

// GetGoose will return a single goose (lock is private access)
func (store *GooseStorage) GetGoose(ID uint64) (*Goose, error) {
	if g, ok := store.geese[ID]; ok {
		return g, nil
	}

	return nil, errors.New("Could not findd goose")
}

// GetGeese will return a copy of the list of all the geese
func (store *GooseStorage) GetGeese() []Goose {
	geese := make([]Goose, len(store.geese))
	idx := 0
	for _, g := range store.geese {
		// shallow copy, but anything sensitive is private
		geese[idx] = *g
		idx++
	}

	return geese
}
