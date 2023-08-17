package jsonstruct

import (
	"strings"
)

// GetNormalizedName takes in a field name or file name and returns a "normalized" (CamelCase) string suitable for use as a Go
// variable name.
func GetGoName(input string) string {
	var cleaned strings.Builder

	// remove garbage characters, replace separators with ' '
	for _, r := range input {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			cleaned.WriteRune(r)
		} else if r == '_' || r == '.' || r == '-' || r == ' ' {
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

	result := strings.Title(strings.Join(temp, " "))
	result = strings.ReplaceAll(result, " ", "")

	return result
}
