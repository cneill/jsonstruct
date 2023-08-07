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
	// TODO: support multiple comments on a new line above the field definition
	// Comments   []string
	Comment string
}

func (f Field) TypeString() string {
	kind := f.Kind.String()

	switch {
	case f.RawMessage || f.Kind == reflect.Invalid:
		kind = "*json.RawMessage"
	case f.Kind == reflect.Float64:
		if f.RawValue == nil {
			kind = "float64"
		} else if CanBeInt64(f.RawValue.(float64)) {
			kind = "int64"
		}
	case f.Kind == reflect.Map:
		if f.StructType == "" {
			kind = "map[string]any"
		} else {
			kind = fmt.Sprintf("*%s", f.StructType)
		}
	case f.Kind == reflect.Slice:
		switch f.SliceKind {
		case reflect.Invalid:
			kind = "[]*json.RawMessage"
		case reflect.Struct:
			if f.StructType == "" {
				kind = "[]struct{}"
			} else {
				kind = fmt.Sprintf("[]*%s", f.StructType)
			}
		default:
			kind = fmt.Sprintf("[]%s", f.SliceKind.String())
		}
	}

	return kind
}

// Tag returns the tag to include with a struct field.
func (f Field) TagString() string {
	var result string

	if f.OmitEmpty {
		result = fmt.Sprintf("`json:\"%s,omitempty\"`", f.RawName)
	} else {
		result = fmt.Sprintf("`json:\"%s\"`", f.RawName)
	}

	return result
}

func (f Field) CommentString() string {
	if len(f.Comment) > 0 {
		return fmt.Sprintf("// %s\n", f.Comment)
	}

	return "\n"
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

// String properly formats the Fields with spacing/etc as gofmt would.
func (f Fields) String() string {
	var (
		result      string
		longestType string
		longestName string
		longestTag  string
	)

	for _, field := range f {
		if len(field.Name) > len(longestName) {
			longestName = field.Name
		}

		if ts := field.TypeString(); len(ts) > len(longestType) {
			longestType = ts
		}

		if tag := field.TagString(); len(tag) > len(longestTag) {
			longestTag = tag
		}
	}

	fmtString := fmt.Sprintf("%%-%ds%%-%ds%%-%ds%%s", len(longestName)+1, len(longestType)+1, len(longestTag)+1)

	for _, field := range f {
		content := strings.TrimSpace(fmt.Sprintf(fmtString,
			field.Name,
			field.TypeString(),
			field.TagString(),
			field.CommentString(),
		))
		result += fmt.Sprintf("\t%s\n", content)
	}

	return result
}
