package gari

import (
	"regexp"
	"strings"
)

var (
	reUpper      = regexp.MustCompile(`[A-Z]`)
	reUnderbar   = regexp.MustCompile(`_(.)`)
	reMultiSpace = regexp.MustCompile(`\s{2,}`)
)

// Convert snake-case string to lower camel-case.
// * snake_case -> snakeCase
func snakeToLowerCamelCase(str string) string {
	return strings.ReplaceAll(
		reUnderbar.ReplaceAllStringFunc(str, strings.ToUpper),
		"_", "")
}

// Convert snake-case string to upper camel-case.
// * snake_case -> SnakeCase
func snakeToUpperCamelCase(str string) string {
	s := snakeToLowerCamelCase(str)
	return strings.ToUpper(s[0:1]) + s[1:]
}

// Convert upper camel-case string to snake-case.
// * UpperCamelCase -> upper_camel_case
func upperCamelToSnakeCase(input string) string {
	str := reUpper.ReplaceAllStringFunc(input,
		func(matched string) string {
			return "_" + strings.ToLower(matched)
		})
	if str[0:1] == "_" {
		return str[1:]
	}
	return str
}

func squashSpace(input string) string {
	return reMultiSpace.ReplaceAllString(input, " ")
}
