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

	t := Tag{
		"tif",
		"OI",
		map[string]string{"id": "ID"},
		String("salut"),
	}

	t3 := Tag{
		"tif",
		"Default",
		map[string]string{"id": "qdfsd", "attr": "okok"},
		String("salut"),
	}

	t2 := Tag{
		"tif",
		"Classification",
		map[string]string{"id": "ID"},
		&TagList{
			[]Tag{t, t3},
		},
	}

	list := TagList{
		[]Tag{t, t2},
	}

	fmt.Println(list.Print(0))
}
