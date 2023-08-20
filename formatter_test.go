package jsonstruct_test

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
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
			expected: "type Simple struct {\n        A int64 `json:\"a\"`\n}",
		},
	}

	formatter, err := jsonstruct.NewFormatter(&jsonstruct.FormatterOptions{})
	assert.Nil(t, err)

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.input.SetName(jsonstruct.GetGoName(test.name))
			output := formatter.FormatStruct(test.input)
			assert.Equal(t, test.expected, output)
		})
	}
}

func TestFormatStringFiles(t *testing.T) {
	t.Parallel()

	testFilePaths, err := filepath.Glob("test/*.json")
	assert.Nil(t, err)

	for _, testFilePath := range testFilePaths {
		testFileDir, testFileName := path.Split(testFilePath)
		expectedFileName := fmt.Sprintf("%s_result.txt", strings.TrimSuffix(testFileName, path.Ext(testFileName)))
		expectedFilePath := path.Join(testFileDir, expectedFileName)

		testFile, err := os.Open(testFilePath)
		assert.Nil(t, err)

		defer testFile.Close()

		expectedContents, err := os.ReadFile(expectedFilePath)
		assert.Nil(t, err)

		parser := jsonstruct.NewParser(testFile, slog.Default())
		js, err := parser.Start()
		assert.Nil(t, err)

		formatterOpts := &jsonstruct.FormatterOptions{}
		if strings.Contains(testFileName, "comment") {
			formatterOpts.ValueComments = true
		}

		formatter, err := jsonstruct.NewFormatter(formatterOpts)
		assert.Nil(t, err)

		output := formatter.FormatStruct(js...)

		assert.Equal(t, strings.TrimSpace(string(expectedContents)), output)
	}
}
