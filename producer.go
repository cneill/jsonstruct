package jsonstruct

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
)

// Producer defines the options for how structs will be structured.
type Producer struct {
	// SortFields will sort the fields of the resulting struct alphabetically.
	SortFields bool
	// VerboseValueComments will include a comment above every struct field with the value(s) received from the examples provided.
	VerboseValueComments bool
	// Name will override the name of the main struct
	Name string
}

var skippedExamples = map[reflect.Kind]bool{
	reflect.Invalid: true,
	// reflect.Bool:    true,
	reflect.Pointer: true,
}

// GetFieldsFromRaw takes a map[string]any and returns the Fields represented.
func (p *Producer) GetFieldsFromRaw(input Raw) (Fields, error) {
	results := Fields{}

	for k, v := range input {
		kind := GetFieldKind(v)
		// TODO: mark the incorrect float64s int64s here and deal with their values?
		field := Field{
			RawName:  k,
			RawValue: v,
			Name:     GetNormalizedName(k),
			Value:    reflect.ValueOf(v),
			Kind:     kind,
		}

		// TODO: figure out a smarter way to deal with this...
		if _, ok := skippedExamples[kind]; !ok && p.VerboseValueComments {
			if exStr := field.ExampleString(); exStr != "" {
				field.Comments = []string{fmt.Sprintf("Ex: %s", exStr)}
			}
		}

		if kind == reflect.Slice {
			skind, err := GetSliceKind(v)
			if err != nil {
				return nil, fmt.Errorf("failed to get slice kind: %w", err)
			}

			field.SliceKind = skind
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
			child, err := p.StructFromRaw(field.StructType, Raw(v.(map[string]any)))
			if err != nil {
				return nil, err
			}
			field.Child = child
		}

		results = append(results, field)
	}

	return results, nil
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
	contents, err := os.ReadFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	name := NameFromInputFile(inputFile)
	if p.Name != "" {
		name = p.Name
	}

	js, err := p.StructFromBytes(name, contents)
	if err != nil {
		return nil, err
	}

	if p.SortFields {
		js.Sort()
	}

	return js, nil
}

// StructFromStdin reads stdin and returns a JSONStruct, or error.
func (p *Producer) StructFromStdin() (*JSONStruct, error) {
	contents, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	name := p.Name
	if name == "" {
		name = "STDIN"
	}

	js, err := p.StructFromBytes(name, contents)
	if err != nil {
		return nil, err
	}

	if p.SortFields {
		js.Sort()
	}

	return js, nil
}

// StructFromSlice looks at a slice of some type and returns a JSONStruct based on the values contained therein.
func (p *Producer) StructFromSlice(name string, value any) (*JSONStruct, error) {
	var (
		typeOf = reflect.TypeOf(value)
		kind   = typeOf.Kind()
	)

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
		if v, ok := iface.(map[string]any); ok {
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
