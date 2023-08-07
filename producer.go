package jsonstruct

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
)

// Producer defines the options for how structs will be structured.
type Producer struct {
	// SortFields will sort the fields of the resulting struct alphabetically.
	SortFields bool
	// ValueComments will include a comment with every struct field with the value(s) received from the examples provided.
	ValueComments bool
	// Inline will use inline structs instead of creating new types for each JSON object detected.
	Inline bool
	// Name will override the name of the main struct.
	Name string
	// TODO: allow customization of intialisms
}

// Raw is a convenience type for map[string]any which is the default raw message type from encoding/json.
type Raw map[string]any

var skippedExamples = map[reflect.Kind]bool{
	reflect.Invalid: true,
	// reflect.Bool:    true,
	reflect.Pointer: true,
}

// GetFieldsFromRaw takes a map[string]any and returns the Fields represented.
func (p *Producer) GetFieldsFromRaw(input Raw) (Fields, error) {
	results := Fields{}

	for key, val := range input {
		kind := GetFieldKind(val)
		// TODO: mark the incorrect float64s int64s here and deal with their values?
		field := Field{
			RawName:  key,
			RawValue: val,
			Name:     GetNormalizedName(key),
			Value:    reflect.ValueOf(val),
			Kind:     kind,
		}

		// TODO: figure out a smarter way to deal with this...
		if _, ok := skippedExamples[kind]; !ok && p.ValueComments {
			if exStr := field.ExampleString(); exStr != "" {
				// field.Comments = []string{fmt.Sprintf("Ex: %s", exStr)}
				field.Comment = fmt.Sprintf("Ex: %s", exStr)
			}
		}

		if kind == reflect.Slice {
			skind, err := GetSliceKind(val)
			if err != nil {
				return nil, fmt.Errorf("failed to get slice kind: %w", err)
			}

			field.SliceKind = skind
			// TODO: figure out how to better deal with plurals here
			if field.SliceKind == reflect.Struct {
				field.StructType = GetNormalizedName(key)
				// what we actually have here is a slice of struct, so we need to step through each of the provided objects and
				// aggregate the fields into one object
				child, err := p.structFromSlice(field.StructType, val)
				if err != nil {
					return nil, err
				}

				field.Child = child
			}
		} else if kind == reflect.Map {
			field.StructType = GetNormalizedName(key)
			child, err := p.structFromRaw(field.StructType, Raw(val.(map[string]any)))
			if err != nil {
				return nil, err
			}
			field.Child = child
		}

		results = append(results, field)
	}

	return results, nil
}

// structFromRaw returns a JSONStruct constructed from the provided name and Raw object.
func (p *Producer) structFromRaw(name string, raw Raw) (*JSONStruct, error) {
	fields, err := p.GetFieldsFromRaw(raw)
	if err != nil {
		return nil, err
	}

	js := &JSONStruct{
		Name:   name,
		Fields: fields,
	}

	return p.FormatStruct(js), nil
}

// StructsFromReader reads the contents of 'r' and returns any structs (or arrays of structs) it can derive from them.
func (p *Producer) StructsFromReader(name string, r io.Reader) ([]*JSONStruct, error) {
	results := []*JSONStruct{}

	decoder := json.NewDecoder(r)
	decoder.UseNumber()

	for i := 0; ; i++ {
		var (
			indexName = fmt.Sprintf("%s%d", name, i)
			jsonRaw   = json.RawMessage{}
		)

		if err := decoder.Decode(&jsonRaw); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}

		var (
			err        error
			tempResult *JSONStruct
		)

		switch jsonRaw[0] {
		case '{':
			tempResult, err = p.parseObject(indexName, jsonRaw)
		case '[':
			tempResult, err = p.parseArray(indexName, jsonRaw)
		default:
			return nil, fmt.Errorf("expecting either array or object, invalid character '%c' looking for beginning of value", jsonRaw[0])
		}

		if err != nil {
			return nil, err
		}

		results = append(results, tempResult)
	}

	return results, nil
}

func (p *Producer) parseObject(name string, input json.RawMessage) (*JSONStruct, error) {
	raw := Raw{}
	if err := json.Unmarshal(input, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse object: %w", err)
	}

	return p.structFromRaw(name, raw)
}

func (p *Producer) parseArray(name string, input json.RawMessage) (*JSONStruct, error) {
	raws := []Raw{}
	if err := json.Unmarshal(input, &raws); err != nil {
		return nil, fmt.Errorf("failed to parse array: %w", err)
	}

	return p.structFromSlice(name, raws)
}

func (p *Producer) StructsFromBytes(name string, contents []byte) ([]*JSONStruct, error) {
	return p.StructsFromReader(name, bytes.NewReader(contents))
}

func (p *Producer) StructsFromExampleFiles(inputFiles ...string) ([]*JSONStruct, error) {
	results := []*JSONStruct{}

	for _, inputFile := range inputFiles {
		contents, err := os.ReadFile(inputFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %q: %w", inputFile, err)
		}

		name := NameFromInputFile(inputFile)
		if p.Name != "" {
			name = p.Name
		}

		structs, err := p.StructsFromBytes(name, contents)
		if err != nil {
			return nil, fmt.Errorf("failed to parse file %q: %w", inputFile, err)
		}

		results = append(results, structs...)
	}

	return results, nil
}

// structFromSlice looks at a slice of some type and returns a JSONStruct based on the values contained therein.
func (p *Producer) structFromSlice(name string, value any) (*JSONStruct, error) {
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
			return nil, fmt.Errorf("got a slice item that was not a map[string]interface - not a struct (%T)", iface)
		}

		js, err := p.structFromRaw(name, raw)
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

// FormatStruct formats a JSONStruct based on the options configured on the Producer.
func (p *Producer) FormatStruct(js *JSONStruct) *JSONStruct {
	if p.SortFields {
		js.Sort()
	}

	return js
}
