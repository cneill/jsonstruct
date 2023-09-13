package jsonstruct_test

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/cneill/jsonstruct"

	"github.com/stretchr/testify/assert"
)

//nolint:funlen // it's a table-driven test :shrug:
func TestParser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected *jsonstruct.JSONStruct
		errors   bool
	}{
		{
			name:  "simple",
			input: `{"a": 1}`,
			expected: (&jsonstruct.JSONStruct{}).AddFields(
				(&jsonstruct.Field{}).SetName("a").SetValue(int64(1)),
			),
			errors: false,
		},
		{
			name:  "nested_object",
			input: `{"a": {"b": 1}}`,
			expected: jsonstruct.New().SetName("").AddFields(
				(&jsonstruct.Field{}).SetName("a").SetValue(
					jsonstruct.New().AddFields(
						(&jsonstruct.Field{}).SetName("b").SetValue(int64(1)),
					),
				),
			),
			errors: false,
		},
		{
			name:  "arrays",
			input: `{"a": [1.0, 2.0, 3.0], "b": [true, false], "c": ["a", "B", "c"]}`,
			expected: jsonstruct.New().SetName("").AddFields(
				jsonstruct.NewField().SetName("a").SetValue([]any{float64(1), float64(2), float64(3)}),
				jsonstruct.NewField().SetName("b").SetValue([]any{true, false}),
				jsonstruct.NewField().SetName("c").SetValue([]any{"a", "B", "c"}),
			),
			errors: false,
		},
		// TODO: make this work by handling arrays of objects differently and not just spewing every example
		/*
			{
				name:  "array_of_objects",
				input: `[{"test": 1}, {"test": 2}, {"test": 3}]`,
				expected: jsonstruct.New().SetName("").AddFields(
					jsonstruct.NewField().SetName("test").SetValue(int64(1)),
				),
				errors: false,
			},
		*/
		{
			name:   "int_key",
			input:  `{1: 2}`,
			errors: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := strings.NewReader(test.input)
			p := jsonstruct.NewParser(r, slog.Default())
			structs, err := p.Start()
			if test.errors {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)

			assert.Equal(t, 1, len(structs))
			assert.Equal(t, len(test.expected.Fields()), len(structs[0].Fields()))

			expectedFields := test.expected.Fields()
			outputFields := structs[0].Fields()

			if len(expectedFields) != len(outputFields) {
				t.FailNow()
			}

			for i := 0; i < len(expectedFields); i++ {
				assert.Equal(t, expectedFields[i], outputFields[i])
			}
		})
	}
}

func FuzzParser(f *testing.F) {
	seeds := []string{
		`{"a": 1}`,
		`{"a": [], "b": 2, "c": {"d": 1}}`,
		`{"test": {"test": {"test": {"test": "test"}}}}`,
		`[{"test": "test", "test2": {"test": "test", "test2": [{"test": "test"}]}}]`,
		`{"test": 1e10000000000}`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		raw := &json.RawMessage{}

		if len(input) == 0 || input[0] != '{' {
			// only take objects for now
			return
		}

		if err := json.Unmarshal([]byte(input), raw); err != nil {
			// do we want to test valid JSON?
			return
		}

		r := strings.NewReader(input)
		p := jsonstruct.NewParser(r, slog.Default())
		_, err := p.Start()
		if errors.Is(err, jsonstruct.ErrOverflow) {
			return
		}
		assert.Nil(t, err)
	})
}
