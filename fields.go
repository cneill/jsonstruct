package jsonstruct

import (
	"fmt"
	"reflect"
	"sort"
)

// Field represents a single struct field.
type Field struct {
	goName       string
	originalName string
	rawValue     any
	optional     bool
}

func (f *Field) SetName(originalName string) *Field {
	f.goName = GetGoName(originalName)
	f.originalName = originalName

	return f
}

func (f *Field) SetValue(value any) *Field {
	f.rawValue = value

	return f
}

func (f *Field) SetOptional() *Field {
	f.optional = true

	return f
}

// Name returns the name of this field as it will be rendered in the final struct.
func (f Field) Name() string {
	return f.goName
}

// Tag returns the JSON tag as it will be rendered in the final struct.
func (f Field) Tag() string {
	if f.optional {
		return fmt.Sprintf("`json: \"%s,omitempty\"", f.Name())
	}

	return fmt.Sprintf("`json: \"%s\"`", f.Name())
}

// Type returns the type of the field as it will be rendered in the final struct.
func (f Field) Type() string {
	switch f.rawValue.(type) {
	case int64:
		return "int64"
	case float64:
		return "float64"
	case string:
		return "string"
	case bool:
		return "bool"
	}

	if f.rawValue == nil {
		return "*json.RawMessage"
	}

	if f.IsSlice() {
		return f.SliceType()
	}

	if f.IsStruct() {
		return fmt.Sprintf("*%s", f.goName)
	}

	return "any"
}

func (f Field) SliceType() string {
	rawVal := reflect.ValueOf(f.rawValue)
	rawType := reflect.TypeOf(f.rawValue)

	// we got a non-slice here
	if rawType.Kind() != reflect.Slice {
		return ""
	}

	if f.IsStructSlice() {
		return fmt.Sprintf("[]*%s", f.Name())
	}

	var sliceType string

	for i := 0; i < rawVal.Len(); i++ {
		idxVal := rawVal.Index(i)
		kind := idxVal.Type().Kind()

		if kind == reflect.Pointer || kind == reflect.Interface {
			idxVal = idxVal.Elem()
		}

		idxType := idxVal.Type()

		if sliceType != "" && idxType.String() != sliceType {
			sliceType = "*json.RawMessage"

			break
		}

		sliceType = idxType.String()
	}

	return fmt.Sprintf("[]%s", sliceType)
}

// Value returns the string version of RawValue.
func (f Field) Value() string {
	switch val := f.rawValue.(type) {
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		return fmt.Sprintf("%f", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case string:
		return val
	default:
		if val == nil {
			return "null"
		}

		// TODO: simple arrays?

		return ""
	}
}

// Comment returns the string used for example value comments.
func (f Field) Comment() string {
	val := f.Value()
	if val != "" {
		return fmt.Sprintf("// Example: %s", f.Value())
	}

	return ""
}

// IsStruct returns true if RawValue is of kind struct.
func (f Field) IsStruct() bool {
	kind := reflect.TypeOf(f.rawValue).Kind()

	return kind == reflect.Struct
}

// GetStruct gets a the JSONStruct in RawValue if f is a struct, otherwise returns an empty JSONStruct.
func (f Field) GetStruct() JSONStruct {
	switch {
	case f.IsStruct():
		js, ok := f.rawValue.(JSONStruct)
		if !ok {
			return JSONStruct{}
		}

		return js.SetName(f.Name())
	case f.IsStructSlice():
		return f.GetSliceStruct()
	default:
		return JSONStruct{}
	}
}

func (f Field) GetSliceStruct() JSONStruct {
	result := (&JSONStruct{}).SetName(f.Name())

	anySlice, ok := f.rawValue.([]any)
	if !ok {
		return JSONStruct{}
	}

	jss, err := anySliceToJSONStructs(anySlice)
	if err != nil {
		return JSONStruct{}
	}

	foundFields := map[string][]*Field{}

	// we have a slice of structs, each of which may or may not contain the full set of fields
	for _, js := range jss {
		for _, field := range js.Fields {
			foundFields[field.Name()] = append(foundFields[field.Name()], field)
		}
	}

	for _, fields := range foundFields {
		if len(fields) != len(jss) {
			fields[0].SetOptional()
		}

		result.AddFields(fields[0])
	}

	return result
}

// IsSlice returns true if RawValue is of kind slice.
func (f Field) IsSlice() bool {
	kind := reflect.TypeOf(f.rawValue).Kind()

	return kind == reflect.Slice
}

func (f Field) IsStructSlice() bool {
	anySlice, ok := f.rawValue.([]any)
	if !ok {
		return false
	}

	if _, err := anySliceToJSONStructs(anySlice); err != nil {
		return false
	}

	return true
}

func (f Field) Equals(input Field) bool {
	switch {
	case f.Name() != input.Name():
		return false
	case f.Type() != input.Type():
		return false
	case f.Tag() != input.Tag():
		return false
	}

	return true
}

// Fields is a convenience type for a slice of Field structs.
type Fields []*Field

func (f Fields) SortAlphabetically() {
	sort.Slice(f, func(i, j int) bool {
		return f[i].goName < f[j].goName
	})
}
