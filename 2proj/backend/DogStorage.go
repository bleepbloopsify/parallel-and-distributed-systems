package main

import (
	"encoding/json"
	"errors"
)

// Dog is the item we are storing
type Dog struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	HasToy      bool   `json:"has_toy"`
	TimesPet    int64  `json:"times_pet"`
	WantsToPlay bool   `json:"wants_to_play"`
}

// DogStorage is the struct we export and methods to operate on it
type DogStorage struct {
	Dogs map[int32]*Dog

	nextID int32 // This is our primary key
	mut    chan bool
}

// InitStorage intializes the storage unit
func InitStorage() *DogStorage {
	return &DogStorage{
		Dogs:   map[int32]*Dog{},
		nextID: 0,
		mut:    make(chan bool, 1),
	}
}

// SeedStorage starts the in memory store with more data
func (store *DogStorage) SeedStorage() {
	var data interface{}
	json.Unmarshal([]byte("{\"name\":\"Doggo1\", \"description\":\"The best doggo 11/10\"}"), &data)
	store.CreateDog(data)
	json.Unmarshal([]byte("{\"name\":\"Doggo2\", \"description\":\"The Smol doggo 12/10\"}"), &data)
	store.CreateDog(data)
	json.Unmarshal([]byte("{\"name\":\"Chonkers\", \"description\":\"The chonkiest doge 13/10\"}"), &data)
	store.CreateDog(data)
}

type createDogPayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateDog takes care of dog creation and adds it to the store
func (store *DogStorage) CreateDog(v interface{}) (*Dog, error) {
	pl, ok := v.(map[string]interface{})
	if !ok {
		return nil, errors.New("Bad json payload")
	}

	dog := &Dog{
		Name:        pl["name"].(string),
		Description: pl["description"].(string),

		HasToy:      false,
		TimesPet:    0,
		WantsToPlay: true,
	}

	// Critical region
	store.mut <- true

	defer func() {
		<-store.mut
	}()

	dog.ID = store.nextID
	store.nextID++
	store.Dogs[dog.ID] = dog

	return dog, nil
}

type readDogPayload struct {
	ID int32 `json:"id"`
}

// ReadDog fetches a dog from the store. only errors for 404s
func (store *DogStorage) ReadDog(v interface{}) (*Dog, error) {
	pl, ok := v.(map[string]interface{})
	if !ok {
		return nil, errors.New("Bad json input")
	}

	ID, ok := pl["id"].(float64)
	if !ok {
		return nil, errors.New("Bad ID field")
	}

	if dog, found := store.Dogs[int32(ID)]; !found {
		return nil, errors.New("no dog found")
	} else {
		return dog, nil
	}
}

// UpdateDog fetches and updates a dog from the store. only errors on 404
func (store *DogStorage) UpdateDog(v interface{}) (*Dog, error) {
	pl, ok := v.(map[string]interface{})
	if !ok {
		return nil, errors.New("Bad json payload")
	}

	store.mut <- true
	defer func() {
		<-store.mut
	}()

	ID, ok := pl["id"].(float64)
	if !ok {
		return nil, errors.New("Missing ID field")
	}

	if sDog, found := store.Dogs[int32(ID)]; found {
		if name, ok := pl["name"].(string); ok {
			sDog.Name = name
		}

		if description, ok := pl["description"].(string); ok {
			sDog.Description = description
		}

		return sDog, nil
	}

	return nil, errors.New("dog not found")
}

// DeleteDog deletes the dog from the storage
func (store *DogStorage) DeleteDog(v interface{}) (interface{}, error) {
	pl, ok := v.(map[string]interface{})
	if !ok {
		return nil, errors.New("Bad json payload")
	}

	store.mut <- true
	defer func() {
		<-store.mut
	}()

	ID, ok := pl["id"].(float64)
	if !ok {
		return nil, errors.New("Missing ID field")
	}

	if _, found := store.Dogs[int32(ID)]; found {
		delete(store.Dogs, int32(ID))
		return struct {
			Success bool `json:"success"`
		}{Success: true}, nil // we need some body to return
	}

	return nil, errors.New("dog not found")
}

// ListDogs gives a listing back for dogs
func (store *DogStorage) ListDogs(v interface{}) ([]*Dog, error) {
	// This doesn't modify the store so we don't need a lock
	dogs := make([]*Dog, len(store.Dogs))
	idx := 0
	for _, dog := range store.Dogs {
		dogs[idx] = dog
		idx++
	}

	return dogs, nil
}
