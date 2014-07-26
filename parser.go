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
	DECLARATION_TAG              = -100
	DECLARATION_ATTRIBUTES       = -101
	DECLARATION_ATTRIBUTES_KEY   = -102
	DECLARATION_ATTRIBUTES_VALUE = -103
	DECLARATION_TAG_CONTENT      = -104
	DECLARATION_STRING           = -105
	STRING_CONTENT               = -106
)

type Stack struct {
	str  string
	kind int
}

type Stacktrace struct {
	stacktrace []Stack
}

var (
	stacktrace *Stacktrace // Actions in progress
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
	var index int = 0
	var strxsl string
	var current *Tag
	var lastStack, tempLast *Stack
	stacktrace = &Stacktrace{}

	taglist = &TagList{values: []*Tag{}}

	for number, line := range content {
		index = 0
		if line == "" {
			continue
		}

		if stacktrace.Size() == 0 {
			if strings.Index(line, "@?xml") != 0 {
				return strxsl, gerror(number, "An XSLgen must start with @xml tag")
			}

			index = 5
			lastStack = stacktrace.Append(Stack{"", DECLARATION_TAG})
			current = &Tag{parent: taglist, name: "?xml", appended: false}
		}

		for ; index < len(line); index++ {
			c := line[index]

			if c == '#' && lastStack.kind != STRING_CONTENT && lastStack.kind != DECLARATION_ATTRIBUTES_VALUE {
				break
			}

			// If string, adds the char to the string
			if lastStack.kind == STRING_CONTENT {
				if c != '"' {
					lastStack.str += string(c)
				} else {
					current.children = String(lastStack.str)
					current.appended = true
					current.parent.values = append(current.parent.values, current)

					_, _ = stacktrace.RemoveLastElement()         // End string_content
					_, lastStack = stacktrace.RemoveLastElement() // End declaration_string
				}

				// If string declaration, adds the char to the string
			} else if lastStack.kind == DECLARATION_STRING && (c == ' ' || c == '"') {
				if c == '"' {
					lastStack = stacktrace.Append(Stack{"", STRING_CONTENT})
				}

				// Tag declaration & name
			} else if (lastStack.kind == DECLARATION_TAG && c != '[') || (lastStack.kind == DECLARATION_TAG_CONTENT && c == '@' || c == ' ' || c == '&') {
				// New tag
				if c == '@' {
					if current != nil && current.appended == false {
						current.appended = true
						current.parent.values = append(current.parent.values, current)
					}

					current = &Tag{parent: taglist, name: "", appended: false, namespace: "xsl"}

					if lastStack.kind == DECLARATION_TAG {
						_, _ = stacktrace.RemoveLastElement()
					}

					lastStack = stacktrace.Append(Stack{"", DECLARATION_TAG})

					// Start block
				} else if c == '{' {
					current.appended = true
					current.parent.values = append(current.parent.values, current)
					taglist = &TagList{parent: taglist, values: []*Tag{}}
					current = nil

					lastStack = stacktrace.Append(Stack{"", DECLARATION_TAG_CONTENT})

					// End block
				} else if c == '}' {
					if current.appended == false {
						current.appended = true
						current.parent.values = append(current.parent.values, current)
					}

					taglist.parent.values[len(taglist.parent.values)-1].children = taglist

					taglist = taglist.parent
					current = taglist.values[len(taglist.values)-1]

					_, _ = stacktrace.RemoveLastElement()         // End tag
					_, lastStack = stacktrace.RemoveLastElement() // End tag content

					// String
				} else if c == ':' {
					lastStack = stacktrace.Append(Stack{"", DECLARATION_STRING})

					// Othewise
				} else if c != ' ' {
					current.name += string(c)
				}

				// Attribute
			} else if lastStack.kind == DECLARATION_ATTRIBUTES && c != ']' {
				if c != ' ' {
					lastStack = stacktrace.Append(Stack{"\"" + string(c), DECLARATION_ATTRIBUTES_KEY})
				}

				// Attribute key
			} else if lastStack.kind == DECLARATION_ATTRIBUTES_KEY {
				// Double points : separator key:value
				if c == ':' {
					lastStack.str += "\":"
					tempLast, lastStack = stacktrace.RemoveLastElement()
					lastStack.str += tempLast.str

					lastStack = stacktrace.Append(Stack{"", DECLARATION_ATTRIBUTES_VALUE})
				} else {
					lastStack.str += string(c)
				}

				// Attribute value
			} else if lastStack.kind == DECLARATION_ATTRIBUTES_VALUE && c != ']' {
				lastStack.str += string(c)

				if c == ',' {
					tempLast, lastStack = stacktrace.RemoveLastElement()
					lastStack.str += tempLast.str
				}

				// Otherwise
			} else {
				if c == '[' {
					lastStack = stacktrace.Append(Stack{"{", DECLARATION_ATTRIBUTES})
				} else if c == ']' {

					if lastStack.kind == DECLARATION_ATTRIBUTES_VALUE {
						tempLast, lastStack = stacktrace.RemoveLastElement()
						lastStack.str += tempLast.str
					} else if lastStack.kind != DECLARATION_ATTRIBUTES {
						return strxsl, gerror(number, "Unexpected token ']'")
					}

					lastStack.str += "}"
					if err = json.Unmarshal([]byte(lastStack.str), &current.attributes); err != nil {
						return strxsl, gerror(number, err.Error())
					}

					_, lastStack = stacktrace.RemoveLastElement()
				} else {
					return strxsl, gerror(number, fmt.Sprintf("Unexpected token %c", c))
				}
			}
		}
	}

	if current != nil && current.appended == false {
		taglist.values = append(taglist.values, current)
	}

	return taglist.Print(0), nil
}

// Returns an error message
func gerror(line int, message string) error {
	return errors.New(fmt.Sprintf("[Line %d] %s", line+1, message))
}

// RemoveLastElement removes the last element of the stackstrace
// Returns the removed element & the current last element
func (this *Stacktrace) RemoveLastElement() (tmpLast *Stack, currentLast *Stack) {
	var last = &this.stacktrace[len(this.stacktrace)-1]

	this.stacktrace = this.stacktrace[0 : len(this.stacktrace)-1]

	if len(this.stacktrace) > 0 {
		return last, &this.stacktrace[len(this.stacktrace)-1]
	}

	return last, nil
}

// Append appends an elements to the stacktrace and returns it
func (this *Stacktrace) Append(stack Stack) *Stack {
	this.stacktrace = append(this.stacktrace, stack)
	return &this.stacktrace[len(this.stacktrace)-1]
}

// Size returns stacktrace size
func (this *Stacktrace) Size() int {
	return len(this.stacktrace)
}
