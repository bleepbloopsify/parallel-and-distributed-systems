package main

// Kitten is a struct that we use to keep track of our kittens
type Kitten struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	Hungry           bool   `json:"is_hungry"`
	TimesFed         int64  `json:"times_fed"`
	TimesRefusedFood int64  `json:"times_refused_food"`
}

// KittenStorage is a storage struct for kittens.
// This keeps track of the next proper ID for the kitten as well.
type KittenStorage struct {
	Kittens       map[int64]*Kitten
	NextAvailable int64
	Mutex         chan bool
}

// InitStorage initializes the storage for kittens
func InitStorage() *KittenStorage {
	mutex := make(chan bool, 1)
	mutex <- true

	return &KittenStorage{
		Kittens:       map[int64]*Kitten{},
		NextAvailable: 0,
		Mutex:         mutex,
	}
}

// CreateKitten Creates and stores a kitten in the store, and returns the kitten created
func (store *KittenStorage) CreateKitten(name string) *Kitten {

	kitten := &Kitten{
		Name:             name,
		Hungry:           true,
		TimesFed:         0,
		TimesRefusedFood: 0,
	}

	// Critical region
	<-store.Mutex
	kitten.ID = store.NextAvailable
	store.Kittens[kitten.ID] = kitten
	store.NextAvailable++
	store.Mutex <- true
	// End critical region

	return kitten
}

// EditKitten takes an ID and updates the fields in the memory store
func (store *KittenStorage) EditKitten(id int64, name string) *Kitten {
	<-store.Mutex
	defer func() {
		store.Mutex <- true
	}()
	if kitten, found := store.Kittens[id]; found {

		if name != "" {
			kitten.Name = name // this is how forms work
		}

		return kitten
	}

	return nil
}

// FeedKitten defines feeding kitten behavior
func (store *KittenStorage) FeedKitten(id int64) *Kitten {
	<-store.Mutex
	defer func() {
		store.Mutex <- true
	}()
	if kitten, found := store.Kittens[id]; found {
		if !kitten.Hungry { // kitten won't eat if its hungry
			kitten.TimesRefusedFood++
			return kitten
		}
		kitten.Hungry = false
		kitten.TimesFed++
		return kitten
	}

	return nil
}

// ReadKitten Returns the state of a kitten in storage
func (store *KittenStorage) ReadKitten(id int64) *Kitten {
	<-store.Mutex
	defer func() {
		store.Mutex <- true
	}()
	if kitten, found := store.Kittens[id]; found {
		return kitten
	}
	return nil
}

// DeleteKitten Pops a kitten out of the storage
func (store *KittenStorage) DeleteKitten(id int64) *Kitten {
	<-store.Mutex
	defer func() {
		store.Mutex <- true
	}()
	if kitten, found := store.Kittens[id]; found && kitten != nil {
		copied := kitten
		store.Kittens[id] = nil
		delete(store.Kittens, id)

		return copied
	}

	return nil
}

// ListKittens returns a list of all kittens
func (store *KittenStorage) ListKittens() []*Kitten {
	<-store.Mutex
	defer func() {
		store.Mutex <- true
	}()

	idx := 0
	kittens := make([]*Kitten, len(store.Kittens))
	for _, kitten := range store.Kittens {
		kittens[idx] = kitten
		idx++
	}

	return kittens
}

// PlayingKittens will make kittens hungry again
func (store *KittenStorage) PlayingKittens() {
	<-store.Mutex
	defer func() {
		store.Mutex <- true
	}()

	for _, kitten := range store.Kittens {
		if kitten.Hungry {
			continue
		}

		kitten.Hungry = true
		kitten.TimesRefusedFood = 0
	}
}

// ByID sorts by ID
type ByID []*Kitten

func (a ByID) Len() int           { return len(a) }
func (a ByID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByID) Less(i, j int) bool { return a[i].ID < a[j].ID }
