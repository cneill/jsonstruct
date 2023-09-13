package jsonstruct

import "fmt"

// JSONStruct contains the raw information about a JSON object to be rendered as a Go struct.
type JSONStruct struct {
	name   string
	fields Fields
	// nestLevel tells the formatter what level of indentation to apply if inline structs are enabled.
	nestLevel int
	// inSlice tells the Formatter that this struct is part of a slice and should be de-duplicated rather than repeated.
	inSlice bool
}

// NewJSONStruct returns an initialized JSONStruct.
func New() *JSONStruct {
	return &JSONStruct{
		name:   "JSONStruct",
		fields: Fields{},
	}
}

// SetName sets the name to be used as a type for the JSONStruct.
func (j *JSONStruct) SetName(name string) *JSONStruct {
	if j != nil {
		j.name = name
	}

	return j
}

func (j *JSONStruct) Name() string { return j.name }

// AddFields appends Field objects to the JSONStruct.
func (j *JSONStruct) AddFields(fields ...*Field) *JSONStruct {
	for _, field := range fields {
		if field.IsStruct() || field.IsStructSlice() {
			field.GetStruct().SetNestLevel(j.nestLevel + 1)
		}
	}

	j.fields = append(j.fields, fields...)

	return j
}

func (j *JSONStruct) Fields() Fields { return j.fields }

// AddInlineLevels recursively sets the inlineLevel value for this JSONStruct, as well as its struct fields.
func (j *JSONStruct) SetNestLevel(i int) *JSONStruct {
	if j != nil {
		j.nestLevel = i
	}

	return j
}

func (j *JSONStruct) NestLevel() int { return j.nestLevel }

func (j *JSONStruct) SetInSlice() *JSONStruct {
	if j != nil {
		j.inSlice = true
	}

	return j
}

// JSONStructs is a convenience type for a slice of JSONStruct structs.
type JSONStructs []*JSONStruct

func anySliceToJSONStructs(input []any) (JSONStructs, error) {
	result := JSONStructs{}

	for i, item := range input {
		js, ok := item.(*JSONStruct)
		if !ok {
			return nil, fmt.Errorf("item %d was not a *JSONStruct", i)
		}

		result = append(result, js)
	}

	return result, nil
}
