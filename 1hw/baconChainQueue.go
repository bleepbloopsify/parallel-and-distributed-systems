package main

import "fmt"

type baconItem struct {
	Actor string
	Movie string
}

// Node is a utility class so we can have methods on the chain (also deepcopy sucks so bad i want to die)
type Node struct {
	Chain []baconItem
}

// InitQueue is a factory for nodes
func InitQueue(target string, movie string) Node {

	chain := []baconItem{
		baconItem{
			Actor: target,
			Movie: movie,
		},
	}

	return Node{
		Chain: chain,
	}
}

// AddToQueue is basically append, but it does a deepcopy
func (node Node) AddToQueue(actor string, movie string) Node {

	newItem := baconItem{
		Actor: actor,
		Movie: movie,
	}

	newChain := make([]baconItem, len(node.Chain))
	for i := range node.Chain {
		newChain[i] = node.Chain[i]
	} // consider this a shallow deep copy because GO SUCKS

	return Node{
		Chain: append(newChain, newItem),
	}
} // This creates a copy so we can reuse the old one

// FetchCurrentMovie grabs the current item in the chain
func (node Node) FetchCurrentMovie() string {
	lastItem := node.Chain[len(node.Chain)-1]

	return lastItem.Movie
}

// PrintNode displays the final product
func (node Node) PrintNode(baconName string) {
	chain := node.Chain

	for i := range chain {
		if i < len(chain)-1 {
			curr, next := chain[i], chain[i+1]

			fmt.Printf("%s was in %s with %s\n", curr.Actor, curr.Movie, next.Actor)
		}
	}

	final := chain[len(chain)-1]
	fmt.Printf("%s was in %s with %s\n", final.Actor, final.Movie, baconName)
	fmt.Printf("Found with a KBN of %d\n", len(chain))
}
