package jsonstruct

import (
	"fmt"
	"reflect"
)

// Raw is a convenience type for map[string]any which is the default raw message type from encoding/json.
type Raw map[string]any

// JSONStruct is a struct produced from examples.
type JSONStruct struct {
	Name   string
	Fields Fields
}

func (j *JSONStruct) String() string {
	result := fmt.Sprintf("type %s struct {\n", j.Name)

	for _, field := range j.Fields {
		result += "\t" + field.String() + "\n"
	}

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

func (j *JSONStruct) Sort() {
	j.Fields.Sort()

	for _, field := range j.Fields {
		if field.Child != nil {
			field.Child.Sort()
		}
	}
}
