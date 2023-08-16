package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

var eof = rune(0x00)

type objectParser struct {
	d   *json.Decoder
	buf any
}

func (o *objectParser) parseObject() (*JSONStruct, error) {
	result := &JSONStruct{
		Fields: Fields{},
	}

	if o.buf == nil {
		t, err := o.d.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to get next token (expecting '{'): %w", err)
		}

		o.buf = t
	}

	if err := o.parseDelim('{'); err != nil {
		return nil, fmt.Errorf("failed to get start of object: %w", err)
	}

	for o.d.More() {
		field := &Field{}

		t, err := o.d.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to get key token: %w", err)
		}

		o.buf = t

		key, err := o.parseString()
		if err != nil {
			return nil, fmt.Errorf("failed to parse key: %w", err)
		}

		field.Name = GetGoName(key)
		field.Tag = key

		val, err := o.parseValue()
		if err != nil {
			return nil, fmt.Errorf("failed to parse value: %w", err)
		}

		field.StrValue = stringValue(val)
		field.RawValue = val

		result.Fields = append(result.Fields, field)
	}

	t, err := o.d.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get next token (expecting '}'): %w", err)
	}

	o.buf = t

	if err := o.parseDelim('}'); err != nil {
		return nil, fmt.Errorf("failed to get start of object: %w", err)
	}

	return result, nil
}

func (o *objectParser) parseDelim(delim rune) error {
	delimToken, ok := o.buf.(json.Delim)
	if !ok {
		return fmt.Errorf("not delim")
	}

	l.Debug("got json.Delim", "delim", delimToken)

	if rune(delimToken) != delim {
		return fmt.Errorf("not start of object (expecting '%c')", delim)
	}

	return nil
}

func (o *objectParser) parseBool() (bool, error) {
	tokenBool, ok := o.buf.(bool)
	if !ok {
		return false, fmt.Errorf("not bool")
	}

	l.Debug("got bool", "bool", tokenBool)

	return tokenBool, nil
}

func (o *objectParser) parseString() (string, error) {
	tokenStr, ok := o.buf.(string)
	if !ok {
		return "", fmt.Errorf("not string")
	}

	l.Debug("got string", "string", tokenStr)

	return tokenStr, nil
}

func (o *objectParser) parseInt64() (int64, error) {
	tokenNumber, ok := o.buf.(json.Number)
	if !ok {
		return 0, fmt.Errorf("not json.Number")
	}

	l.Debug("got Number", "json.Number", tokenNumber)

	tokenInt, err := tokenNumber.Int64()
	if err != nil {
		return 0, fmt.Errorf("not int64: %w", err)
	}

	l.Debug("got int", "int", tokenInt)

	return tokenInt, nil
}

func (o *objectParser) parseFloat64() (float64, error) {
	tokenNumber, ok := o.buf.(json.Number)
	if !ok {
		return 0, fmt.Errorf("not json.Number")
	}

	l.Debug("got Number", "json.Number", tokenNumber)

	tokenFloat, err := tokenNumber.Float64()
	if err != nil {
		return 0, fmt.Errorf("not float64: %w", err)
	}

	l.Debug("got float", "float", tokenFloat)

	return tokenFloat, nil
}

func (o *objectParser) parseArray() ([]any, error) {
	result := []any{}

	if err := o.parseDelim('['); err != nil {
		return nil, fmt.Errorf("failed to get start of array: %w", err)
	}

	for o.d.More() {
		val, err := o.parseValue()
		if err != nil {
			return nil, err
		}

		result = append(result, val)
	}

	t, err := o.d.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get next token: %w", err)
	}

	o.buf = t

	if err := o.parseDelim(']'); err != nil {
		return nil, fmt.Errorf("failed to get start of array: %w", err)
	}

	return result, nil
}

func (o *objectParser) parseValue() (any, error) {
	token, err := o.d.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get next token: %w", err)
	}

	o.buf = token

	switch val := token.(type) {
	case json.Delim:
		if val == '[' {
			return o.parseArray()
		} else if val == '{' {
			return o.parseObject()
		}
	case bool:
		return o.parseBool()
	case string:
		return o.parseString()
	case json.Number:
		if strings.Contains(string(val), ".") {
			return o.parseFloat64()
		}

		return o.parseInt64()
	default:
		// got null
		return nil, nil
	}

	return nil, fmt.Errorf("should not reach this point")
}
