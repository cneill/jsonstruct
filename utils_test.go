package jsonstruct_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cneill/jsonstruct"
)

func TestGetGoName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"separators1", "this_is-a.test_name", "ThisIsATestName"},
		{"separators2", "remote_-.URL", "RemoteURL"},
		{"weirdcase", "ThiSiSaNumber2", "ThiSiSaNumber2"},
		{"separators_with_initialism", "This_Is_An_ID", "ThisIsAnID"},
		{"underscore_start", "_underscored", "Underscored"},
		{"garbage_characters", "($@%)@$%)(@", "Unknown"},
		{"garbage_characters_with_content", "@)(#$)@(#$)@#($garbage@#)$@)#($@)#($", "Garbage"},
		{"garbage_separator", "@t@e@s@t", "Test"},
		{"spaces", "       spaces", "Spaces"},
		{"multiple_spaces", "here are some spaces", "HereAreSomeSpaces"},
		{"leading_number", "0", "JSON0"},
		{"dot_special_case", ".", "Dot"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, jsonstruct.GetGoName(test.input))
		})
	}
}

func FuzzGetGoName(f *testing.F) {
	validRegex := regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`)

	seeds := []string{
		"this_is-a.test_name",
		"remote_-.URL",
		"ThiSiSaNumber2",
		"This_Is_An_ID",
		"_underscored",
		"($@%)@$%)(@",
		"@)(#$)@(#$)@#($garbage@#)$@)#($@)#($",
		"@t@e@s@t",
		"       spaces",
		"here are some spaces",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		transformed := jsonstruct.GetGoName(input)
		assert.True(t, validRegex.MatchString(transformed) || transformed == "")
	})
}

func TestGetFileGoName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "file_name.json", "FileName"},
		{"path", "/path/to/file.json", "File"},
		{"path_no_extension", "/path/to/file", "File"},
		// TODO: handle multiple extensions {"json.gz", "file_name.json.gz", "FileName"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, jsonstruct.GetFileGoName(test.input))
		})
	}
}
