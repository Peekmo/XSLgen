package main

import (
	"flag"
	"fmt"
)

var (
	file string
)

// Main point
func main() {
	initFlags()
	flag.Parse()

	if file == "" {
		panic("Fatal Error : File not defined")
	}

	data, err := GetContent(file)
	if err != nil {
		panic(err)
	}

	xslstr, err := Parse(data)
	if err != nil {
		panic(err)
	}

	fmt.Println(xslstr)
}

// initFlages defines all flag
func initFlags() {
	flag.StringVar(&file, "file", "", "Path to the file to parse")
}
