package main

import (
	"fmt"
	"os"
)

type Box struct {
	Name  string
	Value int
}

// BigBox makes a big box
func BigBox(foo string, value int) Box {
	return Box{
		Name:  foo,
		Value: value,
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func readFirstLine(fname string) string {

	f, err := os.Open(fname)
	check(err)

	defer f.Close()

	buf := make([]byte, 0xff)
	n, err := f.Read(buf)
	check(err)

	fmt.Printf("Read %d bytes\n", n)
	fmt.Printf("Read %s from file", buf)

	return string(buf)
}

func oldmain() {
	boxByName := map[string]Box{}

	box := BigBox("hello", 1)

	boxByName[box.Name] = box

	fmt.Println("hello")

	fmt.Printf("hello, i am a box with %s and value %d\n", box.Name, box.Value)

	_, ok := boxByName["hello"]

	if ok {
		fmt.Println("Found value!")
	} else {
		fmt.Println("no val")
	}

	for i := 0; i < 10; i++ {
		fmt.Printf("A")
	}

	readFirstLine("exercises/exercise1.go")
}

func main() {

	box := map[string]*Box{}

	box["hello"] = &Box{Name: "hello", Value: 1}

	yeet := box["hello"]

	yeet.Name = "hi"

	fmt.Println(box["hello"])
}
