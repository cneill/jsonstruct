package jsonstruct_test

import (
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/cneill/jsonstruct"

	"github.com/stretchr/testify/assert"
)

var (
	errNoGofmt = fmt.Errorf("failed to find 'gofmt' binary")

	structPrefix = "package temp\n\ntype test struct {\n"
	structSuffix = "}"
)

func goFmt(fields jsonstruct.Fields) (string, error) {
	goFmt, err := exec.LookPath("gofmt")
	if err != nil {
		return "", errNoGofmt
	}

	contents := structPrefix
	contents += fields.String()
	contents += structSuffix

	cmd := &exec.Cmd{
		Path:  goFmt,
		Stdin: strings.NewReader(contents),
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run gofmt: %w", err)
	}

	outputStr := strings.TrimSpace(string(output))

	return outputStr, nil
}

func fullStruct(fields jsonstruct.Fields) string {
	return fmt.Sprintf("%s%s%s", structPrefix, fields.String(), structSuffix)
}

func TestFieldsStringMatchesGoFmt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		inputs jsonstruct.Fields
	}{
		{
			name: "vanilla_fields",
			inputs: jsonstruct.Fields{
				jsonstruct.Field{RawName: "test", Name: "Test", Kind: reflect.String},
				jsonstruct.Field{RawName: "test2", Name: "Test2", Kind: reflect.Int64},
				jsonstruct.Field{RawName: "long_test_name", Name: "LongTestName", Kind: reflect.String},
				jsonstruct.Field{RawName: "test_float", Name: "TestFloat", Kind: reflect.Float64, RawValue: 1.0},
			},
		},
		{
			name: "commented_fields",
			inputs: jsonstruct.Fields{
				jsonstruct.Field{RawName: "test", Name: "Test", Kind: reflect.String, Comment: "test"},
				jsonstruct.Field{RawName: "test2", Name: "Test2", Kind: reflect.Int64, Comment: "test2"},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			output := fullStruct(test.inputs)
			goFmted, err := goFmt(test.inputs)
			if err != nil {
				t.Errorf("error on test %q: %v", test.name, err)
			}

			assert.Equal(t, goFmted, output)
		})
	}
}
