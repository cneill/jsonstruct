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
			expected: "\ntype Simple struct {\n\tA int64 `json:\"a\"`\n}\n",
		},
	}

	formatter, err := jsonstruct.NewFormatter(&jsonstruct.FormatterOptions{})
	assert.Nil(t, err)

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.input.SetName(jsonstruct.GetGoName(test.name))
			output, err := formatter.FormatStructs(test.input)
			assert.Nil(t, err)
			assert.Equal(t, test.expected, output)
		})
	}
}

func TestFormatStringFiles(t *testing.T) {
	t.Parallel()

	testFilePaths, err := filepath.Glob("test/*.json")
	assert.Nil(t, err)

	for _, testFilePath := range testFilePaths {
		testFilePath := testFilePath
		t.Run(testFilePath, func(t *testing.T) {
			t.Parallel()
			testFileDir, testFileName := path.Split(testFilePath)
			expectedFileName := fmt.Sprintf("%s_result.txt", strings.TrimSuffix(testFileName, path.Ext(testFileName)))
			expectedFilePath := path.Join(testFileDir, expectedFileName)

			testFile, err := os.Open(testFilePath)
			assert.Nil(t, err)

			defer testFile.Close()

			expectedContents, err := os.ReadFile(expectedFilePath)
			assert.Nil(t, err)

			parser := jsonstruct.NewParser(testFile, slog.Default())
			jStruct, err := parser.Start()
			assert.Nil(t, err)

			formatterOpts := &jsonstruct.FormatterOptions{}
			if strings.Contains(testFileName, "comment") {
				formatterOpts.ValueComments = true
			}

			if strings.Contains(testFileName, "inline") {
				formatterOpts.InlineStructs = true
			}

			formatter, err := jsonstruct.NewFormatter(formatterOpts)
			assert.Nil(t, err)

			output, err := formatter.FormatStructs(jStruct...)
			assert.Nil(t, err)

			assert.Equal(t, strings.TrimSpace(string(expectedContents)), strings.TrimSpace(output))
		})
	}
}
