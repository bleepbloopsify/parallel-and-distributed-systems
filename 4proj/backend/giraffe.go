package main

import "./protos"

// CreateGiraffe serves as a constructor for giraffe
func CreateGiraffe(idx uint64, name string) *protos.Giraffe {
	return &protos.Giraffe{
		Idx:        idx,
		Name:       name,
		NeckLength: 0,
	}
}
