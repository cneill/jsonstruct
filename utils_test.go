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
		input  string
		output string
	}{
		{"this_is-a.test_name", "ThisIsATestName"},
		{"remote_-.URL", "RemoteURL"},
		{"ThiSiSaNumber2", "ThiSiSaNumber2"},
		{"This_Is_An_ID", "ThisIsAnID"},
		{"_underscored", "Underscored"},
		{"($@%)@$%)(@", ""},
		{"@)(#$)@(#$)@#($garbage@#)$@)#($@)#($", "Garbage"},
		{"@t@e@s@t", "Test"},
		{"$aVariableName", "AVariableName"},
		{"       spaces", "Spaces"},
		{"here are some spaces", "HereAreSomeSpaces"},
	}

	for _, test := range tests {
		output := jsonstruct.GetNormalizedName(test.input)
		assert.Equal(t, output, test.output)
	}
}

type testStruct struct{}

func TestGetSliceKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input  any
		output reflect.Kind
	}{
		{[]float64{1.5}, reflect.Float64},
		{[]int{1}, reflect.Int64},
	}

	for _, test := range tests {
		output, err := jsonstruct.GetSliceKind(test.input)
		assert.Nil(t, err)
		assert.Equal(t, test.output, output)
	}
}

func TestNameFromInputFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input  string
		output string
	}{
		{"test_file.json", "TestFile"},
		{"garbagefile", "Garbagefile"},
		{"CapitalizedFileName.JSON", "CapitalizedFileName"},
	}

	for _, test := range tests {
		output := jsonstruct.NameFromInputFile(test.input)
		assert.Equal(t, output, test.output)
	}
}
