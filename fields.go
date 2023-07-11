package jsonstruct

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Field describes a struct field in a JSONStruct.
type Field struct {
	RawName    string
	RawValue   any
	Name       string
	Value      reflect.Value
	Kind       reflect.Kind
	SliceKind  reflect.Kind
	StructType string
	Child      *JSONStruct
	RawMessage bool
	OmitEmpty  bool
	Comments   []string
}

// Tag returns the tag to include with a struct field.
func (f Field) Tag() string {
	var result string

	if f.OmitEmpty {
		result = fmt.Sprintf("`json:\"%s,omitempty\"`", f.RawName)
	} else {
		result = fmt.Sprintf("`json:\"%s\"`", f.RawName)
	}

	return result
}

func (f Field) String() string {
	result := fmt.Sprintf("%s %s %s", f.Name, f.TypeString(), f.Tag())

	if len(f.Comments) > 1 {
		result = fmt.Sprintf("// %s\n%s", strings.Join(f.Comments, "\n//"), result)
	} else if len(f.Comments) == 1 {
		result = fmt.Sprintf("%s // %s", result, f.Comments[0])
	}

	return result
}

func (f Field) TypeString() string {
	kind := f.Kind.String()

	switch {
	case f.RawMessage:
		kind = "*json.RawMessage"
	case f.Kind == reflect.Float64:
		if CanBeInt64(f.RawValue.(float64)) {
			kind = "int64"
		}
	case f.Kind == reflect.Map:
		if f.StructType == "" {
			kind = "map[string]any"
		} else {
			kind = fmt.Sprintf("*%s", f.StructType)
		}
	case f.Kind == reflect.Slice:
		if f.SliceKind == reflect.Struct {
			if f.StructType == "" {
				kind = "[]struct{}"
			} else {
				kind = fmt.Sprintf("[]*%s", f.StructType)
			}
		} else {
			kind = fmt.Sprintf("[]%s", f.SliceKind.String())
		}
	}

	return kind
}

// Equals compares 2 Field objects to see if they are have the same fields.
// TODO: improve this
func (f Field) Equals(compare Field) bool {
	switch {
	case f.Name != compare.Name:
		return false
	case f.RawName != compare.RawName:
		return false
	case f.Kind != compare.Kind:
		return false
	case f.SliceKind != compare.SliceKind:
		return false
	case f.StructType != compare.StructType:
		return false
	}

	return true
}

// GetExampleString returns the "// Ex: [whatever]" text when value comments are enabled.
func (f Field) ExampleString() string {
	return getExampleString(f.RawValue)
}

type Fields []Field

func (f Fields) Sort() {
	sort.Slice(f, func(i, j int) bool {
		return strings.ToLower(f[i].Name) < strings.ToLower(f[j].Name)
	})
}
