package jsonstruct

import (
	"fmt"
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
	key = strings.Title(key)
	key = strings.ReplaceAll(key, " ", "")
	return key
}

// GetSliceKind takes a slice object and returns the Kind of slice represented - defaults to reflect.String if unknown.
func GetSliceKind(value any) reflect.Kind {
	typeOf := reflect.TypeOf(value)
	kind := typeOf.Kind()

	if kind != reflect.Slice {
		panic(fmt.Errorf("must provide a value with Kind == Slice"))
	}

	valOf := reflect.ValueOf(value)

	if valOf.Len() == 0 {
		fmt.Fprintf(os.Stderr, "Got an empty array, defaulting to string")
		return reflect.String
	}

	elemVal := valOf.Index(0)
	iface := elemVal.Interface()
	switch iface.(type) {
	case int, int8, int16, int32, int64:
		return reflect.Int64
	case float32, float64:
		return reflect.Float64
	case string:
		return reflect.String
	}

	elemType := reflect.TypeOf(iface)
	elemKind := elemType.Kind()

	if elemKind == reflect.Map {
		return reflect.Struct
	}

	fmt.Fprintf(os.Stderr, "Not sure what to do with an array of %s... defaulting to string\n", elemKind)
	return reflect.String
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
