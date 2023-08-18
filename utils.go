package jsonstruct

import (
	"fmt"
	"path"
	"strings"
)

// GetGoName takes in a field name or file name and returns a "normalized" (CamelCase) string suitable for use as a Go
// variable name.
func GetGoName(input string) string {
	var cleaned strings.Builder

	// remove garbage characters, replace separators with ' '
	for _, r := range input {
		if isAlphaNum(r) {
			cleaned.WriteRune(r)
		} else if isSeparator(r) {
			cleaned.WriteRune(' ')
		}
	}

	// look for initialisms to capitalize
	words := strings.Split(cleaned.String(), " ")
	temp := []string{}

	for _, word := range words {
		tmpWord := strings.ToUpper(word)
		if _, ok := CommonInitialisms[tmpWord]; ok {
			word = strings.ToUpper(tmpWord)
		}

		temp = append(temp, word)
	}

	// TODO: replace this with golang.org/x/text/cases.Title()
	result := strings.Title(strings.Join(temp, " "))
	result = strings.ReplaceAll(result, " ", "")

	return result
}

func GetFileGoName(filePath string) string {
	_, fileName := path.Split(filePath)
	ext := path.Ext(fileName)

	return GetGoName(strings.TrimSuffix(fileName, ext))
}

func isAlphaNum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

func isSeparator(r rune) bool {
	return r == '_' || r == '.' || r == '-' || r == ' '
}

func anySliceToJSONStructs(input []any) (JSONStructs, error) {
	result := JSONStructs{}

	for i, item := range input {
		js, ok := item.(*JSONStruct)
		if !ok {
			return nil, fmt.Errorf("item %d was not a *JSONStruct", i)
		}

		result = append(result, js)
	}

	return result, nil
}

func simpleAnyToString(input any) string {
	switch val := input.(type) {
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		return fmt.Sprintf("%f", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case string:
		return fmt.Sprintf("\"%s\"", val)
	}

	return ""
}
