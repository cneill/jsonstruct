package jsonstruct_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cneill/jsonstruct"
)

func TestFieldType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  any
		output string
	}{
		{"bool", true, "bool"},
		{"int", int64(123), "int64"},
		{"float", float64(1.23), "float64"},
		{"string", "test", "string"},
		{"bool_slice", []bool{true, false, true}, "[]bool"},
		{"int_slice", []int{1, 2, 3}, "[]int"},
		{"float_slice", []float64{1.1, 1.2, 1.3}, "[]float64"},
		{"string_slice", []string{"1", "2", "3"}, "[]string"},
		{"garbage_slice", []any{1, "1", 1.0}, "[]*json.RawMessage"},
		{"any_bool_slice", []any{true, false, true}, "[]bool"},
		{"any_int_slice", []any{int64(1), int64(2), int64(3)}, "[]int64"},
		{"any_float_slice", []any{1.0, 2.0, 3.0}, "[]float64"},
		{"any_string_slice", []any{"1", "2", "3"}, "[]string"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			f := (&jsonstruct.Field{}).SetName(test.name).SetValue(test.input)

			assert.Equal(t, test.output, f.Type())
		})
	}
}
