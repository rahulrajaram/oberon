package main

import (
	"os"

	"github.com/fatih/color"
	"github.com/jessevdk/go-flags"
)

type ArgumentParseResult bool

const (
	ERROR   ArgumentParseResult = true
	SUCCESS                     = false
)

const (
	SOURCE_INVALID = "argument error: must pass a valid \"-source\" argument"
)

type Arguments struct {
	result    ArgumentParseResult
	err       error
	arguments map[string]string
}

var opts struct {
	Source string `short:"s" long:"source" description:"the Oberon file to parse"`
}

func parse() Arguments {
	_, err := flags.ParseArgs(&opts, os.Args)
	var args map[string]string = make(map[string]string)

	if err != nil {
		color.Red(err.Error())
		return Arguments{
			result: ERROR,
			err:    err,
		}
	}
	args["source"] = opts.Source
	return Arguments{
		result:    SUCCESS,
		arguments: args,
	}
}
