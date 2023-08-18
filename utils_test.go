package jsonstruct_test

import (
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

			output := jsonstruct.GetGoName(test.input)
			assert.Equal(t, output, test.output)
		})
	}
}
