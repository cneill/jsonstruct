package jsonstruct

import (
	"fmt"
	"strings"

	"mvdan.cc/gofumpt/format"
)

// FormatterOptions defines how the Formatter will produce its output.
type FormatterOptions struct {
	// SortFields returns fields in alphabetically sorted order.
	SortFields bool

	// ValueComments annotates the produced structs with "Example" comments including the values originally passed in
	// for this field.
	ValueComments bool

	// InlineStructs causes objects within the main object to be rendered inline rather than getting their own types.
	InlineStructs bool
}

// OK ensures that the options passed in are valid.
func (f *FormatterOptions) OK() error {
	return nil
}

// Formatter prints out the contents of JSONStructs based on its configuration.
type Formatter struct {
	*FormatterOptions
}

// NewFormatter returns an initialized Formatter.
func NewFormatter(opts *FormatterOptions) (*Formatter, error) {
	if err := opts.OK(); err != nil {
		return nil, fmt.Errorf("invalid formatter options: %w", err)
	}

	f := &Formatter{
		FormatterOptions: opts,
	}

	return f, nil
}

func (f *Formatter) FormatStructs(inputs ...*JSONStruct) (string, error) {
	// this is required by gofumpt, it's removed at the end
	preamble := "package temp\n"
	structStr := preamble

	for inputNum, input := range inputs {
		if f.SortFields {
			input.fields.SortAlphabetically()
		}

		formatted, err := f.formatStructNesting(0, input)
		if err != nil {
			return "", fmt.Errorf("failed to format struct %d: %w", inputNum, err)
		}

		structStr += formatted

		// we already inlined all the struct fields, so no need to print out their type declarations at the end
		if f.InlineStructs {
			continue
		}

		for _, field := range input.Fields() {
			if field.IsStruct() || field.IsStructSlice() {
				formatted, err := f.FormatStructs(field.GetStruct())
				if err != nil {
					return "", fmt.Errorf("failed to format nested struct: %w", err)
				}

				structStr += formatted
			}
		}
	}

	formatted, err := format.Source([]byte(structStr), format.Options{})
	if err != nil {
		fmt.Printf("GOFUMPT INPUT:\n%s\n", structStr)
		return "", fmt.Errorf("failed to run gofumpt on generated structs: %w", err)
	}

	structStr = strings.ReplaceAll(string(formatted), preamble, "")

	return structStr, nil
}

// formatStructNetsting exists to allow us to track nesting without asking for it in FormatStructs, and so we can
// gofumpt only on the entire result, not all its pieces
func (f *Formatter) formatStructNesting(nest int, input *JSONStruct) (string, error) {
	structStr := ""

	// here, we don't want a "type" and we don't know if this is a struct or []struct, so just leave that to formatField
	// to prepend
	if f.InlineStructs && nest > 0 {
		structStr += " {\n"
	} else {
		structStr += fmt.Sprintf("type %s struct {\n", input.Name())
	}

	for _, field := range input.Fields() {
		fieldStr, err := f.formatField(nest, field)
		if err != nil {
			return "", err
		}

		structStr += fieldStr
	}

	structStr += "}"

	// we're printing out the "main" / non-inlined struct(s), give 'em some white space
	if !f.InlineStructs || nest == 0 {
		structStr += "\n\n"
	}

	return structStr, nil
}

// formatField handles formatting inline structs, []structs, and regular fields.
func (f *Formatter) formatField(nest int, field *Field) (string, error) {
	fieldStr := ""

	if f.InlineStructs && (field.IsStruct() || field.IsStructSlice()) {
		inlineStruct, err := f.formatStructNesting(nest+1, field.GetStruct())
		if err != nil {
			return "", fmt.Errorf("failed to get nested struct: %w", err)
		}

		sType := "struct"
		if field.IsStructSlice() {
			sType = "[]struct"
		}

		fieldStr += fmt.Sprintf("%s %s %s %s", field.Name(), sType, inlineStruct, field.Tag())
	} else {
		fieldStr += fmt.Sprintf("%s %s %s", field.Name(), field.Type(), field.Tag())
	}

	if f.ValueComments {
		fieldStr += fmt.Sprintf(" %s", field.Comment())
	}

	fieldStr += "\n"

	return fieldStr, nil
}
