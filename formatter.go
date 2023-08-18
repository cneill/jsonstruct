package jsonstruct

import (
	"fmt"
	"strings"
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

func (f *Formatter) FormatString(input ...*JSONStruct) (string, error) {
	structStrings := []string{}

	// TODO: handle arrays containing structs of the same kind differently
	// TODO: handle inline structs

	for _, js := range input {
		if f.SortFields {
			js.Fields.SortAlphabetically()
		}

		fieldStrings := strings.Join(f.fieldStrings(js.Fields...), "\n\t")
		formatted := fmt.Sprintf("type %s struct {\n\t%s\n}", js.Name, fieldStrings)
		structStrings = append(structStrings, formatted)

		// we've already printed out all the relevant structs inline
		if f.InlineStructs {
			continue
		}

		// if we're not inlining structs, find all the fields of type struct / []struct and print their type definitions
		// out too
		for _, field := range js.Fields {
			if field.IsStruct() || field.IsStructSlice() {
				formatted, err := f.FormatString(field.GetStruct())
				if err != nil {
					return "", fmt.Errorf("failed to format child struct %q: %w", field.Name(), err)
				}

				structStrings = append(structStrings, formatted)
			}
		}
	}

	return strings.Join(structStrings, "\n\n"), nil
}

func (f *Formatter) fieldStrings(fields ...*Field) []string {
	var (
		results = []string{}
		buckets = f.fieldBuckets(fields...)
	)

	// for each bucket, set field spacing based on longest name/type/tag of its neighbors
	for _, bucket := range buckets {
		var longestName, longestType, longestTag int

		for _, field := range bucket {
			if name := field.Name(); len(name) > longestName-1 {
				longestName = len(name) + 1
			}

			if typ := field.Type(); len(typ) > longestType-1 {
				longestType = len(typ) + 1
			}

			if tag := field.Tag(); len(tag) > longestTag-1 {
				longestTag = len(tag) + 1
			}
		}

		for _, field := range bucket {
			if f.ValueComments {
				fmtString := fmt.Sprintf("%%-%ds%%-%ds%%-%ds%%s", longestName, longestType, longestTag)
				results = append(results, fmt.Sprintf(fmtString, field.Name(), field.Type(), field.Tag(), field.Comment()))

				continue
			}

			fmtString := fmt.Sprintf("%%-%ds%%-%ds%%s", longestName, longestType)
			results = append(results, fmt.Sprintf(fmtString, field.Name(), field.Type(), field.Tag()))
		}
	}

	return results
}

func (f *Formatter) fieldBuckets(fields ...*Field) [][]*Field {
	// TODO: handle the case of comments on previous lines when that's a possibility
	if !f.InlineStructs {
		return [][]*Field{fields}
	}

	buckets := [][]*Field{}
	bucket := []*Field{}

	for _, field := range fields {
		if field.IsStruct() {
			bucket = append(bucket, field)
			buckets = append(buckets, bucket)
			bucket = []*Field{}
		} else {
			bucket = append(bucket, field)
		}
	}

	return buckets
}
