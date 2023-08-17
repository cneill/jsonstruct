package jsonstruct

import (
	"fmt"
	"reflect"
	"sort"
)

// Field represents a single struct field.
type Field struct {
	GoName       string
	OriginalName string
	RawValue     any
}

// Name returns the name of this field as it will be rendered in the final struct.
func (f Field) Name() string {
	return f.GoName
}

// Tag returns the JSON tag as it will be rendered in the final struct.
func (f Field) Tag() string {
	return fmt.Sprintf("`json: \"%s\"`", f.OriginalName)
}

// Type returns the type of the field as it will be rendered in the final struct.
func (f Field) Type() string {
	switch f.RawValue.(type) {
	case int64:
		return "int64"
	case float64:
		return "float64"
	case string:
		return "string"
	case bool:
		return "bool"
	}

	if f.RawValue == nil {
		return "*json.RawMessage"
	}

	if f.IsSlice() {
		return f.SliceType()
	}

	if f.IsStruct() {
		return fmt.Sprintf("*%s", f.GoName)
	}

	return "DUNNO BOSS"
}

func (f Field) SliceType() string {
	rawVal := reflect.ValueOf(f.RawValue)
	rawType := reflect.TypeOf(f.RawValue)

	// we got a non-slice here
	if rawType.Kind() != reflect.Slice {
		return ""
	}

	var sliceType string

	for i := 0; i < rawVal.Len(); i++ {
		idxVal := rawVal.Index(i).Elem()
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
	switch val := f.RawValue.(type) {
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
	return fmt.Sprintf("// Example: %s", f.Value())
}

// IsStruct returns true if RawValue is of kind struct.
func (f Field) IsStruct() bool {
	kind := reflect.TypeOf(f.RawValue).Kind()

	return kind == reflect.Struct
}

// GetStruct gets a the JSONStruct in RawValue if f is a struct, otherwise returns an empty JSONStruct.
func (f Field) GetStruct() JSONStruct {
	if !f.IsStruct() {
		return JSONStruct{}
	}

	js, ok := f.RawValue.(JSONStruct)
	if !ok {
		return JSONStruct{}
	}

	js.Name = f.Name()

	return js
}

// IsSlice returns true if RawValue is of kind slice.
func (f Field) IsSlice() bool {
	kind := reflect.TypeOf(f.RawValue).Kind()

	return kind == reflect.Slice
}

// Fields is a convenience type for a slice of Field structs.
type Fields []Field

func (f Fields) SortAlphabetically() {
	sort.Slice(f, func(i, j int) bool {
		return f[i].GoName < f[j].GoName
	})
}
