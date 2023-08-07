package jsonstruct_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cneill/jsonstruct"
)

func TestGetNormalizedName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		output string
	}{
		{"separators1", "this_is-a.test_name", "ThisIsATestName"},
		{"separators2", "remote_-.URL", "RemoteURL"},
		{"weirdcase", "ThiSiSaNumber2", "ThiSiSaNumber2"},
		{"separators_with_initialism", "This_Is_An_ID", "ThisIsAnID"},
		{"underscore_start", "_underscored", "Underscored"},
		{"garbage_characters", "($@%)@$%)(@", ""},
		{"garbage_characters_with_content", "@)(#$)@(#$)@#($garbage@#)$@)#($@)#($", "Garbage"},
		{"garbage_separator", "@t@e@s@t", "Test"},
		{"spaces", "       spaces", "Spaces"},
		{"multiple_spaces", "here are some spaces", "HereAreSomeSpaces"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			output := jsonstruct.GetNormalizedName(test.input)
			assert.Equal(t, output, test.output)
		})
	}
}

func TestGetSliceKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  any
		output reflect.Kind
	}{
		{"floats", []float64{1.5, 2.0, 5.0}, reflect.Float64},
		{"ints", []int{1, 2, 3}, reflect.Int64},
		{"floats_any", []any{1.0, 2.0, 3.0}, reflect.Float64},
		{"ints_any", []any{1, 2, 3}, reflect.Int64},
		{"nested_any", []any{map[string]string{"test": "test"}}, reflect.Struct}, // maps => struct
		{"empty_ints", []int64{}, reflect.Invalid},                               // empty slices of any kind => Invalid
		{"empty_any", []any{}, reflect.Invalid},                                  // empty slices of any kind => Invalid
		{"random1", []any{1, "b", 3.0}, reflect.Invalid},                         // slices with varying kind => Invalid
		{"random2", []any{"test", struct{}{}, 1.5, nil}, reflect.Invalid},        // slices with varying kind => Invalid
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			output, err := jsonstruct.GetSliceKind(test.input)
			assert.Nil(t, err)
			assert.Equal(t, test.output, output)
		})
	}

	if _, err := jsonstruct.GetSliceKind(struct{}{}); err == nil {
		t.Errorf("should throw error when non-slice is passed to GetSliceKind")
	}
}

func TestNameFromInputFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		output string
	}{
		{"underscores", "test_file.json", "TestFile"},
		{"uncapitalized_camel", "garbageFile", "GarbageFile"},
		{"capitalized", "CapitalizedFileName.JSON", "CapitalizedFileName"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			output := jsonstruct.NameFromInputFile(test.input)
			assert.Equal(t, output, test.output)
		})
	}
}
