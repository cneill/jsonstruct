package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

var testObject = `{
	"test": "test",
	"Number": 1.5432,
	"NumberTwo": 1.0,
	"numberThree": 1,
	"NumberFour": 69420,
	"Nested": {
		"test": "ohno"
	}
}`

var testArray = `[
	{
		"test": "test"
	},
	{
		"test": "test2"
	}
]`

var (
	l *slog.Logger

	ErrDataEmpty = errors.New("data provided was empty")
	ErrNotObject = errors.New("received data was not a JSON object")
	ErrNotArray  = errors.New("received data was not a JSON array")
)

type JSONStruct struct {
	Name   string
	Fields Fields
}

// UnmarshalJSON takes one JSON object's bytes as input and marshals it into a JSONStruct object.
func (j *JSONStruct) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return ErrDataEmpty
	}

	if data[0] != '{' {
		return ErrNotObject
	}

	l.Debug(fmt.Sprintf("%s\n", data))

	decoder := json.NewDecoder(bytes.NewReader(data))
	// make sure we get the actual string so we can decide if it's an int64 or a float64
	decoder.UseNumber()

	// state of the parsing
	delimStack := []string{}
	keys := []string{}

	for {
		t, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return fmt.Errorf("failed to get token: %w", err)
		}

		switch val := t.(type) {
		case json.Delim:
			l.Debug("got delim", "delim", val)
			if val == '{' {
				if len(delimStack) == 0 {
					l.Debug("we're at the start of the first object")
					delimStack = append(delimStack, "{")
					// we're looking for keys now
					key, err := getKey(decoder)
					if err != nil {
						return fmt.Errorf("invalid key: %w", err)
					}
					l.Debug("got key", "key", key)

					keys = append(keys, key)

				}
				// got an object
			} else if val == '[' {
				// got an array
			}
		case bool:
			l.Debug("got bool", "bool", val)
		case json.Number:
			l.Debug("got json.Number", "json.Number", val)
			if strings.Contains(string(val), ".") {
				f, err := val.Float64()
				if err != nil {
					return fmt.Errorf("failed to parse float64: %w", err)
				}

				l.Debug("got float", "float", f)
			} else {
				i, err := val.Int64()
				if err != nil {
					return fmt.Errorf("failed to parse int64: %w", err)
				}

				l.Debug("got int", "int", i)
			}
		case string:
			l.Debug("got string", "string", val)
		default:
			l.Debug("got null", "nil", val)
		}
	}

	return nil
}

func getKey(decoder *json.Decoder) (string, error) {
	t, err := decoder.Token()
	if err != nil {
		return "", fmt.Errorf("failed to parse object key: %w", err)
	}

	val, ok := t.(string)
	if !ok {
		return "", fmt.Errorf("object key was not a string")
	}

	return val, nil
}

func handleArray(decoder *json.Decoder) {
}

type JSONStructs []*JSONStruct

func (j *JSONStructs) UnmarshalJSON(data []byte) error {
	results := JSONStructs{}

	if len(data) == 0 {
		return ErrDataEmpty
	}

	if data[0] != '[' {
		return ErrNotArray
	}

	l.Debug("got an array of some kind...")

	rawStructs := []json.RawMessage{}
	if err := json.Unmarshal(data, &rawStructs); err != nil {
		return fmt.Errorf("failed to unmarshal JSON array: %w", err)
	}

	l.Debug("got array items", "count", len(rawStructs))

	for _, rawStruct := range rawStructs {
		jsonStruct := &JSONStruct{}
		if err := json.Unmarshal([]byte(rawStruct), jsonStruct); err != nil {
			return err
		}

		results = append(results, jsonStruct)
	}

	*j = results

	return nil
}

type Field struct {
	Name string
	Type string
	Tag  string
}

type Fields []*Field

type Formatter struct{}

func parseBytes(data []byte) (JSONStructs, error) {
	if len(data) == 0 {
		l.Debug("no content")

		return nil, nil
	}

	switch data[0] {
	case '{':
		l.Debug("got an object")

		result, err := parseObjectBytes(data)
		if err != nil {
			return nil, err
		}

		return JSONStructs{result}, nil
	case '[':
		l.Debug("got an array")

		return parseArrayBytes(data)
	default:
		return nil, fmt.Errorf("invalid first character of JSON data")
	}
}

func parseObjectBytes(data []byte) (*JSONStruct, error) {
	result := &JSONStruct{}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON object: %w", err)
	}

	return result, nil
}

func parseArrayBytes(data []byte) ([]*JSONStruct, error) {
	results := JSONStructs{}

	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("failed to parse JSON array: %w", err)
	}

	return results, nil
}

/*
- make sure the JSON is well-formed
- parse it into a stably-ordered map OR a slice of maps
-
*/

func run() error {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	l = slog.New(h)

	j, err := parseBytes([]byte(testObject))
	if err != nil {
		return err
	}

	for _, result := range j {
		l.Debug(fmt.Sprintf("%+v\n", result))
	}

	// j, err = parseBytes([]byte(testArray))
	// if err != nil {
	// 	return err
	// }

	// for _, result := range j {
	// 	l.Debug(fmt.Sprintf("%+v\n", result))
	// }

	return nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
