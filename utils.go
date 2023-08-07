package jsonstruct

import (
	"fmt"
	"math"
	"path"
	"reflect"
	"strings"
)

// GetNormalizedName takes in a field name or file name and returns a "normalized" (CamelCase) string suitable for use as a Go
// variable name.
func GetNormalizedName(input string) string {
	var cleaned strings.Builder

	// remove garbage characters, replace separators with ' '
	for _, r := range input {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			cleaned.WriteRune(r)
		} else if r == '_' || r == '.' || r == '-' || r == ' ' {
			cleaned.WriteRune(' ')
		}
	}

	// look for initialisms to capitalize
	words := strings.Split(cleaned.String(), " ")
	temp := []string{}

	for _, word := range words {
		tmpWord := strings.ToUpper(word)
		if _, ok := CommonInitialisms[tmpWord]; ok {
			word = strings.ToUpper(tmpWord)
		}

		temp = append(temp, word)
	}

	result := strings.Title(strings.Join(temp, " "))
	result = strings.ReplaceAll(result, " ", "")

	return result
}

// GetSliceKind takes any slice object and returns the Kind of elements within the slice. If all elements are of the
// same Kind (as determined by 'fuzzyKind' below), it returns that Kind. If the slice contains multiple Kinds, this
// function returns reflect.Invalid.
func GetSliceKind(value any) (reflect.Kind, error) {
	var (
		typeOf = reflect.TypeOf(value)
		kind   = typeOf.Kind()
	)

	if kind != reflect.Slice {
		return reflect.Invalid, fmt.Errorf("must provide a value with Kind == Slice")
	}

	valOf := reflect.ValueOf(value)

	// If we have no elements in this slice, just return Invalid and render it as json.RawMessage later.
	if valOf.Len() == 0 {
		return reflect.Invalid, nil
	}

	var firstKind reflect.Kind

	for i := 0; i < valOf.Len(); i++ {
		elemVal := valOf.Index(i)
		iface := elemVal.Interface()
		fuzzy := fuzzyKind(iface)

		if i == 0 {
			firstKind = fuzzy
		} else if firstKind != reflect.Invalid && fuzzy != firstKind {
			// We have a slice with multiple Kinds, so return Invalid and render it as json.RawMessage later.
			return reflect.Invalid, nil
		}
	}

	return firstKind, nil
}

func fuzzyKind(input any) reflect.Kind {
	switch input.(type) {
	case int, int8, int16, int32, int64:
		return reflect.Int64
	case float32, float64:
		return reflect.Float64
	case string:
		return reflect.String
	}

	elemType := reflect.TypeOf(input)

	if elemKind := elemType.Kind(); elemKind == reflect.Map {
		return reflect.Struct
	}

	return reflect.Invalid
}

// GetFieldKind takes any object and returns the Kind represented.
func GetFieldKind(value any) reflect.Kind {
	if value == nil {
		return reflect.String
	}

	typeOf := reflect.TypeOf(value)

	kind := typeOf.Kind()
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.Int64
	case reflect.Float32, reflect.Float64:
		return reflect.Float64
	case reflect.Bool, reflect.String, reflect.Map, reflect.Slice:
		return kind
	}

	fmt.Printf("KIND: %s\nTypeOf: %#v\n", kind, typeOf)
	fmt.Printf("%#v\n", value)

	return reflect.String
}

// NameFromInputFile strips the file path and extension, using GetNormalizedName to return a struct name.
func NameFromInputFile(inputFile string) string {
	_, fName := path.Split(inputFile)
	name := strings.TrimSuffix(inputFile, path.Ext(fName))

	return GetNormalizedName(name)
}

// CanBeInt64 checks whether a float value can be converted into an int64 without a loss of precision. Helps find e.g.
// IDs, counts, and so on.
// TODO: come up with a better way of finding e.g. 1.0 - currently, truncating 1.0 will satisfy this and result in an
// int64 type when it should be a float64.
func CanBeInt64(f float64) bool {
	return f > float64(math.MinInt64) && f < float64(math.MaxInt64) && f == math.Trunc(f)
}

// getExampleString returns the "// Ex: [whatever]" text when value comments are enabled.
func getExampleString(rawValue any) string {
	var (
		val    = reflect.ValueOf(rawValue)
		result string
	)

	if !val.IsValid() {
		return "nil"
	}

	switch val.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, ok := rawValue.(int64)
		if !ok {
			return ""
		}

		result = fmt.Sprintf("%d", intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		uintValue, ok := rawValue.(uint64)
		if !ok {
			return ""
		}

		result = fmt.Sprintf("%d", uintValue)
	case reflect.Bool:
		boolValue, ok := rawValue.(bool)
		if !ok {
			return ""
		}

		result = fmt.Sprintf("%t", boolValue)
	case reflect.Float32, reflect.Float64:
		floatValue, ok := rawValue.(float64)
		if !ok {
			return ""
		}

		if CanBeInt64(floatValue) {
			result = fmt.Sprintf("%d", int64(floatValue))
		} else {
			result = fmt.Sprintf("%.2f", floatValue)
		}
	case reflect.String:
		strValue, ok := rawValue.(string)
		if !ok {
			return ""
		}

		result = fmt.Sprintf("\"%s\"", strValue)
	case reflect.Slice:
		result = getSliceExampleString(rawValue)
	case reflect.Map:
		// TODO?
		return "object"
	default:
		result = ""
	}

	return result
}

// TODO: have a max number of included values
func getSliceExampleString(rawValue any) string {
	var (
		val    = reflect.ValueOf(rawValue)
		result string
	)

	if !val.IsValid() {
		return "nil"
	}

	result = "["

	for i := 0; i < val.Len(); i++ {
		itemVal := val.Index(i)
		itemRaw := itemVal.Interface()

		result += fmt.Sprintf("%s, ", getExampleString(itemRaw))
	}

	result = strings.TrimSuffix(result, ", ")
	result += "]"

	return result
}
