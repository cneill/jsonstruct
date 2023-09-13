package jsonstruct

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"strconv"
	"strings"
)

var ErrOverflow = errors.New("provided number was too large")

type Parser struct {
	log      *slog.Logger
	decoder  *json.Decoder
	current  any
	previous any
	buf      any
	started  bool
}

func NewParser(input io.Reader, logger *slog.Logger) *Parser {
	decoder := json.NewDecoder(input)
	decoder.UseNumber()

	return &Parser{
		decoder: decoder,
		log:     logger,
	}
}

func (p *Parser) Start() (JSONStructs, error) {
	results := JSONStructs{}

	for i := 0; ; i++ {
		first, err := p.peek()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to start parser: %w", err)
		}

		p.log.Debug("successfully read first token")

		delim, ok := first.(json.Delim)
		if !ok {
			return nil, fmt.Errorf("expecting to start with a json.Delim, got %+v", first)
		}

		p.log.Debug("successfully read a JSON delimiter", "delim", delim)

		switch delim {
		case '{':
			js, err := p.parseObject()
			if err != nil {
				return nil, fmt.Errorf("failed to parse object: %w", err)
			}

			results = append(results, js)
		case '[':
			jsRaw, err := p.parseArray()
			if err != nil {
				return nil, fmt.Errorf("failed to parse array: %w", err)
			}

			js := getSliceStruct(jsRaw)

			results = append(results, js)
		}
	}

	// return nil, fmt.Errorf("invalid starting token: %+v", first)
	return results, nil
}

func (p *Parser) next() error {
	p.started = true

	// we have a previously-buffered token
	if p.buf != nil {
		p.previous = p.current
		p.current = p.buf
		p.buf = nil

		p.log.Debug("setting current token to buffered token", "buf", p.current)

		return nil
	}

	token, err := p.decoder.Token()
	if err != nil {
		return fmt.Errorf("failed to get next token: %w", err)
	}

	p.previous = p.current
	p.current = token

	p.log.Debug("got next token", "current", p.current)

	return nil
}

func (p *Parser) backup() error {
	if !p.started {
		return fmt.Errorf("no tokens to back up")
	}

	// buffer the current token, pick it up with the subsequent next() call, back current up to previous
	p.buf, p.current = p.current, p.previous

	p.log.Debug("putting current token in buffer, setting current to previous token", "current", p.current, "buf", p.buf)

	return nil
}

func (p *Parser) peek() (any, error) {
	if err := p.next(); err != nil {
		return nil, fmt.Errorf("failed to get next token when peeking: %w", err)
	}

	result := p.current

	if err := p.backup(); err != nil {
		return nil, fmt.Errorf("failed to back up while peeking: %w", err)
	}

	return result, nil
}

func (p *Parser) parseObject() (*JSONStruct, error) {
	result := New()

	if err := p.next(); err != nil {
		return result, err
	}

	if err := p.parseDelim('{'); err != nil {
		return result, fmt.Errorf("failed to get start of object: %w", err)
	}

	for p.decoder.More() {
		if err := p.next(); err != nil {
			return result, err
		}

		key, err := p.parseString()
		if err != nil {
			return result, fmt.Errorf("failed to parse key: %w", err)
		}

		p.log.Debug("parsed key", "key", key)

		val, err := p.parseValue()
		if err != nil {
			return result, fmt.Errorf("failed to parse value: %w", err)
		}

		field := (&Field{}).SetName(key).SetValue(val)

		result.AddFields(field)
	}

	if err := p.next(); err != nil {
		return result, fmt.Errorf("failed to get next token (expecting '}')")
	}

	if err := p.parseDelim('}'); err != nil {
		return result, fmt.Errorf("failed to get start of object: %w", err)
	}

	return result, nil
}

func (p *Parser) parseDelim(delim rune) error {
	delimToken, ok := p.current.(json.Delim)
	if !ok {
		return fmt.Errorf("not delim")
	}

	p.log.Debug("got json.Delim", "delim", delimToken)

	if rune(delimToken) != delim {
		return fmt.Errorf("not start of object (expecting '%c')", delim)
	}

	return nil
}

func (p *Parser) parseBool() (bool, error) {
	tokenBool, ok := p.current.(bool)
	if !ok {
		return false, fmt.Errorf("not bool")
	}

	p.log.Debug("got bool", "bool", tokenBool)

	return tokenBool, nil
}

