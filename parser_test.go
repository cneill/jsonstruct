package jsonstruct_test

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/cneill/jsonstruct"

	"github.com/stretchr/testify/assert"
)

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
			name:  "nested",
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
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, 1, len(structs))
			assert.Equal(t, len(test.expected.Fields), len(structs[0].Fields))

			for i := 0; i < len(test.expected.Fields); i++ {
				assert.Equal(t, test.expected.Fields[i], structs[0].Fields[i])
			}
		})
	}
}
