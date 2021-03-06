package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/fatih/color"

	lexer "oberon/lexer"
	parser "oberon/parser"
	semantic_analyzer "oberon/semantic_analyzer"
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
	debug, _ := strconv.ParseBool(arguments.arguments["debug"])
	lexerResult, err := lexer.Lexer(contents, debug)
	if err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}
	if debug {
		for _, ch := range *lexerResult.Lexemes {
			fmt.Println(ch)
		}
	}
	tree, err1 := parser.Parser(lexerResult.Lexemes, debug)
	if err1 != nil {
		color.Red(err1.Error())
		os.Exit(1)
	}

	if debug {
		parser.PrintParserTree(tree, 0)
	}

	annotated_tree, err2 := semantic_analyzer.Analyze(tree, debug)
	fmt.Println(annotated_tree)
	fmt.Println(err2)
}
