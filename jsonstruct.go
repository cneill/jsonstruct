package jsonstruct

import (
	"fmt"
	"reflect"
)

// JSONStruct is a struct produced from examples.
type JSONStruct struct {
	Name   string
	Fields Fields
}

func (j *JSONStruct) String() string {
	result := fmt.Sprintf("type %s struct {\n", j.Name)
	result += j.Fields.String()
	result += "}"

	for _, field := range j.Fields {
		if field.Child != nil {
			result += fmt.Sprintf("\n\n%s", field.Child.String())
		}
	}

	return result
}

// Equals compares 2 JSONStruct objects and returns true if they're equal.
func (j *JSONStruct) Equals(compare *JSONStruct) bool {
	return reflect.DeepEqual(j, compare)
}

// Sort sorts fields and the fields of their children by name in alphabetical order.
func (j *JSONStruct) Sort() {
	j.Fields.Sort()

	for _, field := range j.Fields {
		if field.Child != nil {
			field.Child.Sort()
		}
	}
}
