package jsonstruct_test

import (
	"testing"

	"github.com/cneill/jsonstruct"

	"github.com/stretchr/testify/assert"
)

func TestFormatString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *jsonstruct.JSONStruct
		expected string
	}{
		{
			name: "simple",
			input: jsonstruct.New().AddFields(
				jsonstruct.NewField().SetName("a").SetValue(int64(1)),
			),
			expected: "type Simple struct {\n\tA int64 `json:\"a\"`\n}",
		},
		// TODO: add more as fixture files
	}

	f, _ := jsonstruct.NewFormatter(&jsonstruct.FormatterOptions{})

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			test.input.SetName(jsonstruct.GetGoName(test.name))
			output, err := f.FormatString(test.input)
			assert.Nil(t, err)
			assert.Equal(t, test.expected, output)
		})
	}
}
