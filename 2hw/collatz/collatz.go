package main

import (
	"fmt"
)

type collatz struct {
	Value  int
	Length int
}

func main() {
	workQueue := make(chan int)
	collectorQueue := make(chan collatz)
	workerNotifyOfCompletion := make(chan bool)
	collectorDone := make(chan bool)

	go collector(collectorQueue, collectorDone)
	for i := 0; i < 5; i++ {
		go worker(workQueue, collectorQueue, workerNotifyOfCompletion)
	}
	go generator(workQueue)

	for j := 0; j < 5; j++ {
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

func worker(workQueue chan int, collectorQueue chan collatz, workerNotifyOfCompletion chan bool) {
	defer func() {
		workerNotifyOfCompletion <- true
	}()

	for elem := range workQueue {
		obj := collatz{
			Value: elem,
		}
		for elem != 1 {
			if elem%2 == 0 {
				elem = elem / 2
			} else {
				elem = 3*elem + 1
			}
			obj.Length++
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