func (p *Parser) parseString() (string, error) {
	tokenStr, ok := p.current.(string)
	if !ok {
		return "", fmt.Errorf("not string")
	}

	p.log.Debug("got string", "string", tokenStr)

	return tokenStr, nil
}

func (p *Parser) parseNumber() (any, error) {
	tokenNumber, ok := p.current.(json.Number)
	if !ok {
		return 0, fmt.Errorf("not json.Number")
	}

	numberStr := string(tokenNumber)

	p.log.Debug("got Number", "json.Number", tokenNumber)

	if strings.Contains(numberStr, ".") || strings.Contains(numberStr, "e") || strings.Contains(numberStr, "E") {
		// float of some kind
		if _, err := strconv.ParseFloat(numberStr, 64); errors.Is(err, strconv.ErrRange) {
			return p.parseBigFloat()
		}

		return p.parseFloat64()
	}

	if _, err := strconv.ParseInt(numberStr, 10, 64); errors.Is(err, strconv.ErrRange) {
		return p.parseBigInt()
	}

	return p.parseInt64()
}

func (p *Parser) parseInt64() (int64, error) {
	tokenNumber, ok := p.current.(json.Number)
	if !ok {
		return 0, fmt.Errorf("not json.Number")
	}

	p.log.Debug("got Number", "json.Number", tokenNumber)

	tokenInt, err := tokenNumber.Int64()
	if err != nil {
		return 0, fmt.Errorf("not int64: %w", err)
	}

	p.log.Debug("got int", "int", tokenInt)

	return tokenInt, nil
}

func (p *Parser) parseBigInt() (*big.Int, error) {
	tokenNumber, ok := p.current.(json.Number)
	if !ok {
		return big.NewInt(0), fmt.Errorf("not json.Number")
	}

	p.log.Debug("got Number", "json.Number", tokenNumber)

	tokenBigInt, _ := (&big.Int{}).SetString(string(tokenNumber), 10)

	p.log.Debug("got *big.Int float", "int", tokenBigInt)

	return tokenBigInt, nil
}

func (p *Parser) parseFloat64() (float64, error) {
	tokenNumber, ok := p.current.(json.Number)
	if !ok {
		return 0, fmt.Errorf("not json.Number")
	}

	p.log.Debug("got Number", "json.Number", tokenNumber)

	tokenFloat, err := tokenNumber.Float64()
	if err != nil {
		return 0, fmt.Errorf("not float64: %w", err)
	}

	p.log.Debug("got float", "float", tokenFloat)

	return tokenFloat, nil
}

func (p *Parser) parseBigFloat() (*big.Float, error) {
	tokenNumber, ok := p.current.(json.Number)
	if !ok {
		return big.NewFloat(0.0), fmt.Errorf("not json.Number")
	}

	p.log.Debug("got Number", "json.Number", tokenNumber)

	tokenBigFloat, _, err := (&big.Float{}).Parse(string(tokenNumber), 10)
	if err != nil {
		return big.NewFloat(0.0), fmt.Errorf("not big.Float: %w: %w", ErrOverflow, err)
	}

	p.log.Debug("got *big.Float", "float", tokenBigFloat)

	return tokenBigFloat, nil
}

func (p *Parser) parseArray() ([]any, error) {
	result := []any{}

	if err := p.next(); err != nil {
		return nil, fmt.Errorf("failed to get next token: %w", err)
	}

	if err := p.parseDelim('['); err != nil {
		return nil, fmt.Errorf("failed to get start of array: %w", err)
	}

	for p.decoder.More() {
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		result = append(result, val)
	}

	t, err := p.decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get next token: %w", err)
	}

	p.current = t

	if err := p.parseDelim(']'); err != nil {
		return nil, fmt.Errorf("failed to get start of array: %w", err)
	}

	return result, nil
}

func (p *Parser) parseValue() (any, error) {
	token, err := p.decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get next token: %w", err)
	}

	p.current = token

	switch val := token.(type) {
	case json.Delim:
		p.log.Debug("got a delim, trying to back up to parse either object or array", "delim", val)

		if err := p.backup(); err != nil {
			return nil, err
		}

		if val == '[' {
			return p.parseArray()
		} else if val == '{' {
			return p.parseObject()
		}
	case bool:
		return p.parseBool()
	case string:
		return p.parseString()
	case json.Number:
		return p.parseNumber()
	default:
		// got null
		//nolint:nilnil // I literally want to return nil
		return nil, nil
	}

	return nil, fmt.Errorf("should not reach this point")
}
