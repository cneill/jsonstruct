package jsonstruct

import (
	"fmt"
	"math/big"
	"path"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// GetGoName takes in a field name or file name and returns a "normalized" (CamelCase) string suitable for use as a Go
// variable name.
func GetGoName(input string) string {
	var cleaned strings.Builder

	if input == "." {
		return "Dot" // special case
	}

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

	// transform to something like camel-case, leaving capitalizations already present intact
	caser := cases.Title(language.AmericanEnglish, cases.NoLower)
	result := caser.String(strings.Join(temp, " "))
	result = strings.ReplaceAll(result, " ", "")

	// variable names can't start with a number
	if len(result) > 0 && isNumber(rune(result[0])) {
		result = "JSON" + result
	}

	// unnamed struct fields / struct fields of only spaces break things
	if len(strings.TrimSpace(result)) == 0 {
		return "Unknown"
	}

	return result
}

func GetFileGoName(filePath string) string {
	_, fileName := path.Split(filePath)
	ext := path.Ext(fileName)

	return GetGoName(strings.TrimSuffix(fileName, ext))
}

func isAlphaNum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || isNumber(r)
}

func isNumber(r rune) bool {
	return (r >= '0' && r <= '9')
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
	case float64, *big.Float:
		return fmt.Sprintf("%.3f", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case string:
		return fmt.Sprintf("\"%s\"", val)
	}

	return ""
}
