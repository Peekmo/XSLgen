package main

import (
	"fmt"
	"reflect"
	"strings"
)

// Special string which implements Value interface
type String string

// Value is an interface for every type of a tag value
type Value interface {
	Print(depth int) string
}

// Tags metadata
type Tag struct {
	parent     *TagList
	namespace  string
	name       string
	attributes map[string]string
	children   Value
}

// Contains a list of tags
type TagList struct {
	values []Tag
}

// Print Prints String's value
func (this String) Print(depth int) string {
	return fmt.Sprintf("%s", this)
}

// Print Prints the tag list
func (this *TagList) Print(depth int) string {
	s := getTabs(depth)
	var strtag string

	for _, tag := range this.values {
		// Start tag
		strtag += s + "<"

		// Namespace (if any)
		if tag.namespace != "" {
			strtag += tag.namespace + ":"
		}

		// Tag name
		strtag += tag.name

		// Tag attributes
		for key, value := range tag.attributes {
			strtag += " " + key + "=\"" + value + "\""
		}

		if tag.children == nil {
			strtag += "/>\n"
		} else {
			strtag += ">"

			val := strings.Split(reflect.TypeOf(tag.children).String(), ".")
			if val[1] == "TagList" {
				strtag += "\n"
			}

			strtag += tag.children.Print(depth + 1)

			// Start tag
			strtag += "</"

			// Namespace (if any)
			if tag.namespace != "" {
				strtag += tag.namespace + ":"
			}

			// Tag name
			strtag += tag.name + ">\n"
		}
	}

	return strtag
}

// Get a tab string for the given depth
func getTabs(depth int) string {
	var s string
	for i := 0; i < depth; i++ {
		s += "\t"
	}

	return s
}
