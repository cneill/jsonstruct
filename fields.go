package jsonstruct

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Field describes a struct field in a JSONStruct.
type Field struct {
	Name       string
	RawName    string
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
	kind := f.Kind.String()

	if f.Kind == reflect.Map {
		if f.StructType == "" {
			kind = "map[string]any"
		} else {
			kind = fmt.Sprintf("*%s", f.StructType)
		}
	} else if f.Kind == reflect.Slice {
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

	if f.RawMessage {
		kind = "*json.RawMessage"
	}

	result := fmt.Sprintf("%s %s %s", f.Name, kind, f.Tag())

	if len(f.Comments) > 1 {
		result = fmt.Sprintf("// %s\n%s", strings.Join(f.Comments, "\n//"))
	} else if len(f.Comments) == 1 {
		result = fmt.Sprintf("%s // %s", result, f.Comments[0])
	}

	return result
}

// Equals compares 2 Field objects to see if they are have the same fields.
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

type Fields []Field

func (f Fields) Sort() {
	sort.Slice(f, func(i, j int) bool {
		return strings.ToLower(f[i].Name) < strings.ToLower(f[j].Name)
	})
}
