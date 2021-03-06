package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"unicode"

	"../mapreduce"
)

//import "strings"
//import "strconv"
//import "unicode"

// Map function takes a chunk of data from the
// input file and breaks it into a sequence
// of key/value pairs
func Map(value string) []mapreduce.KeyValue {
	kvs := make(map[string]uint64)

	start, end := 0, -1
	inWord := false
	for i, r := range value {
		if !unicode.IsLetter(r) && inWord { // We finished a Word!
			end = i

			word := value[start:end]
			if _, found := kvs[word]; found {
				kvs[word]++
			} else {
				kvs[word] = 1
			}

			inWord = false
		} else if unicode.IsLetter(r) && !inWord {
			start = i
			inWord = true
		}
	}

	mrkvs := []mapreduce.KeyValue{}
	for k, v := range kvs {
		mrkvs = append(mrkvs, mapreduce.KeyValue{Key: k, Value: strconv.FormatUint(v, 10)})
	}

	return mrkvs
}

// Reduce is called once for each key generated by Map, with a list
// of that key's associate values. should return a single
// output value for that key
func Reduce(key string, values []string) string {
	var sum uint64
	for _, v := range values {
		v, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			log.Print(err)
			continue
		}

		sum += v
	}

	return strconv.FormatUint(sum, 10)
}

func main() {
	if len(os.Args) != 4 {
		fmt.Printf("%s: Invalid invocation\n", os.Args[0])
	} else if os.Args[1] == "master" {
		log.Print(os.Args)
		if os.Args[3] == "sequential" {
			mapreduce.RunSingle(5, 3, os.Args[2], Map, Reduce)
		} else {
			mr := mapreduce.MakeMapReduce(5, 3, os.Args[2], os.Args[3])
			// Wait until MR is done
			<-mr.DoneChannel
		}
	} else if os.Args[1] == "worker" {
		mapreduce.RunWorker(os.Args[2], os.Args[3], Map, Reduce, 100)
	} else {
		fmt.Printf("Unexpected input")
	}
}
