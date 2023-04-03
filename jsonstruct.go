package jsonstruct

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
)

// Raw is a convenience type for map[string]interface{}, which is the default raw message type from encoding/json.
type Raw map[string]interface{}

// Producer defines the options for how structs will be structured.
type Producer struct {
	// SortFields will sort the fields of the resulting struct alphabetically.
	SortFields bool
	// VerboseValueComments will include a comment above every struct field with the value(s) received from the examples provided.
	VerboseValueComments bool
}

// JSONStruct is a struct produced from examples.
type JSONStruct struct {
	Name   string
	Fields Fields
}

func (j *JSONStruct) String() string {
	result := fmt.Sprintf("type %s struct {\n", j.Name)
	for _, field := range j.Fields {
		result += "\t" + field.String() + "\n"
	}
	result += "}"

	for _, field := range j.Fields {
		if field.Child != nil {
			result += fmt.Sprintf("\n\n%s", field.Child.String())
		}
	}

	return result
}

// Equals compares 2 JSONStruct objects and returns true if they're equal
// TODO: FINISH
func (j *JSONStruct) Equals(compare *JSONStruct) bool {
	if len(j.Fields) != len(compare.Fields) {
		return false
	}

	return true
}

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
	var result string

	kind := f.Kind.String()
	if f.Kind == reflect.Map {
		if f.StructType == "" {
			kind = "map[string]interface{}"
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

	if len(f.Comments) == 1 {
		result = fmt.Sprintf("%s %s %s // %s", f.Name, kind, f.Tag(), f.Comments[0])
	} else if len(f.Comments) > 1 {
		for _, comment := range f.Comments {
			result += fmt.Sprintf("// %s\n", comment)
		}
		result += fmt.Sprintf("%s %s %s", f.Name, kind, f.Tag())
	} else {
		result = fmt.Sprintf("%s %s %s", f.Name, kind, f.Tag())
	}

	return result
}

// Equals compares 2 Field objects to see if they are have the same fields.
func (f Field) Equals(compare Field) bool {
	if f.Name != compare.Name {
		return false
	} else if f.RawName != compare.RawName {
		return false
	} else if f.Kind != compare.Kind {
		return false
	} else if f.SliceKind != compare.SliceKind {
		return false
	} else if f.StructType != compare.StructType {
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
func GetSliceKind(value interface{}) reflect.Kind {
	typeOf := reflect.TypeOf(value)
	kind := typeOf.Kind()

	if kind != reflect.Slice {
		panic(fmt.Errorf("must provide a value with Kind == Slice"))
	}

	valOf := reflect.ValueOf(value)
	if valOf.Len() > 0 {
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
	} else {
		fmt.Fprintf(os.Stderr, "Got an empty array, defaulting to string")
	}

	return reflect.String
}

// GetFieldKind takes any object and returns the Kind represented.
func GetFieldKind(value interface{}) reflect.Kind {
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

// GetFieldsFromRaw takes a map[string]interface{} and returns the Fields represented.
func (p *Producer) GetFieldsFromRaw(input Raw) (Fields, error) {
	results := Fields{}

	for k, v := range input {
		kind := GetFieldKind(v)
		field := Field{
			Name:    GetNormalizedName(k),
			RawName: k,
			Kind:    kind,
		}
		if kind != reflect.Bool && p.VerboseValueComments {
			val := reflect.ValueOf(v)
			if val.String() != "" {
				receivedValueComment := fmt.Sprintf("Ex: %q", val.String())
				field.Comments = []string{receivedValueComment}
			}
		}

		if kind == reflect.Slice {
			field.SliceKind = GetSliceKind(v)
			// TODO: figure out how to better deal with plurals here
			if field.SliceKind == reflect.Struct {
				field.StructType = GetNormalizedName(k)
				// what we actually have here is a slice of struct, so we need to step through each of the provided objects and
				// aggregate the fields into one object
				child, err := p.StructFromSlice(field.StructType, v)
				if err != nil {
					return nil, err
				}
				field.Child = child
			}
		} else if kind == reflect.Map {
			field.StructType = GetNormalizedName(k)
			child, err := p.StructFromRaw(field.StructType, Raw(v.(map[string]interface{})))
			if err != nil {
				return nil, err
			}
			field.Child = child
		}

		results = append(results, field)
	}

	return results, nil
}

// NameFromInputFile strips the file path and extension, using GetNormalizedName to return a struct name.
func NameFromInputFile(inputFile string) string {
	_, fName := path.Split(inputFile)
	ext := path.Ext(fName)
	name := strings.TrimSuffix(inputFile, ext)
	name = GetNormalizedName(name)
	return name
}

// StructFromRaw returns a JSONStruct constructed from the provided name and Raw object.
func (p *Producer) StructFromRaw(name string, raw Raw) (*JSONStruct, error) {
	fields, err := p.GetFieldsFromRaw(raw)
	if err != nil {
		return nil, err
	}
	js := &JSONStruct{
		Name:   name,
		Fields: fields,
	}

	return js, nil
}

// StructFromBytes unmarshals either a JSON object or an array of JSON objects into a Raw object, and returns a JSONStruct.
func (p *Producer) StructFromBytes(name string, contents []byte) (*JSONStruct, error) {
	raw := Raw{}
	if err := json.Unmarshal(contents, &raw); err == nil {
		return p.StructFromRaw(name, raw)
	}

	raws := []Raw{}
	if err := json.Unmarshal(contents, &raws); err == nil {
		return p.StructFromSlice(name, raws)
	}

	return nil, fmt.Errorf("failed to parse as a JSON object or an array of JSON objects")
}

// StructFromExampleFile reads "inputFile", deriving a struct name from the file name and returning a JSONStruct.
func (p *Producer) StructFromExampleFile(inputFile string) (*JSONStruct, error) {
	contents, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	name := NameFromInputFile(inputFile)

	js, err := p.StructFromBytes(name, contents)
	if err != nil {
		return nil, err
	}

	js.Fields.Sort()

	return js, nil
}

// StructFromSlice looks at a slice of some type and returns a JSONStruct based on the values contained therein.
func (p *Producer) StructFromSlice(name string, value interface{}) (*JSONStruct, error) {
	typeOf := reflect.TypeOf(value)
	kind := typeOf.Kind()

	if kind != reflect.Slice {
		panic(fmt.Errorf("must provide a value with Kind == Slice"))
	}

	valOf := reflect.ValueOf(value)
	if valOf.Len() == 0 {
		return nil, fmt.Errorf("slice length was 0")
	}

	allFields := map[string]Fields{}

	for i := 0; i < valOf.Len(); i++ {
		elemVal := valOf.Index(i)
		iface := elemVal.Interface()
		var raw Raw
		if v, ok := iface.(map[string]interface{}); ok {
			raw = Raw(v)
		} else if v, ok := iface.(Raw); ok {
			raw = v
		} else {
			return nil, fmt.Errorf("got a slice item that was not a map[string]interface - not a struct")
		}
		js, err := p.StructFromRaw(name, raw)
		if err != nil {
			return nil, err
		}
		for _, field := range js.Fields {
			if _, ok := allFields[field.Name]; !ok {
				allFields[field.Name] = Fields{}
			}

			allFields[field.Name] = append(allFields[field.Name], field)
		}
	}

	js := &JSONStruct{
		Name:   name,
		Fields: Fields{},
	}

	for _, fieldSlice := range allFields {
		var field Field

		if len(fieldSlice) == 1 {
			field = fieldSlice[0]
		} else {
			first := fieldSlice[0]
			for i := 1; i < len(fieldSlice); i++ {
				if !first.Equals(fieldSlice[i]) {
					first.RawMessage = true
					break
				}
			}
			field = first
		}

		// we got the field in some but not all objects in the slice
		if len(fieldSlice) < valOf.Len() {
			field.OmitEmpty = true
		}

		js.Fields = append(js.Fields, field)
	}

	return js, nil
}
