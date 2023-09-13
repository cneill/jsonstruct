package jsonstruct_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cneill/jsonstruct"
)

//nolint:gochecknoglobals // can't make it a const so what do
var (
	bigInt, _      = (&big.Int{}).SetString("9223372036854775808", 10)
	bigFloat, _, _ = (&big.Float{}).Parse("1.79769313486231570814527423731704356798070e309", 10)
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
		{"big_int", bigInt, "*big.Int"},
		{"float", float64(1.23), "float64"},
		{"big_float", bigFloat, "*big.Float"},
		{"string", "test", "string"},
		{"struct", jsonstruct.New(), "*Struct"},
		{"null", nil, "*json.RawMessage"},
		{"bool_slice", []bool{true, false, true}, "[]bool"},
		{"int_slice", []int64{1, 2, 3}, "[]int64"},
		{"float_slice", []float64{1.1, 1.2, 1.3}, "[]float64"},
		{"string_slice", []string{"1", "2", "3"}, "[]string"},
		{"garbage_slice", []any{1, "1", 1.0}, "[]*json.RawMessage"},
		{"any_bool_slice", []any{true, false, true}, "[]bool"},
		{"any_int_slice", []any{int64(1), int64(2), int64(3)}, "[]int64"},
		{"any_float_slice", []any{1.0, 2.0, 3.0}, "[]float64"},
		{"any_string_slice", []any{"1", "2", "3"}, "[]string"},
		{"null_slice", []any{nil, nil}, "[]*json.RawMessage"},
		{"structs", []any{jsonstruct.New()}, "[]*Structs"},
		{"nested_int64_slices", []any{[]int64{1, 2, 3}, []int64{4, 5, 6}}, "[][]int64"},
		{"nested_float64_slices", []any{[]float64{1, 2, 3}, []float64{4, 5, 6}}, "[][]float64"},
		{
			name:   "deeply_nested_int64_slices",
			input:  []any{[]any{[]any{[]any{[]int64{1, 2, 3}}}}},
			output: "[][][][][]int64",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			f := jsonstruct.NewField().SetName(test.name).SetValue(test.input)

			assert.Equal(t, test.output, f.Type())
		})
	}
}

func TestFieldValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  any
		output string
	}{
		{"bool", true, "true"},
		{"int", int64(123), "123"},
		{"float", float64(1.23), "1.230"},
		{"string", "test", "\"test\""},
		{"struct", jsonstruct.New(), ""},
		{"null", nil, "null"},
		{"bool_slice", []bool{true, false, true}, "[true, false, true]"},
		{"int_slice", []int64{1, 2, 3}, "[1, 2, 3]"},
		{"float_slice", []float64{1.1, 1.2, 1.3}, "[1.100, 1.200, 1.300]"},
		{"string_slice", []string{"1", "2", "3"}, "[\"1\", \"2\", \"3\"]"},
		{"garbage_slice", []any{1, "1", 1.0}, ""},
		{"any_bool_slice", []any{true, false, true}, "[true, false, true]"},
		{"any_int_slice", []any{int64(1), int64(2), int64(3)}, "[1, 2, 3]"},
		{"any_float_slice", []any{1.0, 2.0, 3.0}, "[1.000, 2.000, 3.000]"},
		{"any_string_slice", []any{"1", "2", "3"}, "[\"1\", \"2\", \"3\"]"},
		{"null_slice", []any{nil, nil}, ""},
		{"structs", []any{jsonstruct.New()}, ""},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			f := jsonstruct.NewField().SetName(test.name).SetValue(test.input)

			assert.Equal(t, test.output, f.Value())
		})
	}
}

func TestFieldEquals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		field1 *jsonstruct.Field
		field2 *jsonstruct.Field
		equal  bool
	}{
		{
			name:   "different_values",
			field1: jsonstruct.NewField().SetName("hah").SetValue(int64(1)),
			field2: jsonstruct.NewField().SetName("hah").SetValue(int64(2)),
			equal:  true,
		},
		{
			name:   "different_names",
			field1: jsonstruct.NewField().SetName("hah1").SetValue(int64(1)),
			field2: jsonstruct.NewField().SetName("hah2").SetValue(int64(1)),
			equal:  false,
		},
		{
			name:   "different_types",
			field1: jsonstruct.NewField().SetName("hah").SetValue(int64(1)),
			field2: jsonstruct.NewField().SetName("hah").SetValue(float64(2)),
			equal:  false,
		},
		{
			name:   "different_raw_names_same_go_names",
			field1: jsonstruct.NewField().SetName("hah_hah_hah").SetValue(int64(1)),
			field2: jsonstruct.NewField().SetName("hah-hah-hah").SetValue(int64(2)),
			equal:  false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.equal, test.field1.Equals(test.field2))
		})
	}
}
