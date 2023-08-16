package main

import (
	"fmt"
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

func stringValue(input any) string {
	switch val := input.(type) {
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		return fmt.Sprintf("%f", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case string:
		return val
	default:
		if val == nil {
			return "null"
		}

		// TODO: simple arrays?

		return ""
	}
}
