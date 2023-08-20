package jsonstruct

// JSONStruct contains the raw information about a JSON object to be rendered as a Go struct.
type JSONStruct struct {
	Name   string
	Fields Fields
	// Inline tells the Formatter to format this as an anonymous inline struct
	Inline bool
	// InSlice tells the Formatter that this struct is part of a slice and should be de-duplicated rather than repeated.
	InSlice bool
}

// NewJSONStruct returns an initialized JSONStruct.
func New() *JSONStruct {
	return &JSONStruct{
		Name:   "JSONStruct",
		Fields: Fields{},
	}
}

// SetName sets the name to be used as a type for the JSONStruct.
func (j *JSONStruct) SetName(name string) *JSONStruct {
	j.Name = name

	return j
}

// AddFields appends Field objects to the JSONStruct.
func (j *JSONStruct) AddFields(fields ...*Field) *JSONStruct {
	j.Fields = append(j.Fields, fields...)

	return j
}

func (j *JSONStruct) SetInline() *JSONStruct {
	j.Inline = true

	return j
}

func (j *JSONStruct) SetInSlice() *JSONStruct {
	j.InSlice = true

	return j
}

// JSONStructs is a convenience type for a slice of JSONStruct structs.
type JSONStructs []*JSONStruct
