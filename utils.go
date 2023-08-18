package jsonstruct

import (
	"fmt"
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

func isAlphaNum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

func isSeparator(r rune) bool {
	return r == '_' || r == '.' || r == '-' || r == ' '
}

func anySliceToJSONStructs(input []any) (JSONStructs, error) {
	result := JSONStructs{}

	for i, item := range input {
		js, ok := item.(JSONStruct)
		if !ok {
			return nil, fmt.Errorf("item %d was not a JSONStruct", i)
		}

		result = append(result, js)
	}

	return result, nil
}
