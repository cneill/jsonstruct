package jsonstruct

import (
	"fmt"
	"os"
	"text/template"
)

var structTemplate = template.Must(template.New("struct").Parse(
	`type {{ .Name }} struct {
{{- range .Fields }}
{{ printf "\t%s\t%s\t%s" .Name .Type .Tag }}
{{- end }}
}
`))

// FormatterOptions defines how the Formatter will produce its output.
type FormatterOptions struct {
	// SortFields returns fields in alphabetically sorted order.
	SortFields bool

	// ValueComments annotates the produced structs with "Example" comments including the values originally passed in
	// for this field.
	ValueComments bool
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

func (f *Formatter) Format(input ...*JSONStruct) error {
	for _, js := range input {
		if f.SortFields {
			js.Fields.SortAlphabetically()
		}

		if err := structTemplate.Execute(os.Stdout, js); err != nil {
			return fmt.Errorf("failed to format struct with name %q: %w", js.Name, err)
		}
	}

	return nil
}
