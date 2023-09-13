package jsonstruct

import (
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"
)

const jsonRawMessage = "*json.RawMessage"

// Field represents a single struct field.
type Field struct {
	goName       string
	originalName string
	rawValue     any
	optional     bool
	isJSONRaw    bool
}

func NewField() *Field {
	return &Field{}
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

func (f *Field) SetJSONRaw() *Field {
	f.isJSONRaw = true

	return f
}

// Name returns the name of this field as it will be rendered in the final struct.
func (f Field) Name() string {
	return f.goName
}

func (f Field) OriginalName() string {
	return f.originalName
}

// Tag returns the JSON tag as it will be rendered in the final struct.
func (f Field) Tag() string {
	if f.originalName == f.Name() {
		return ""
	}

	if f.optional {
		return fmt.Sprintf("`json:\"%s,omitempty\"`", f.originalName)
	}

	return fmt.Sprintf("`json:\"%s\"`", f.originalName)
}

// Type returns the type of the field as it will be rendered in the final struct.
func (f Field) Type() string {
	if f.rawValue == nil || f.isJSONRaw {
		return jsonRawMessage
	}

	switch f.rawValue.(type) {
	case int64:
		return "int64"
	case *big.Int:
		return "*big.Int"
	case float64:
		return "float64"
	case *big.Float:
		return "*big.Float"
	case string:
		return "string"
	case bool:
		return "bool"
	}

	if f.IsSlice() {
		return f.SliceType()
	}

	if f.IsStruct() {
		return fmt.Sprintf("*%s", f.goName)
	}

	return "any"
}

// SliceType returns the type of the slice this field represents.
func (f Field) SliceType() string {
	// rawVal := reflect.ValueOf(f.rawValue)
	rawType := reflect.TypeOf(f.rawValue)

	if rawType.Kind() != reflect.Slice {
		return ""
	}

	// we have a field that represents a struct, whose slice type is its name
	if f.IsStructSlice() {
		return fmt.Sprintf("[]*%s", f.Name())
	}

	return getSliceType(f.rawValue)
}

func getSliceType(input any) string {
	var (
		rawVal = reflect.ValueOf(input)
		// rawType   = reflect.TypeOf(input)
		sliceType string
	)

	// walk the array's elements and figure out if they're all typed the same
	for i := 0; i < rawVal.Len(); i++ {
		var itemType string

		idxVal := rawVal.Index(i)
		kind := idxVal.Type().Kind()

		if kind == reflect.Pointer || kind == reflect.Interface {
			idxVal = idxVal.Elem()
		}

		if !idxVal.IsValid() {
			sliceType = jsonRawMessage

			break
		}

		idxType := idxVal.Type()
		kind = idxType.Kind()

		if kind == reflect.Slice {
			itemType = getSliceType(idxVal.Interface())
		} else {
			itemType = idxType.String()
		}

		// we have encountered multiple types for this field, have to accept anything
		if sliceType != "" && itemType != sliceType {
			sliceType = jsonRawMessage

			break
		}

		sliceType = itemType
	}

	return fmt.Sprintf("[]%s", sliceType)
}

// Value returns the string version of RawValue.
func (f Field) Value() string {
	if val := simpleAnyToString(f.rawValue); val != "" {
		return val
	}

	if f.rawValue == nil {
		return "null"
	}

	if vals := f.SimpleSliceValues(); f.IsSlice() && len(vals) > 0 {
		return fmt.Sprintf("[%s]", strings.Join(vals, ", "))
	}

	return ""
}

// SimpleSliceValues returns a slice of strings with the values of simple slice Fields ([]int64, []float64, []bool,
// []string). If it doesn't recognize the Field as one of these, it returns an empty slice.
func (f Field) SimpleSliceValues() []string {
	results := []string{}

	if !f.IsSlice() {
		return []string{}
	}

	rawVal := reflect.ValueOf(f.rawValue)

	switch f.SliceType() {
	case "[]int64", "[]float64", "[]bool", "[]string":
		for i := 0; i < rawVal.Len(); i++ {
			idxVal := rawVal.Index(i)
			kind := idxVal.Type().Kind()

			if kind == reflect.Pointer || kind == reflect.Interface {
				idxVal = idxVal.Elem()
			}

			if !idxVal.IsValid() {
				return []string{}
			}

			results = append(results, simpleAnyToString(idxVal.Interface()))
		}
	}

	return results
}

// Comment returns the string used for example value comments.
func (f Field) Comment() string {
	comment := ""
	cleanVal := strings.ReplaceAll(f.Value(), "\n", "\\n")

	if val := f.Value(); val != "" {
		comment = fmt.Sprintf("// Example: %s", cleanVal)
	}

	if len(comment) > 50 {
		comment = fmt.Sprintf("%s...", comment[0:47])
	}

	return comment
}

// IsStruct returns true if RawValue is either a struct or a pointer to a struct.
func (f Field) IsStruct() bool {
	if f.rawValue == nil {
		return false
	}

	typ := reflect.TypeOf(f.rawValue)
	kind := typ.Kind()

	if kind == reflect.Ptr {
		kind = typ.Elem().Kind()
	}

	return kind == reflect.Struct
}

// GetStruct gets a the JSONStruct in RawValue if f is a struct or slice of struct, otherwise returns nil.
func (f Field) GetStruct() *JSONStruct {
	switch {
	case f.IsStruct():
		js, ok := f.rawValue.(*JSONStruct)
		if !ok {
			return nil
		}

		return js.SetName(f.Name())
	case f.IsStructSlice():
		return f.GetSliceStruct()
	default:
		return nil
	}
}

func (f Field) GetSliceStruct() *JSONStruct {
	anySlice, ok := f.rawValue.([]any)
	if !ok {
		return nil
	}

	return getSliceStruct(anySlice).SetName(f.Name())
}

func getSliceStruct(input []any) *JSONStruct {
	result := &JSONStruct{}

	jStructs, err := anySliceToJSONStructs(input)
	if err != nil {
		return nil
	}

	// foundFields contains the first instance of a field, while fieldCounts reports the number of structs containing it
	// have to use synced slices here to avoid the reordering that would occur with a map
	foundFields := []*Field{}
	fieldTypes := []string{}
	fieldCounts := []int{}

	// have a slice of structs, each of which may or may not contain the full set of fields - walk each and find the
	// fields that don't reoccur
	for _, jStruct := range jStructs {
		for _, field := range jStruct.Fields() {
			foundIndex := -1

			for i, foundField := range foundFields {
				if field.OriginalName() == foundField.OriginalName() {
					fieldCounts[i]++

					foundIndex = i

					break
				}
			}

			if foundIndex == -1 {
				foundFields = append(foundFields, field)
				fieldTypes = append(fieldTypes, field.Type())
				fieldCounts = append(fieldCounts, 1)
			} else if field.Type() != fieldTypes[foundIndex] {
				foundFields[foundIndex].SetJSONRaw()
			}
		}
	}

	for i := 0; i < len(foundFields); i++ {
		if fieldCounts[i] != len(jStructs) {
			foundFields[i].SetOptional()
		}

		result.AddFields(foundFields[i])
	}

	return result
}

// IsSlice returns true if RawValue is of kind slice.
func (f Field) IsSlice() bool {
	kind := reflect.TypeOf(f.rawValue).Kind()

	return kind == reflect.Slice
}

func (f Field) IsStructSlice() bool {
	return isStructSlice(f.rawValue)
}

func isStructSlice(input any) bool {
	anySlice, ok := input.([]any)
	if !ok {
		return false
	}

	if _, err := anySliceToJSONStructs(anySlice); err != nil {
		return false
	}

	return true
}

// Equals returns true if two Fields share an original name, Go name, and type - does not compare values!
func (f Field) Equals(input *Field) bool {
	switch {
	case f.Name() != input.Name():
		return false
	case f.Type() != input.Type():
		return false
	case f.OriginalName() != input.OriginalName():
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
