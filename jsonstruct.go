package jsonstruct

// JSONStruct contains the raw information about a JSON object to be rendered as a Go struct.
type JSONStruct struct {
	Name   string
	Fields Fields
}

// JSONStructs is a convenience type for a slice of JSONStruct structs.
type JSONStructs []*JSONStruct

// UnmarshalJSON takes one JSON object's bytes as input and marshals it into a JSONStruct object.
/*
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
*/

/*
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
*/
