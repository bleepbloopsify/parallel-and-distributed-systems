package main

import (
	"fmt"
	"sync"
)

// Cache holds the actual cache
type Cache struct {
	d   map[int]int
	mut *sync.RWMutex
}

type collatz struct {
	Value  int
	Length int
}

func main() {
	var lock sync.RWMutex
	cache := Cache{
		d:   map[int]int{},
		mut: &lock,
	}
	workQueue := make(chan int)
	collectorQueue := make(chan collatz)
	workerNotifyOfCompletion := make(chan bool)
	collectorDone := make(chan bool)

	go collector(collectorQueue, collectorDone)
	for i := 0; i < 5; i++ {
		go worker(cache, workQueue, collectorQueue, workerNotifyOfCompletion)
	}
	go generator(workQueue)

	for i := 0; i < 5; i++ {
		<-workerNotifyOfCompletion
	}
	close(collectorQueue)
	close(workerNotifyOfCompletion)
	<-collectorDone // synchronizing with collector
	close(collectorDone)
}

func generator(workQueue chan int) {
	for i := 1; i < 10000000; i++ {
		workQueue <- i
	}
	close(workQueue)
}

func (cache *Cache) collatz(n int) int {
	cache.mut.RLock()
	if v, ok := cache.d[n]; ok {
		cache.mut.RUnlock()
		return v
	}
	cache.mut.RUnlock()

	if n == 1 {
		cache.mut.Lock()
		cache.d[n] = 0
		cache.mut.Unlock()
		cache.mut.RLock()
		defer cache.mut.RUnlock()
		return cache.d[n]
	}

	if n%2 == 0 {
		val := 1 + cache.collatz(n/2)
		cache.mut.Lock()
		cache.d[n] = val
		cache.mut.Unlock()
	} else {
		val := 1 + cache.collatz(3*n+1)
		cache.mut.Lock()
		cache.d[n] = val
		cache.mut.Unlock()
	}

	cache.mut.RLock()
	defer cache.mut.RUnlock()
	return cache.d[n]
}

func worker(cache Cache, workQueue chan int, collectorQueue chan collatz, workerNotifyOfCompletion chan bool) {
	defer func() {
		workerNotifyOfCompletion <- true
	}()
	for elem := range workQueue {
		obj := collatz{
			Value:  elem,
			Length: cache.collatz(elem),
		}
		collectorQueue <- obj
	}
}

func collector(collectorQueue chan collatz, collectorDone chan bool) {
	defer func() {
		collectorDone <- true
	}()
	var longest collatz // Length always gets initialized to 0
	for elem := range collectorQueue {
		if elem.Length > longest.Length {
			longest = elem
		}
	}

	fmt.Printf("Longest sequence starts at %d, length %d\n", longest.Value, longest.Length)
}
