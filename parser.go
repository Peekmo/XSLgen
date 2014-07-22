package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	EOF        int = -1
	IDENTIFIER     = -2
	NUMBER         = -3
)

const (
	DECLARATION_XML              = -100
	DECLARATION_ATTRIBUTES       = -101
	DECLARATION_ATTRIBUTES_KEY   = -102
	DECLARATION_ATTRIBUTES_VALUE = -103

	STRING = -150
)

type Stack struct {
	str  string
	kind int
}

var (
	stacktrace []Stack // Actions in progress
	taglist    *TagList
)

// Reads the file content of the given file
// Removes new lines, additional tabs & spaces
func GetContent(path string) (content []string, err error) {
	input, err := os.Open(path)
	if err != nil {
		return []string{}, err
	}

	defer input.Close()
	scanner := bufio.NewScanner(input)

	var current string
	var data []string
	for scanner.Scan() {
		current = scanner.Text()
		current = strings.TrimSpace(current)

		data = append(data, current)
	}

	if err := scanner.Err(); err != nil {
		return data, err
	}

	return data, nil
}

// Parse the given array to generate the XSL
func Parse(content []string) (xsl string, err error) {
	var strxsl string
	var index int = 0
	var current *Tag
	var lastStack, tempLast *Stack

	taglist = &TagList{values: []*Tag{}}

	for number, line := range content {
		if line == "" {
			continue
		}

		if len(stacktrace) == 0 {
			if strings.Index(line, "@xml") != 0 {
				return strxsl, gerror(number, "An XSLgen must start with @xml tag")
			}

			index = 4
			stacktrace, lastStack = appendStack(stacktrace, &Stack{"", DECLARATION_XML})
			current = &Tag{name: "?xml"}
		}

		for ; index < len(line); index++ {
			c := line[index]
			lastStack = &stacktrace[len(stacktrace)-1]

			// If string, adds the char to the string
			if lastStack.kind == STRING && c != '"' {
				lastStack.str += string(c)

				// Attribute
			} else if lastStack.kind == DECLARATION_ATTRIBUTES && c != ']' {
				if c != ' ' {
					stacktrace, lastStack = appendStack(stacktrace, &Stack{"\"" + string(c), DECLARATION_ATTRIBUTES_KEY})
				}

				// Attribute key
			} else if lastStack.kind == DECLARATION_ATTRIBUTES_KEY {
				// Double points : separator key:value
				if c == ':' {
					lastStack.str += "\":"
					stacktrace, tempLast, lastStack = removeLastStacktraceElement(stacktrace)
					lastStack.str += tempLast.str

					stacktrace, lastStack = appendStack(stacktrace, &Stack{"", DECLARATION_ATTRIBUTES_VALUE})
				} else {
					lastStack.str += string(c)
				}

				// Attribute value
			} else if lastStack.kind == DECLARATION_ATTRIBUTES_VALUE && c != ']' {
				lastStack.str += string(c)

				if c == ',' {
					stacktrace, tempLast, lastStack = removeLastStacktraceElement(stacktrace)
					lastStack.str += tempLast.str
				}

				// Otherwise
			} else {
				if c == '[' {
					stacktrace, lastStack = appendStack(stacktrace, &Stack{"{", DECLARATION_ATTRIBUTES})
				} else if c == ']' {

					if lastStack.kind == DECLARATION_ATTRIBUTES_VALUE {
						stacktrace, tempLast, lastStack = removeLastStacktraceElement(stacktrace)
						lastStack.str += tempLast.str
					} else if lastStack.kind != DECLARATION_ATTRIBUTES {
						return strxsl, gerror(number, "Unexpected token ']'")
					}

					lastStack.str += "}"
					if err = json.Unmarshal([]byte(lastStack.str), &current.attributes); err != nil {
						return strxsl, gerror(number, err.Error())
					}

					stacktrace, _, _ = removeLastStacktraceElement(stacktrace)
				}
			}
		}
	}

	taglist.values = append(taglist.values, current)

	return taglist.Print(0), nil
}

// Returns an error message
func gerror(line int, message string) error {
	return errors.New(fmt.Sprintf("[Line %d] %s", line, message))
}

// Removes the last element from a Stack array & returns it
func removeLastStacktraceElement(stacktrace []Stack) (trace []Stack, tempLast *Stack, lastStack *Stack) {
	var last = &stacktrace[len(stacktrace)-1]
	return stacktrace[0 : len(stacktrace)-1], last, &stacktrace[len(stacktrace)-2]
}

func appendStack(stacktrace []Stack, stack *Stack) (trace []Stack, lastStack *Stack) {
	stacktrace = append(stacktrace, *stack)
	return stacktrace, stack
}
