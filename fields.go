package jsonstruct

import "fmt"

// Field represents a single struct field.
type Field struct {
	GoName       string
	OriginalName string
	StrValue     string
	RawValue     any
}

// Name returns the name of this field as it will be rendered in the final struct.
func (f *Field) Name() string {
	return f.GoName
}

// Tag returns the JSON tag as it will be rendered in the final struct.
func (f *Field) Tag() string {
	return fmt.Sprintf("`json: \"%s\"`", f.OriginalName)
}

// Type returns the type of the field as it will be rendered in the final struct.
func (f *Field) Type() string {
	switch f.RawValue.(type) {
	case int64:
		return "int64"
	case float64:
		return "float64"
	case string:
		return "string"
	case bool:
		return "bool"
	}

	if f.RawValue == nil {
		return "*json.RawMessage"
	}

	return "DUNNO BOSS"
}

// Fields is a convenience type for a slice of Field structs.
type Fields []*Field
