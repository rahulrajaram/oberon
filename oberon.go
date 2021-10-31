package main

import (
	"log"
	"os"
)

func readFile(source string) ([]byte, error) {
	file, err := os.Open(source)
	if err != nil {
		log.Fatal(err)
		return make([]byte, 0), err
	}
	data := make([]byte, 1000)
	_, err = file.Read(data)
	if err != nil {
		log.Fatal(err)
		return make([]byte, 0), err
	}
	return data, nil
}

func main() {
	arguments := parse()
	if arguments.result == ERROR {
		os.Exit(1)
	}
	source := arguments.arguments["source"]
	contents, err := readFile(source)
	if err != nil {
		os.Exit(1)
	}
	lexer(contents)
}
