package jsonstruct

import (
	"fmt"
	"math"
	"os"
	"path"
	"reflect"
	"strings"
)

// GetNormalizedName normalizes a Field name from its JSON counterpart - removing "-", "_", ".", capitalizing properly.
func GetNormalizedName(key string) string {
	key = strings.ReplaceAll(key, "-", " ")
	key = strings.ReplaceAll(key, "_", " ")
	key = strings.ReplaceAll(key, ".", " ")
	words := strings.Split(key, " ")
	temp := []string{}

	for _, word := range words {
		tmpWord := strings.ToUpper(word)
		if _, ok := commonInitialisms[tmpWord]; ok {
			word = strings.ToUpper(tmpWord)
		}

		temp = append(temp, word)
	}

	key = strings.Join(temp, " ")
	// TODO: decide whether golang.org/x/text/cases is worth it
	key = strings.Title(key)
	key = strings.ReplaceAll(key, " ", "")

	return key
}

// GetSliceKind takes a slice object and returns the Kind of slice represented - defaults to reflect.String if unknown.
// TODO: handle cases where an array contains more than one type (e.g. ["a", "b", ["c"]])
func GetSliceKind(value any) (reflect.Kind, error) {
	var (
		typeOf = reflect.TypeOf(value)
		kind   = typeOf.Kind()
	)

	if kind != reflect.Slice {
		return reflect.Invalid, fmt.Errorf("must provide a value with Kind == Slice")
	}

	valOf := reflect.ValueOf(value)

	if valOf.Len() == 0 {
		fmt.Fprintf(os.Stderr, "Got an empty array, defaulting to string")

		return reflect.String, nil
	}

	elemVal := valOf.Index(0)
	iface := elemVal.Interface()

	switch iface.(type) {
	case int, int8, int16, int32, int64:
		return reflect.Int64, nil
	case float32, float64:
		return reflect.Float64, nil
	case string:
		return reflect.String, nil
	}

	elemType := reflect.TypeOf(iface)
	elemKind := elemType.Kind()

	if elemKind == reflect.Map {
		return reflect.Struct, nil
	}

	fmt.Fprintf(os.Stderr, "Not sure what to do with an array of %s... defaulting to string\n", elemKind)

	return reflect.String, nil
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
	ext := path.Ext(fName)
	name := strings.TrimSuffix(inputFile, ext)
	name = GetNormalizedName(name)

	return name
}

// CanBeInt64 checks whether a float value can be converted into an int64 without a loss of precision. Helps find e.g. IDs,
// counts, and so on.
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
		result = fmt.Sprintf("%d", rawValue.(int64))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		result = fmt.Sprintf("%d", rawValue.(uint64))
	case reflect.Bool:
		result = fmt.Sprintf("%t", rawValue.(bool))
	case reflect.Float32, reflect.Float64:
		f := rawValue.(float64)
		if CanBeInt64(f) {
			result = fmt.Sprintf("%d", int64(f))
		} else {
			result = fmt.Sprintf("%.2f", rawValue.(float64))
		}
	case reflect.String:
		result = fmt.Sprintf("\"%s\"", rawValue.(string))
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
