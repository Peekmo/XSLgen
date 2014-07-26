package main

import (
	"fmt"
)

// Main point
func main() {
	data, err := GetContent("/home/peekmo/XSLgen/test.xslg")
	if err != nil {
		panic(err)
	}

	fmt.Println(Parse(data))
}
